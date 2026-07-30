package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// CreateEventRequest is the input for CreateEventAction.
type CreateEventRequest struct {
	UserID                uuid.UUID
	Title                 string
	Description           *string
	StartAt               time.Time
	EndAt                 time.Time
	Timezone              string
	AllDay                bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Color                 *string
}

// CreateEventAction creates a Donna-owned event.
type CreateEventAction struct {
	events    EventService
	publisher DomainEventPublisher
}

// NewCreateEventAction constructs CreateEventAction.
func NewCreateEventAction(events EventService, publisher DomainEventPublisher) *CreateEventAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &CreateEventAction{events: events, publisher: publisher}
}

// Execute runs the create-event workflow.
func (a *CreateEventAction) Execute(ctx context.Context, req CreateEventRequest) (EventResult, error) {
	if a.events == nil {
		return EventResult{}, fmt.Errorf("%w: event service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return EventResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	created, err := a.events.Create(ctx, req.UserID, business.CreateDonnaEventInput{
		Title:                 req.Title,
		Description:           req.Description,
		StartAt:               req.StartAt,
		EndAt:                 req.EndAt,
		Timezone:              req.Timezone,
		AllDay:                req.AllDay,
		Location:              req.Location,
		ReminderOffsetMinutes: req.ReminderOffsetMinutes,
		RecurrenceRule:        req.RecurrenceRule,
		Color:                 req.Color,
	})
	if err != nil {
		return EventResult{}, err
	}
	result := eventFromEntity(created)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "event.created", Payload: result})
	return result, nil
}
