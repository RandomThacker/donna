package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// UpdateReminderRequest is the input for UpdateReminderAction.
type UpdateReminderRequest struct {
	UserID         uuid.UUID
	ReminderID     uuid.UUID
	Title          *string
	Description    *string
	TriggerAt      *time.Time
	Timezone       *string
	RecurrenceRule *string
	Status         *string
	Color          *string
}

// UpdateReminderAction patches a Donna-owned reminder.
type UpdateReminderAction struct {
	reminders ReminderService
	publisher DomainEventPublisher
}

// NewUpdateReminderAction constructs UpdateReminderAction.
func NewUpdateReminderAction(reminders ReminderService, publisher DomainEventPublisher) *UpdateReminderAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &UpdateReminderAction{reminders: reminders, publisher: publisher}
}

// Execute runs the update-reminder workflow.
func (a *UpdateReminderAction) Execute(ctx context.Context, req UpdateReminderRequest) (ReminderResult, error) {
	if a.reminders == nil {
		return ReminderResult{}, fmt.Errorf("%w: reminder service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.ReminderID == uuid.Nil {
		return ReminderResult{}, fmt.Errorf("%w: user and reminder id are required", apperr.ErrValidation)
	}
	updated, err := a.reminders.Update(ctx, req.UserID, req.ReminderID, business.UpdateDonnaReminderInput{
		Title:          req.Title,
		Description:    req.Description,
		TriggerAt:      req.TriggerAt,
		Timezone:       req.Timezone,
		RecurrenceRule: req.RecurrenceRule,
		Status:         req.Status,
		Color:          req.Color,
	})
	if err != nil {
		return ReminderResult{}, err
	}
	result := reminderFromEntity(updated)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "reminder.updated", Payload: result})
	return result, nil
}
