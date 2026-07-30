package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// UpdateTaskRequest is the input for UpdateTaskAction.
type UpdateTaskRequest struct {
	UserID         uuid.UUID
	TaskID         uuid.UUID
	Title          *string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	TagIDs         *[]uuid.UUID
	RecurrenceRule *string
}

// UpdateTaskAction patches permanent task fields and returns tags.
type UpdateTaskAction struct {
	tasks     TaskService
	publisher DomainEventPublisher
}

// NewUpdateTaskAction constructs UpdateTaskAction.
func NewUpdateTaskAction(tasks TaskService, publisher DomainEventPublisher) *UpdateTaskAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &UpdateTaskAction{tasks: tasks, publisher: publisher}
}

// Execute runs the update-task workflow (update + load tags).
func (a *UpdateTaskAction) Execute(ctx context.Context, req UpdateTaskRequest) (TaskResult, error) {
	if a.tasks == nil {
		return TaskResult{}, fmt.Errorf("%w: task service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.TaskID == uuid.Nil {
		return TaskResult{}, fmt.Errorf("%w: user and task id are required", apperr.ErrValidation)
	}
	updated, err := a.tasks.UpdateTask(ctx, req.UserID, req.TaskID, business.UpdateTaskInput{
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Project:        req.Project,
		Labels:         req.Labels,
		TagIDs:         req.TagIDs,
		RecurrenceRule: req.RecurrenceRule,
	})
	if err != nil {
		return TaskResult{}, err
	}
	tags, err := a.tasks.ListTaskTagsForTask(ctx, req.UserID, req.TaskID)
	if err != nil {
		return TaskResult{}, err
	}
	result := taskFromEntity(updated, tags)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "task.updated", Payload: result})
	return result, nil
}
