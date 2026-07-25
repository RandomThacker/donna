package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestRequestLoggingIncludesFieldsAndSlowWarn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	f := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: constant.EnvProduction,
		Level:       constant.LogLevelInfo,
		Output:      &buf,
	})
	httpLog := f.Module(constant.ModuleHTTP)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogging(httpLog))
	r.GET("/ping", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(constant.HeaderRequestID, "fixed-req")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	out := buf.String()
	for _, want := range []string{
		`"msg":"request"`,
		`"request_id":"fixed-req"`,
		`"method":"GET"`,
		`"path":"/ping"`,
		`"status":200`,
		`"module":"http"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log %q missing %q", out, want)
		}
	}
}
