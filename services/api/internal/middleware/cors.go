package middleware

import (
	"net/http"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/gin-gonic/gin"
)

// normalizeOrigin strips a trailing slash so "https://app.vercel.app/" matches
// the browser Origin header "https://app.vercel.app".
func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}

// CORS applies Access-Control headers for configured origins.
// Empty origins list allows no cross-origin browsers (same-origin only).
func CORS(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	allowAll := false
	for _, o := range origins {
		o = normalizeOrigin(o)
		if o == constant.CORSAllowOriginAll {
			allowAll = true
			break
		}
		if o != "" {
			allowed[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader(constant.HeaderOrigin))
		if origin != "" {
			if allowAll {
				// Credentials cannot be used with wildcard origins.
				c.Header(constant.HeaderAccessControlAllowOrigin, constant.CORSAllowOriginAll)
			} else if _, ok := allowed[origin]; ok {
				c.Header(constant.HeaderAccessControlAllowOrigin, origin)
				c.Header(constant.HeaderAccessControlAllowCredentials, "true")
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
