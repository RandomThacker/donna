package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// CreateTaskRequest is the input for CreateTaskAction.
type CreateTaskRequest struct {
	UserID         uuid.UUID
	Title          string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	TagIDs         []uuid.UUID
	RecurrenceRule *string
	Date           time.Time
	Source         string
}

// CreateTaskAction creates a task and its journal occurrence for a day.
type CreateTaskAction struct {
	tasks     TaskService
	publisher DomainEventPublisher
}

// NewCreateTaskAction constructs CreateTaskAction.
func NewCreateTaskAction(tasks TaskService, publisher DomainEventPublisher) *CreateTaskAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &CreateTaskAction{tasks: tasks, publisher: publisher}
}

// Execute runs the create-task workflow.
func (a *CreateTaskAction) Execute(ctx context.Context, req CreateTaskRequest) (TaskOccurrenceResult, error) {
	if a.tasks == nil {
		return TaskOccurrenceResult{}, fmt.Errorf("%w: task service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return TaskOccurrenceResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	created, err := a.tasks.CreateTask(ctx, req.UserID, business.CreateTaskInput{
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Project:        req.Project,
		Labels:         req.Labels,
		TagIDs:         req.TagIDs,
		RecurrenceRule: req.RecurrenceRule,
		Date:           req.Date,
		Source:         req.Source,
	})
	if err != nil {
		return TaskOccurrenceResult{}, err
	}
	result := taskOccurrenceFromEntity(created)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "task.created", Payload: result})
	return result, nil
}
