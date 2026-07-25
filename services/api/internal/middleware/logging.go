package middleware

import (
	"log/slog"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/gin-gonic/gin"
)

// RequestLogging logs method, path, status, duration, client IP, and request ID.
func RequestLogging(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Info("request",
			constant.LogAttrRequestID, GetRequestID(c),
			constant.LogAttrMethod, c.Request.Method,
			constant.LogAttrPath, c.Request.URL.Path,
			constant.LogAttrStatus, c.Writer.Status(),
			constant.LogAttrDurationMS, time.Since(start).Milliseconds(),
			constant.LogAttrClientIP, c.ClientIP(),
		)
	}
}
