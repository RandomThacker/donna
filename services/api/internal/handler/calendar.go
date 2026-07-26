package handler

import (
	"errors"
	"net/http"
	"time"

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

// CalendarHandler maps calendar HTTP endpoints to the calendar business layer.
type CalendarHandler struct {
	svc *business.CalendarService
	log *logger.Logger
}

// NewCalendarHandler constructs a CalendarHandler.
func NewCalendarHandler(svc *business.CalendarService, log *logger.Logger) *CalendarHandler {
	return &CalendarHandler{svc: svc, log: log}
}

// SyncSources handles POST /calendar/sync (orchestrated sources + events sync).
func (h *CalendarHandler) SyncSources(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	result, err := h.svc.SyncSources(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageCalendarSynced, calendarPipelineResponse(result))
}

// EnsureFreshSources handles POST /calendar/sync/ensure (startup / on-demand stale check).
func (h *CalendarHandler) EnsureFreshSources(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	result, err := h.svc.EnsureFresh(c.Request.Context(), userID, constant.CalendarSyncStaleAfter)
	if err != nil {
		h.writeError(c, err)
		return
	}
	msg := constant.MessageOK
	if !result.Skipped {
		msg = constant.MessageCalendarSynced
	}
	response.OK(c, msg, calendarPipelineResponse(result))
}

// ListSources handles GET /calendar/sources (reads Donna DB only).
func (h *CalendarHandler) ListSources(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	sources, err := h.svc.ListSources(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.CalendarSourcesResponse{
		Sources: model.CalendarSourcesFromEntities(sources.Sources),
		Sync:    model.CalendarAccountSyncFromEntity(sources.Account),
	})
}

// SyncEvents handles POST /calendar/events/sync.
func (h *CalendarHandler) SyncEvents(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	result, err := h.svc.SyncEvents(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageCalendarEventsSynced, calendarEventSyncResponse(result))
}

// ListEvents handles GET /calendar/events (Donna DB only).
func (h *CalendarHandler) ListEvents(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required", constant.ErrorCodeUnauthorized, "missing user")
		return
	}

	now := time.Now().UTC()
	from := now.Add(-7 * 24 * time.Hour)
	to := now.Add(30 * 24 * time.Hour)
	if raw := c.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "from must be RFC3339")
			return
		}
		from = parsed.UTC()
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, "to must be RFC3339")
			return
		}
		to = parsed.UTC()
	}

	events, err := h.svc.ListEvents(c.Request.Context(), userID, from, to)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, model.CalendarEventsResponse{
		Events: model.CalendarEventsFromEntities(events),
		From:   from.Format(time.RFC3339Nano),
		To:     to.Format(time.RFC3339Nano),
	})
}

func calendarPipelineResponse(result business.CalendarPipelineResult) model.CalendarSyncResponse {
	started := result.StartedAt
	finished := result.FinishedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	if finished.IsZero() {
		finished = started
	}
	failures := make([]model.CalendarSyncFailureResponse, 0, len(result.Failures))
	for _, f := range result.Failures {
		failures = append(failures, model.CalendarSyncFailureResponse{
			CalendarSourceID:   f.CalendarSourceID,
			ProviderCalendarID: f.ProviderCalendarID,
			Name:               f.Name,
			Stage:              f.Stage,
			Error:              f.Error,
		})
	}
	resp := model.CalendarSyncResponse{
		Trigger:            result.Trigger,
		Status:             result.Status,
		StartedAt:          started.UTC().Format(time.RFC3339Nano),
		FinishedAt:         finished.UTC().Format(time.RFC3339Nano),
		DurationMs:         result.DurationMs,
		CalendarsProcessed: result.CalendarsProcessed,
		SourcesCreated:     result.SourcesCreated,
		SourcesUpdated:     result.SourcesUpdated,
		SourcesDeleted:     result.SourcesDeleted,
		EventsCreated:      result.EventsCreated,
		EventsUpdated:      result.EventsUpdated,
		EventsDeleted:      result.EventsDeleted,
		Failures:           failures,
		Sources:            model.CalendarSourcesFromEntities(result.Sources),
		Incremental:        result.Incremental,
		Skipped:            result.Skipped,
		SyncStatus:         result.SyncStatus,
	}
	if result.RunID != (uuid.UUID{}) {
		resp.RunID = result.RunID.String()
	}
	return resp
}

func calendarEventSyncResponse(result business.CalendarEventSyncResult) model.CalendarEventSyncResponse {
	syncedAt := result.SyncedAt
	if syncedAt.IsZero() {
		syncedAt = time.Now().UTC()
	}
	return model.CalendarEventSyncResponse{
		Events:       model.CalendarEventsFromEntities(result.Events),
		CreatedCount: result.CreatedCount,
		UpdatedCount: result.UpdatedCount,
		RemovedCount: result.RemovedCount,
		SyncedAt:     syncedAt.UTC().Format(time.RFC3339Nano),
		DurationMs:   result.DurationMs,
		SourceCount:  result.SourceCount,
	}
}

func (h *CalendarHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	case errors.Is(err, apperr.ErrInvalid):
		response.Error(c, http.StatusBadRequest, "invalid request", constant.ErrorCodeInvalidRequest, err.Error())
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "calendar request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "unexpected error")
	}
}
