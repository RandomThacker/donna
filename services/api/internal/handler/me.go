package handler

import (
	"errors"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// MeHandler serves the authenticated current-user endpoint.
type MeHandler struct {
	users *business.UserService
}

// NewMeHandler constructs a MeHandler.
func NewMeHandler(users *business.UserService) *MeHandler {
	return &MeHandler{users: users}
}

// Me handles GET /me.
func (h *MeHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			// Stale JWT after DB wipe / deleted user — drop the cookie so login can proceed.
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     constant.CookieSession,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
		return
	}

	response.OK(c, constant.MessageOK, model.UserFromEntity(user))
}
