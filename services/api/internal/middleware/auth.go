package middleware

import (
	"net/http"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireAuth validates a Bearer token or donna_session cookie and stores user_id.
func RequireAuth(issuer *session.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader(constant.HeaderAuthorization))
		if token == "" {
			if cookie, err := c.Cookie(constant.CookieSession); err == nil {
				token = cookie
			}
		}
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing session")
			c.Abort()
			return
		}

		claims, err := issuer.Parse(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "invalid session")
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil || userID == uuid.Nil {
			response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "invalid session subject")
			c.Abort()
			return
		}

		c.Set(constant.ContextKeyUserID, userID)
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user id set by RequireAuth.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(constant.ContextKeyUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
