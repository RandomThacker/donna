package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/gin-gonic/gin"
)

// RequestID ensures every request has an ID in Gin context, request context, and response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(constant.HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Set(constant.ContextKeyRequestID, id)
		c.Writer.Header().Set(constant.HeaderRequestID, id)

		ctx := logger.WithFields(c.Request.Context(), logger.Fields{RequestID: id})
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestID returns the request ID from Gin context, or empty string.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(constant.ContextKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return logger.FieldsFrom(c.Request.Context()).RequestID
}

// SetUserID attaches an authenticated user id to the request logging context.
// Call from auth middleware once M2 lands.
func SetUserID(c *gin.Context, userID string) {
	if userID == "" {
		return
	}
	c.Set(constant.LogAttrUserID, userID)
	ctx := logger.WithFields(c.Request.Context(), logger.Fields{UserID: userID})
	c.Request = c.Request.WithContext(ctx)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
