package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// ListDayTasksRequest lists journal occurrences for one civil day.
type ListDayTasksRequest struct {
	UserID uuid.UUID
	Date   time.Time
}

// ListDayTasksResult is the Action DTO for a day of tasks.
type ListDayTasksResult struct {
	Date        time.Time
	Occurrences []TaskOccurrenceResult
}

// TaskDayLister lists journal occurrences for chat task resolution.
type TaskDayLister interface {
	ListDayTaskOccurrences(ctx context.Context, userID uuid.UUID, date time.Time) ([]TaskOccurrenceResult, error)
}

// ListDayTasksAction returns task occurrences for a civil date (chat resolution).
type ListDayTasksAction struct {
	tasks TaskDayLister
}

// NewListDayTasksAction constructs ListDayTasksAction.
func NewListDayTasksAction(tasks TaskDayLister) *ListDayTasksAction {
	return &ListDayTasksAction{tasks: tasks}
}

// Execute lists tasks for the given day.
func (a *ListDayTasksAction) Execute(ctx context.Context, req ListDayTasksRequest) (ListDayTasksResult, error) {
	if a.tasks == nil {
		return ListDayTasksResult{}, fmt.Errorf("%w: task day lister is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return ListDayTasksResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if req.Date.IsZero() {
		return ListDayTasksResult{}, fmt.Errorf("%w: date is required", apperr.ErrValidation)
	}
	occs, err := a.tasks.ListDayTaskOccurrences(ctx, req.UserID, req.Date)
	if err != nil {
		return ListDayTasksResult{}, err
	}
	return ListDayTasksResult{Date: req.Date.UTC(), Occurrences: occs}, nil
}
