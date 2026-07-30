package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// DeleteEventRequest is the input for DeleteEventAction.
type DeleteEventRequest struct {
	UserID  uuid.UUID
	EventID uuid.UUID
}

// DeleteEventAction soft-deletes a Donna-owned event.
type DeleteEventAction struct {
	events    EventService
	publisher DomainEventPublisher
}

// NewDeleteEventAction constructs DeleteEventAction.
func NewDeleteEventAction(events EventService, publisher DomainEventPublisher) *DeleteEventAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &DeleteEventAction{events: events, publisher: publisher}
}

// Execute runs the delete-event workflow.
func (a *DeleteEventAction) Execute(ctx context.Context, req DeleteEventRequest) error {
	if a.events == nil {
		return fmt.Errorf("%w: event service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.EventID == uuid.Nil {
		return fmt.Errorf("%w: user and event id are required", apperr.ErrValidation)
	}
	if err := a.events.Delete(ctx, req.UserID, req.EventID); err != nil {
		return err
	}
	_ = a.publisher.Publish(ctx, DomainEvent{
		Name:    "event.deleted",
		Payload: map[string]string{"event_id": req.EventID.String()},
	})
	return nil
}
