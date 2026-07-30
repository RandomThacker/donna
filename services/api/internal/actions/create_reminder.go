package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// CreateReminderRequest is the input for CreateReminderAction.
type CreateReminderRequest struct {
	UserID         uuid.UUID
	Title          string
	Description    *string
	TriggerAt      time.Time
	Timezone       string
	RecurrenceRule *string
	Color          *string
}

// CreateReminderAction creates a Donna-owned reminder.
type CreateReminderAction struct {
	reminders ReminderService
	publisher DomainEventPublisher
}

// NewCreateReminderAction constructs CreateReminderAction.
func NewCreateReminderAction(reminders ReminderService, publisher DomainEventPublisher) *CreateReminderAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &CreateReminderAction{reminders: reminders, publisher: publisher}
}

// Execute runs the create-reminder workflow.
func (a *CreateReminderAction) Execute(ctx context.Context, req CreateReminderRequest) (ReminderResult, error) {
	if a.reminders == nil {
		return ReminderResult{}, fmt.Errorf("%w: reminder service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return ReminderResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	created, err := a.reminders.Create(ctx, req.UserID, business.CreateDonnaReminderInput{
		Title:          req.Title,
		Description:    req.Description,
		TriggerAt:      req.TriggerAt,
		Timezone:       req.Timezone,
		RecurrenceRule: req.RecurrenceRule,
		Color:          req.Color,
	})
	if err != nil {
		return ReminderResult{}, err
	}
	result := reminderFromEntity(created)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "reminder.created", Payload: result})
	return result, nil
}
