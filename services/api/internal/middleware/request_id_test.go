package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestRequestIDGeneratesAndPropagates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, middleware.GetRequestID(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	id := w.Header().Get(constant.HeaderRequestID)
	if id == "" {
		t.Fatal("expected X-Request-ID header")
	}
	if w.Body.String() != id {
		t.Fatalf("body %q != header %q", w.Body.String(), id)
	}
}

func TestRequestIDUsesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(constant.HeaderRequestID, "fixed-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get(constant.HeaderRequestID); got != "fixed-id-123" {
		t.Fatalf("header = %q", got)
	}
}
