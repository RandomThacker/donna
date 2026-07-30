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

// DonnaEventHandler maps Donna event HTTP endpoints to Actions.
type DonnaEventHandler struct {
	list    *business.DonnaEventService
	create  *actions.CreateEventAction
	update  *actions.UpdateEventAction
	delete  *actions.DeleteEventAction
	log     *logger.Logger
}

// NewDonnaEventHandler constructs a DonnaEventHandler.
func NewDonnaEventHandler(
	list *business.DonnaEventService,
	create *actions.CreateEventAction,
	update *actions.UpdateEventAction,
	delete *actions.DeleteEventAction,
	log *logger.Logger,
) *DonnaEventHandler {
	return &DonnaEventHandler{list: list, create: create, update: update, delete: delete, log: log}
}

// List handles GET /donna/events.
func (h *DonnaEventHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	events, err := h.list.List(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"events": model.DonnaEventsFromEntities(events)})
}

// Create handles POST /donna/events.
func (h *DonnaEventHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	var req model.CreateDonnaEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "start_at must be RFC3339")
		return
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "end_at must be RFC3339")
		return
	}
	event, err := h.create.Execute(c.Request.Context(), actions.CreateEventRequest{
		UserID:                userID,
		Title:                 req.Title,
		Description:           req.Description,
		StartAt:               startAt,
		EndAt:                 endAt,
		Timezone:              req.Timezone,
		AllDay:                req.AllDay,
		Location:              req.Location,
		ReminderOffsetMinutes: req.ReminderOffsetMinutes,
		RecurrenceRule:        req.RecurrenceRule,
		Color:                 req.Color,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, constant.MessageOK, donnaEventFromAction(event))
}

// Update handles PATCH /donna/events/:id.
func (h *DonnaEventHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	var req model.UpdateDonnaEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	startAt, err := parseOptionalRFC3339(req.StartAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "start_at must be RFC3339")
		return
	}
	endAt, err := parseOptionalRFC3339(req.EndAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "end_at must be RFC3339")
		return
	}
	event, err := h.update.Execute(c.Request.Context(), actions.UpdateEventRequest{
		UserID:                userID,
		EventID:               eventID,
		Title:                 req.Title,
		Description:           req.Description,
		StartAt:               startAt,
		EndAt:                 endAt,
		Timezone:              req.Timezone,
		AllDay:                req.AllDay,
		Location:              req.Location,
		ReminderOffsetMinutes: req.ReminderOffsetMinutes,
		RecurrenceRule:        req.RecurrenceRule,
		Status:                req.Status,
		Color:                 req.Color,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, donnaEventFromAction(event))
}

// Delete handles DELETE /donna/events/:id.
func (h *DonnaEventHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid event id", constant.ErrorCodeInvalidRequest, err.Error())
		return
	}
	if err := h.delete.Execute(c.Request.Context(), actions.DeleteEventRequest{
		UserID:  userID,
		EventID: eventID,
	}); err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, gin.H{"deleted": true})
}

func (h *DonnaEventHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "donna event request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "donna event failed")
	}
}

func parseOptionalRFC3339(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
