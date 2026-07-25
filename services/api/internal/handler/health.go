package handler

import (
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// HealthHandler maps health endpoints to the business layer.
type HealthHandler struct {
	svc *business.HealthService
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(svc *business.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// Health is liveness: process is up (no DB dependency).
func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, constant.MessageOK, model.HealthFromEntity(h.svc.Liveness()))
}

// Ready is readiness: config loaded and database reachable.
func (h *HealthHandler) Ready(c *gin.Context) {
	status, err := h.svc.Readiness(c.Request.Context())
	if err != nil {
		response.Error(
			c,
			http.StatusServiceUnavailable,
			constant.MessageServiceNotReady,
			constant.ErrorCodeDBUnavailable,
			err.Error(),
		)
		return
	}
	response.OK(c, constant.MessageOK, model.ReadyFromEntity(status))
}

// Version returns build version metadata.
func (h *HealthHandler) Version(c *gin.Context) {
	response.OK(c, constant.MessageOK, model.VersionFromEntity(h.svc.Version()))
}
