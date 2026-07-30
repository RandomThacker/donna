package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// QueryTimelineRequest is the input for QueryTimelineAction.
type QueryTimelineRequest struct {
	UserID uuid.UUID
	From   time.Time
	To     time.Time
}

// QueryTimelineAction returns unified timeline items for a window.
type QueryTimelineAction struct {
	timeline TimelineQueryService
}

// NewQueryTimelineAction constructs QueryTimelineAction.
func NewQueryTimelineAction(timeline TimelineQueryService) *QueryTimelineAction {
	return &QueryTimelineAction{timeline: timeline}
}

// Execute runs the timeline query workflow.
func (a *QueryTimelineAction) Execute(ctx context.Context, req QueryTimelineRequest) (TimelineResult, error) {
	if a.timeline == nil {
		return TimelineResult{}, fmt.Errorf("%w: timeline service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return TimelineResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	from := req.From.UTC()
	to := req.To.UTC()
	if from.IsZero() || to.IsZero() {
		return TimelineResult{}, fmt.Errorf("%w: from and to are required", apperr.ErrValidation)
	}
	if !to.After(from) {
		return TimelineResult{}, fmt.Errorf("%w: to must be after from", apperr.ErrValidation)
	}
	items, err := a.timeline.List(ctx, req.UserID, from, to)
	if err != nil {
		return TimelineResult{}, err
	}
	out := make([]TimelineItemResult, 0, len(items))
	for _, item := range items {
		out = append(out, timelineItemFromEntity(item))
	}
	return TimelineResult{Items: out, From: from, To: to}, nil
}
