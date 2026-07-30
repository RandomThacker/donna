package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// DeleteReminderRequest is the input for DeleteReminderAction.
type DeleteReminderRequest struct {
	UserID     uuid.UUID
	ReminderID uuid.UUID
}

// DeleteReminderAction soft-deletes a Donna-owned reminder.
type DeleteReminderAction struct {
	reminders ReminderService
	publisher DomainEventPublisher
}

// NewDeleteReminderAction constructs DeleteReminderAction.
func NewDeleteReminderAction(reminders ReminderService, publisher DomainEventPublisher) *DeleteReminderAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &DeleteReminderAction{reminders: reminders, publisher: publisher}
}

// Execute runs the delete-reminder workflow.
func (a *DeleteReminderAction) Execute(ctx context.Context, req DeleteReminderRequest) error {
	if a.reminders == nil {
		return fmt.Errorf("%w: reminder service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.ReminderID == uuid.Nil {
		return fmt.Errorf("%w: user and reminder id are required", apperr.ErrValidation)
	}
	if err := a.reminders.Delete(ctx, req.UserID, req.ReminderID); err != nil {
		return err
	}
	_ = a.publisher.Publish(ctx, DomainEvent{
		Name:    "reminder.deleted",
		Payload: map[string]string{"reminder_id": req.ReminderID.String()},
	})
	return nil
}
