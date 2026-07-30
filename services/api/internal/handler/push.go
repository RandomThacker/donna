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
)

// PushHandler maps Web Push subscription endpoints to the business layer.
type PushHandler struct {
	svc            *business.PushSubscriptionService
	vapidPublicKey string
	log            *logger.Logger
}

// NewPushHandler constructs a PushHandler.
func NewPushHandler(svc *business.PushSubscriptionService, vapidPublicKey string, log *logger.Logger) *PushHandler {
	return &PushHandler{svc: svc, vapidPublicKey: vapidPublicKey, log: log}
}

// Subscribe handles POST /push/subscribe.
func (h *PushHandler) Subscribe(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.SubscribePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	ua := req.UserAgent
	if ua == nil {
		if header := strings.TrimSpace(c.Request.UserAgent()); header != "" {
			ua = &header
		}
	}
	sub, err := h.svc.Subscribe(c.Request.Context(), userID, business.SubscribePushInput{
		Endpoint:  req.Endpoint,
		P256dh:    req.Keys.P256dh,
		Auth:      req.Keys.Auth,
		UserAgent: ua,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, model.PushSubscriptionFromEntity(sub))
}

// Unsubscribe handles DELETE /push/unsubscribe.
func (h *PushHandler) Unsubscribe(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.UnsubscribePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.svc.Unsubscribe(c.Request.Context(), userID, req.Endpoint); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"unsubscribed": true})
}

// VAPIDPublicKey handles GET /push/vapid-public-key.
func (h *PushHandler) VAPIDPublicKey(c *gin.Context) {
	if strings.TrimSpace(h.vapidPublicKey) == "" {
		response.Error(c, http.StatusServiceUnavailable, "web push not configured", constant.ErrorCodeNotConfigured, "missing VAPID public key")
		return
	}
	response.OK(c, constant.MessageOK, model.VAPIDPublicKeyResponse{PublicKey: h.vapidPublicKey})
}

func (h *PushHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "push request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "push failed")
	}
}
