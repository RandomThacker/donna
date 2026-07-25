package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/gin-gonic/gin"
)

// RequestID ensures every request has an ID in context and response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(constant.HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Set(constant.ContextKeyRequestID, id)
		c.Writer.Header().Set(constant.HeaderRequestID, id)
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
	return ""
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
