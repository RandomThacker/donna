package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// PersonalityHandler maps personality settings HTTP endpoints to Actions.
type PersonalityHandler struct {
	get      *actions.GetPersonalityAction
	update   *actions.UpdatePersonalityAction
	catalog  *actions.ListPersonalityCatalogAction
	preview  *actions.PreviewPersonalityAction
	log      *logger.Logger
}

// NewPersonalityHandler constructs a PersonalityHandler.
func NewPersonalityHandler(
	get *actions.GetPersonalityAction,
	update *actions.UpdatePersonalityAction,
	catalog *actions.ListPersonalityCatalogAction,
	preview *actions.PreviewPersonalityAction,
	log *logger.Logger,
) *PersonalityHandler {
	return &PersonalityHandler{get: get, update: update, catalog: catalog, preview: preview, log: log}
}

// Get handles GET /settings/personality.
func (h *PersonalityHandler) Get(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	profile, err := h.get.Execute(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.PersonalityProfileFromDomain(profile))
}

// Update handles PATCH /settings/personality.
func (h *PersonalityHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.UpdatePersonalityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	profile, err := h.update.Execute(c.Request.Context(), userID, business.PersonalityUpdateInput{
		PersonalityID:      req.PersonalityID,
		DisplayName:        req.DisplayName,
		Nickname:           req.Nickname,
		EmojiLevel:         req.EmojiLevel,
		HumorLevel:         req.HumorLevel,
		GreetingStyle:      req.GreetingStyle,
		EncouragementLevel: req.EncouragementLevel,
		ResponseStyle:      req.ResponseStyle,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.PersonalityProfileFromDomain(profile))
}

// Catalog handles GET /settings/personality/catalog.
func (h *PersonalityHandler) Catalog(c *gin.Context) {
	if _, ok := middleware.UserIDFromContext(c); !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	defs, err := h.catalog.Execute()
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"personalities": model.PersonalityDefinitionsFromDomain(defs)})
}

// Preview handles POST /settings/personality/preview.
func (h *PersonalityHandler) Preview(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.PersonalityPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var override *personality.Profile
	if req.PersonalityID != nil || req.DisplayName != nil || req.Nickname != nil ||
		req.EmojiLevel != nil || req.HumorLevel != nil || req.GreetingStyle != nil ||
		req.EncouragementLevel != nil || req.ResponseStyle != nil {
		p := personality.DefaultProfile(userID)
		if req.PersonalityID != nil {
			p.PersonalityID = strings.TrimSpace(*req.PersonalityID)
		}
		if req.DisplayName != nil {
			p.DisplayName = strings.TrimSpace(*req.DisplayName)
		}
		if req.Nickname != nil {
			p.Nickname = strings.TrimSpace(*req.Nickname)
		}
		if req.EmojiLevel != nil {
			p.EmojiLevel = personality.Level(strings.TrimSpace(*req.EmojiLevel))
		}
		if req.HumorLevel != nil {
			p.HumorLevel = personality.Level(strings.TrimSpace(*req.HumorLevel))
		}
		if req.GreetingStyle != nil {
			p.GreetingStyle = strings.TrimSpace(*req.GreetingStyle)
		}
		if req.EncouragementLevel != nil {
			p.EncouragementLevel = personality.Level(strings.TrimSpace(*req.EncouragementLevel))
		}
		if req.ResponseStyle != nil {
			p.ResponseStyle = strings.TrimSpace(*req.ResponseStyle)
		}
		override = &p
	}
	samples, err := h.preview.Execute(c.Request.Context(), actions.PreviewPersonalityRequest{
		UserID:   userID,
		Timezone: req.Timezone,
		Override: override,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.PersonalityPreviewFromMap(samples))
}

func (h *PersonalityHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "personality request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "personality failed")
	}
}
