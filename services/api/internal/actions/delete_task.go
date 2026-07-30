package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// DeleteTaskRequest is the input for DeleteTaskAction.
type DeleteTaskRequest struct {
	UserID uuid.UUID
	TaskID uuid.UUID
}

// DeleteTaskAction soft-deletes a task and its occurrences.
type DeleteTaskAction struct {
	tasks     TaskService
	publisher DomainEventPublisher
}

// NewDeleteTaskAction constructs DeleteTaskAction.
func NewDeleteTaskAction(tasks TaskService, publisher DomainEventPublisher) *DeleteTaskAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &DeleteTaskAction{tasks: tasks, publisher: publisher}
}

// Execute runs the delete-task workflow.
func (a *DeleteTaskAction) Execute(ctx context.Context, req DeleteTaskRequest) error {
	if a.tasks == nil {
		return fmt.Errorf("%w: task service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.TaskID == uuid.Nil {
		return fmt.Errorf("%w: user and task id are required", apperr.ErrValidation)
	}
	if err := a.tasks.DeleteTask(ctx, req.UserID, req.TaskID); err != nil {
		return err
	}
	_ = a.publisher.Publish(ctx, DomainEvent{
		Name:    "task.deleted",
		Payload: map[string]string{"task_id": req.TaskID.String()},
	})
	return nil
}
