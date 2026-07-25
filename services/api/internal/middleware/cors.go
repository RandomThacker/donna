package middleware

import (
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/gin-gonic/gin"
)

// CORS applies Access-Control headers for configured origins.
// Empty origins list allows no cross-origin browsers (same-origin only).
func CORS(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	allowAll := false
	for _, o := range origins {
		if o == constant.CORSAllowOriginAll {
			allowAll = true
			break
		}
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader(constant.HeaderOrigin)
		if origin != "" {
			if allowAll {
				c.Header(constant.HeaderAccessControlAllowOrigin, constant.CORSAllowOriginAll)
			} else if _, ok := allowed[origin]; ok {
				c.Header(constant.HeaderAccessControlAllowOrigin, origin)
				c.Header(constant.HeaderVary, constant.HeaderOrigin)
			}
		}

		c.Header(constant.HeaderAccessControlAllowMethods, constant.CORSAllowMethods)
		c.Header(constant.HeaderAccessControlAllowHeaders, constant.CORSAllowHeaders)
		c.Header(constant.HeaderAccessControlExposeHeaders, constant.HeaderRequestID)
		c.Header(constant.HeaderAccessControlMaxAge, constant.CORSMaxAgeSeconds)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
