package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/response"
	"github.com/gin-gonic/gin"
)

// TimelineHandler maps timeline HTTP endpoints to Actions.
type TimelineHandler struct {
	query *actions.QueryTimelineAction
	log   *logger.Logger
}

// NewTimelineHandler constructs a TimelineHandler.
func NewTimelineHandler(query *actions.QueryTimelineAction, log *logger.Logger) *TimelineHandler {
	return &TimelineHandler{query: query, log: log}
}

// List handles GET /timeline?from=&to=.
func (h *TimelineHandler) List(c *gin.Context) {
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

	result, err := h.query.Execute(c.Request.Context(), actions.QueryTimelineRequest{
		UserID: userID,
		From:   from,
		To:     to,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.OK(c, constant.MessageOK, timelineFromAction(result))
}

func (h *TimelineHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrValidation):
		response.Error(c, http.StatusBadRequest, "validation failed", constant.ErrorCodeValidation, err.Error())
	case errors.Is(err, apperr.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not found", constant.ErrorCodeNotFound, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", constant.ErrorCodeForbidden, err.Error())
	default:
		if h.log != nil {
			h.log.Error(c.Request.Context(), "timeline request failed", constant.LogAttrError, err)
		}
		response.Error(c, http.StatusInternalServerError, "internal error", constant.ErrorCodeInternal, "timeline failed")
	}
}
