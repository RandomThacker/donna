package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AutomationHandler maps automation HTTP endpoints to Actions.
type AutomationHandler struct {
	list         *actions.ListAutomationsAction
	templates    *actions.ListAutomationTemplatesAction
	create       *actions.CreateAutomationAction
	update       *actions.UpdateAutomationAction
	delete       *actions.DeleteAutomationAction
	history      *actions.ListAutomationHistoryAction
	historyAll   *actions.ListAllAutomationHistoryAction
	execution    *actions.GetAutomationExecutionAction
	analytics    *actions.GetAutomationAnalyticsAction
	run          *actions.RunAutomationAction
	preview      *actions.PreviewAutomationAction
	environment  string
	log          *logger.Logger
}

// NewAutomationHandler constructs an AutomationHandler.
func NewAutomationHandler(
	list *actions.ListAutomationsAction,
	templates *actions.ListAutomationTemplatesAction,
	create *actions.CreateAutomationAction,
	update *actions.UpdateAutomationAction,
	delete *actions.DeleteAutomationAction,
	history *actions.ListAutomationHistoryAction,
	historyAll *actions.ListAllAutomationHistoryAction,
	execution *actions.GetAutomationExecutionAction,
	analytics *actions.GetAutomationAnalyticsAction,
	run *actions.RunAutomationAction,
	preview *actions.PreviewAutomationAction,
	environment string,
	log *logger.Logger,
) *AutomationHandler {
	return &AutomationHandler{
		list: list, templates: templates, create: create, update: update, delete: delete,
		history: history, historyAll: historyAll, execution: execution, analytics: analytics,
		run: run, preview: preview,
		environment: environment, log: log,
	}
}

func (h *AutomationHandler) developerMode() bool {
	return h.environment == constant.EnvDevelopment || h.environment == ""
}

// List handles GET /automations.
func (h *AutomationHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	autos, err := h.list.Execute(c.Request.Context(), actions.ListAutomationsRequest{UserID: userID})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"automations": automationsFromActions(autos)})
}

// ListTemplates handles GET /automations/templates.
func (h *AutomationHandler) ListTemplates(c *gin.Context) {
	if _, ok := middleware.UserIDFromContext(c); !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	templates, err := h.templates.Execute()
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"templates": model.AutomationTemplatesFromCatalog(templates)})
}

// ListAllHistory handles GET /automations/history.
func (h *AutomationHandler) ListAllHistory(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	limit := historyLimit(c)
	rows, err := h.historyAll.Execute(c.Request.Context(), userID, limit)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{
		"executions": automationExecutionsFromActions(rows, false),
	})
}

// Analytics handles GET /automations/analytics.
func (h *AutomationHandler) Analytics(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	stats, err := h.analytics.Execute(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, automationAnalyticsFromBusiness(stats))
}

// History handles GET /automations/:id/history.
func (h *AutomationHandler) History(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	autoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid automation id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	limit := historyLimit(c)
	rows, err := h.history.Execute(c.Request.Context(), userID, autoID, limit)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{
		"executions": automationExecutionsFromActions(rows, false),
	})
}

// GetExecution handles GET /automations/executions/:id.
func (h *AutomationHandler) GetExecution(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	execID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid execution id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	exec, err := h.execution.Execute(c.Request.Context(), userID, execID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, automationExecutionFromAction(exec, h.developerMode()))
}

// Create handles POST /automations.
func (h *AutomationHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	triggerType, triggerTime := "", ""
	if req.Trigger != nil {
		triggerType = req.Trigger.Type
		triggerTime = req.Trigger.Time
	}
	var delivery []string
	if req.Delivery != nil {
		delivery = req.Delivery.Channels
	}
	auto, err := h.create.Execute(c.Request.Context(), actions.CreateAutomationRequest{
		UserID:           userID,
		Name:             req.Name,
		Description:      req.Description,
		Enabled:          req.Enabled,
		TriggerType:      triggerType,
		TriggerTime:      triggerTime,
		Timezone:         req.Timezone,
		Commands:         model.AutomationCommandsToEntities(req.Commands),
		DeliveryChannels: delivery,
		TemplateID:       req.TemplateID,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, automationFromAction(auto))
}

// Update handles PATCH /automations/:id.
func (h *AutomationHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	autoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid automation id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var triggerType, triggerTime *string
	if req.Trigger != nil {
		tt := req.Trigger.Type
		tm := req.Trigger.Time
		if tt != "" {
			triggerType = &tt
		}
		if tm != "" {
			triggerTime = &tm
		}
	}
	var delivery []string
	if req.Delivery != nil {
		delivery = req.Delivery.Channels
	}
	var commands []entity.AutomationCommand
	if req.Commands != nil {
		commands = model.AutomationCommandsToEntities(req.Commands)
	}
	auto, err := h.update.Execute(c.Request.Context(), actions.UpdateAutomationRequest{
		UserID:           userID,
		AutomationID:     autoID,
		Name:             req.Name,
		Description:      req.Description,
		Enabled:          req.Enabled,
		TriggerType:      triggerType,
		TriggerTime:      triggerTime,
		Timezone:         req.Timezone,
		Commands:         commands,
		DeliveryChannels: delivery,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, automationFromAction(auto))
}

// Run handles POST /automations/:id/run.
func (h *AutomationHandler) Run(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	if h.run == nil {
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "run not configured")
		return
	}
	autoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid automation id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	result, err := h.run.Execute(c.Request.Context(), actions.RunAutomationRequest{
		UserID:       userID,
		AutomationID: autoID,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, automationRunFromBusiness(result, false))
}

// Preview handles POST /automations/:id/preview.
func (h *AutomationHandler) Preview(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	if h.preview == nil {
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "preview not configured")
		return
	}
	autoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid automation id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	result, err := h.preview.Execute(c.Request.Context(), actions.PreviewAutomationRequest{
		UserID:       userID,
		AutomationID: autoID,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, automationRunFromBusiness(result, true))
}

// Delete handles DELETE /automations/:id.
func (h *AutomationHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	autoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid automation id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.delete.Execute(c.Request.Context(), actions.DeleteAutomationRequest{
		UserID:       userID,
		AutomationID: autoID,
	}); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

func historyLimit(c *gin.Context) int {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}
	return limit
}

func (h *AutomationHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "automation request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "automation failed")
	}
}
