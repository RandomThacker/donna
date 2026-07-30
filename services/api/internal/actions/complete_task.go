package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// CompleteTaskRequest is the input for CompleteTaskAction.
type CompleteTaskRequest struct {
	UserID       uuid.UUID
	OccurrenceID uuid.UUID
	Completed    bool
}

// CompleteTaskAction marks a task occurrence complete or incomplete.
type CompleteTaskAction struct {
	tasks     TaskService
	publisher DomainEventPublisher
}

// NewCompleteTaskAction constructs CompleteTaskAction.
func NewCompleteTaskAction(tasks TaskService, publisher DomainEventPublisher) *CompleteTaskAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &CompleteTaskAction{tasks: tasks, publisher: publisher}
}

// Execute runs the complete/incomplete occurrence workflow.
func (a *CompleteTaskAction) Execute(ctx context.Context, req CompleteTaskRequest) (TaskOccurrenceResult, error) {
	if a.tasks == nil {
		return TaskOccurrenceResult{}, fmt.Errorf("%w: task service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.OccurrenceID == uuid.Nil {
		return TaskOccurrenceResult{}, fmt.Errorf("%w: user and occurrence id are required", apperr.ErrValidation)
	}
	updated, err := a.tasks.UpdateOccurrence(ctx, req.UserID, req.OccurrenceID, req.Completed)
	if err != nil {
		return TaskOccurrenceResult{}, err
	}
	result := taskOccurrenceFromEntity(updated)
	name := "task.uncompleted"
	if req.Completed {
		name = "task.completed"
	}
	_ = a.publisher.Publish(ctx, DomainEvent{Name: name, Payload: result})
	return result, nil
}
