package middleware

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/gin-gonic/gin"
)

// RequestLogging logs every HTTP request with correlation fields.
// Requests slower than SlowRequestThreshold are logged at WARN.
func RequestLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		ctx := c.Request.Context()
		status := c.Writer.Status()

		args := []any{
			constant.LogAttrMethod, c.Request.Method,
			constant.LogAttrPath, c.Request.URL.Path,
			constant.LogAttrStatus, status,
			constant.LogAttrDurationMS, duration.Milliseconds(),
			constant.LogAttrClientIP, c.ClientIP(),
			constant.LogAttrUserAgent, c.Request.UserAgent(),
		}

		switch {
		case duration >= constant.SlowRequestThreshold:
			log.Warn(ctx, "slow request", args...)
		default:
			log.Info(ctx, "request", args...)
		}
	}
}
