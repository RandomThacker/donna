package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/model"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DonnaReminderHandler maps Donna reminder HTTP endpoints to Actions.
type DonnaReminderHandler struct {
	list   *business.DonnaReminderService
	create *actions.CreateReminderAction
	update *actions.UpdateReminderAction
	delete *actions.DeleteReminderAction
	log    *logger.Logger
}

// NewDonnaReminderHandler constructs a DonnaReminderHandler.
func NewDonnaReminderHandler(
	list *business.DonnaReminderService,
	create *actions.CreateReminderAction,
	update *actions.UpdateReminderAction,
	delete *actions.DeleteReminderAction,
	log *logger.Logger,
) *DonnaReminderHandler {
	return &DonnaReminderHandler{list: list, create: create, update: update, delete: delete, log: log}
}

// List handles GET /donna/reminders.
func (h *DonnaReminderHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	reminders, err := h.list.List(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"reminders": model.DonnaRemindersFromEntities(reminders)})
}

// Create handles POST /donna/reminders.
func (h *DonnaReminderHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateDonnaReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	triggerAt, err := time.Parse(time.RFC3339, req.TriggerAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "trigger_at must be RFC3339")
		return
	}
	reminder, err := h.create.Execute(c.Request.Context(), actions.CreateReminderRequest{
		UserID:         userID,
		Title:          req.Title,
		Description:    req.Description,
		TriggerAt:      triggerAt,
		Timezone:       req.Timezone,
		RecurrenceRule: req.RecurrenceRule,
		Color:          req.Color,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, donnaReminderFromAction(reminder))
}

// Update handles PATCH /donna/reminders/:id.
func (h *DonnaReminderHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid reminder id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateDonnaReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	triggerAt, err := parseOptionalRFC3339(req.TriggerAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "trigger_at must be RFC3339")
		return
	}
	reminder, err := h.update.Execute(c.Request.Context(), actions.UpdateReminderRequest{
		UserID:         userID,
		ReminderID:     reminderID,
		Title:          req.Title,
		Description:    req.Description,
		TriggerAt:      triggerAt,
		Timezone:       req.Timezone,
		RecurrenceRule: req.RecurrenceRule,
		Status:         req.Status,
		Color:          req.Color,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, donnaReminderFromAction(reminder))
}

// Delete handles DELETE /donna/reminders/:id.
func (h *DonnaReminderHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid reminder id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.delete.Execute(c.Request.Context(), actions.DeleteReminderRequest{
		UserID:     userID,
		ReminderID: reminderID,
	}); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

func (h *DonnaReminderHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "donna reminder request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "donna reminder failed")
	}
}
