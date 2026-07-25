package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

func TestJSONEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.OK(c, constant.MessageOK, map[string]string{"status": constant.StatusOK})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("success = %v", body["success"])
	}
	if body["message"] != constant.MessageOK {
		t.Fatalf("message = %v", body["message"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["status"] != constant.StatusOK {
		t.Fatalf("data = %#v", body["data"])
	}
}

func TestErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(
		c,
		http.StatusServiceUnavailable,
		constant.MessageServiceNotReady,
		constant.ErrorCodeDBUnavailable,
		"ping failed",
	)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("success = %v", body["success"])
	}
	errBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("error missing: %#v", body)
	}
	if errBody["code"] != constant.ErrorCodeDBUnavailable {
		t.Fatalf("code = %v", errBody["code"])
	}
}
