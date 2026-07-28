package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IntegrationHandler maps integration OAuth endpoints (calendar connect, not login).
type IntegrationHandler struct {
	svc                *business.IntegrationService
	log                *logger.Logger
	frontendSuccessURL string
}

// NewIntegrationHandler constructs an IntegrationHandler.
func NewIntegrationHandler(svc *business.IntegrationService, log *logger.Logger, frontendSuccessURL string) *IntegrationHandler {
	return &IntegrationHandler{
		svc:                svc,
		log:                log,
		frontendSuccessURL: frontendSuccessURL,
	}
}

// ListConnectedAccounts returns the caller's live connected accounts.
func (h *IntegrationHandler) ListConnectedAccounts(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	accounts, err := h.svc.ListConnectedAccounts(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.ConnectedAccountsFromEntities(accounts))
}

// BeginGoogleConnect starts Google calendar OAuth (requires Donna session).
func (h *IntegrationHandler) BeginGoogleConnect(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	authURL, err := h.svc.BeginGoogleConnect(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback completes Google calendar OAuth.
func (h *IntegrationHandler) GoogleCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		response.Error(c, http.StatusBadRequest, "google oauth denied", constant.ErrorCodeOAuthFailed, errParam)
		return
	}
	_, err := h.svc.CompleteGoogleConnect(c.Request.Context(), c.Query("code"), c.Query("state"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.finishConnect(c, constant.AuthProviderGoogle)
}

// BeginMicrosoftConnect starts Microsoft calendar OAuth (requires Donna session).
func (h *IntegrationHandler) BeginMicrosoftConnect(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	authURL, err := h.svc.BeginMicrosoftConnect(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// MicrosoftCallback completes Microsoft calendar OAuth.
func (h *IntegrationHandler) MicrosoftCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		response.Error(c, http.StatusBadRequest, "microsoft oauth denied", constant.ErrorCodeOAuthFailed, errParam)
		return
	}
	_, err := h.svc.CompleteMicrosoftConnect(c.Request.Context(), c.Query("code"), c.Query("state"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	h.finishConnect(c, constant.AuthProviderMicrosoft)
}

// Disconnect soft-deletes a connected account owned by the caller.
func (h *IntegrationHandler) Disconnect(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	accountID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid account id", constant.ErrorCodeValidation, "id must be a uuid")
		return
	}
	if err := h.svc.Disconnect(c.Request.Context(), userID, accountID); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, nil)
}

// ListICS returns Calendar URL (ICS) integrations for the caller.
func (h *IntegrationHandler) ListICS(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	views, err := h.svc.ListICS(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, icsViewsToResponse(views))
}

// ConnectICS creates a Calendar URL (ICS) feed integration.
func (h *IntegrationHandler) ConnectICS(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.ConnectICSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeValidation, err.Error())
		return
	}
	view, err := h.svc.ConnectICS(c.Request.Context(), userID, business.ConnectICSRequest{
		Name:        req.Name,
		ICSURL:      req.ICSURL,
		SyncEnabled: req.SyncEnabled,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, model.ICSIntegrationFromEntity(view.Account, view.SyncEnabled, view.EventCount))
}

// UpdateICS patches a Calendar URL (ICS) feed integration.
func (h *IntegrationHandler) UpdateICS(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	accountID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid account id", constant.ErrorCodeValidation, "id must be a uuid")
		return
	}
	var req model.UpdateICSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeValidation, err.Error())
		return
	}
	view, err := h.svc.UpdateICS(c.Request.Context(), userID, accountID, business.UpdateICSRequest{
		Name:        req.Name,
		ICSURL:      req.ICSURL,
		SyncEnabled: req.SyncEnabled,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.ICSIntegrationFromEntity(view.Account, view.SyncEnabled, view.EventCount))
}

// DeleteICS removes a Calendar URL (ICS) feed integration.
func (h *IntegrationHandler) DeleteICS(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	accountID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid account id", constant.ErrorCodeValidation, "id must be a uuid")
		return
	}
	if err := h.svc.DeleteICS(c.Request.Context(), userID, accountID); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, nil)
}

// SyncICS triggers an immediate sync for one ICS feed.
func (h *IntegrationHandler) SyncICS(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	accountID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid account id", constant.ErrorCodeValidation, "id must be a uuid")
		return
	}
	view, err := h.svc.SyncICS(c.Request.Context(), userID, accountID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.ICSIntegrationFromEntity(view.Account, view.SyncEnabled, view.EventCount))
}

func icsViewsToResponse(views []business.ICSIntegrationView) []model.ICSIntegrationResponse {
	out := make([]model.ICSIntegrationResponse, 0, len(views))
	for _, view := range views {
		out = append(out, model.ICSIntegrationFromEntity(view.Account, view.SyncEnabled, view.EventCount))
	}
	return out
}

func (h *IntegrationHandler) finishConnect(c *gin.Context, provider string) {
	if accept := c.GetHeader("Accept"); accept == "application/json" || c.Query("format") == "json" {
		response.OK(c, constant.MessageOK, gin.H{"provider": provider, "connected": true})
		return
	}
	dest := strings.TrimSpace(h.frontendSuccessURL)
	if dest == "" {
		dest = "/"
	}
	c.Redirect(http.StatusFound, dest)
}

func (h *IntegrationHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	case errors.Is(err, apperr.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "invalid request", constant.ErrorCodeInvalidRequest, err.Error())
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		response.Error(c, http.StatusConflict, "conflict", constant.ErrorCodeConflict, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "integration request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
	}
}
