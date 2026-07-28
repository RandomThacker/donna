package handler

import (
	"errors"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/httpx"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// MeHandler serves the authenticated current-user endpoint.
type MeHandler struct {
	users        *business.UserService
	cookieSecure bool
}

// NewMeHandler constructs a MeHandler.
func NewMeHandler(users *business.UserService, cookieSecure bool) *MeHandler {
	return &MeHandler{users: users, cookieSecure: cookieSecure}
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
			http.SetCookie(c.Writer, httpx.SessionCookie(
				constant.CookieSession,
				"",
				-1,
				h.cookieSecure,
			))
			response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
		return
	}

	response.OK(c, constant.MessageOK, model.UserFromEntity(user))
}
