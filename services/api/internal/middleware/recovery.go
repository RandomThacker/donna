package middleware

import (
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// Recovery catches panics, logs them via the module logger, and returns a 500 envelope.
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error(c.Request.Context(), "panic recovered",
					constant.LogAttrError, rec,
					constant.LogAttrPath, c.Request.URL.Path,
					constant.LogAttrMethod, c.Request.Method,
				)
				response.Error(
					c,
					http.StatusInternalServerError,
					constant.MessageInternalError,
					constant.ErrorCodeInternal,
					constant.MessageUnexpectedPanic,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
