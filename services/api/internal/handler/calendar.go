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

// SyncSources handles POST /calendar/sync (always available manual sync).
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
	response.OK(c, constant.MessageCalendarSynced, calendarSyncResponse(result))
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
	response.OK(c, msg, calendarSyncResponse(result))
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

func calendarSyncResponse(result business.CalendarSyncResult) model.CalendarSyncResponse {
	syncedAt := result.SyncedAt
	if syncedAt.IsZero() {
		syncedAt = time.Now().UTC()
	}
	return model.CalendarSyncResponse{
		Sources:      model.CalendarSourcesFromEntities(result.Sources),
		CreatedCount: result.CreatedCount,
		UpdatedCount: result.UpdatedCount,
		RemovedCount: result.RemovedCount,
		SyncedAt:     syncedAt.UTC().Format(time.RFC3339Nano),
		DurationMs:   result.DurationMs,
		Incremental:  result.Incremental,
		Skipped:      result.Skipped,
		SyncStatus:   result.SyncStatus,
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
