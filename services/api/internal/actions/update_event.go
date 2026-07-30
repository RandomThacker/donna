package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

// UpdateEventRequest is the input for UpdateEventAction.
type UpdateEventRequest struct {
	UserID                uuid.UUID
	EventID               uuid.UUID
	Title                 *string
	Description           *string
	StartAt               *time.Time
	EndAt                 *time.Time
	Timezone              *string
	AllDay                *bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Status                *string
	Color                 *string
}

// UpdateEventAction patches a Donna-owned event.
type UpdateEventAction struct {
	events    EventService
	publisher DomainEventPublisher
}

// NewUpdateEventAction constructs UpdateEventAction.
func NewUpdateEventAction(events EventService, publisher DomainEventPublisher) *UpdateEventAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &UpdateEventAction{events: events, publisher: publisher}
}

// Execute runs the update-event workflow.
func (a *UpdateEventAction) Execute(ctx context.Context, req UpdateEventRequest) (EventResult, error) {
	if a.events == nil {
		return EventResult{}, fmt.Errorf("%w: event service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.EventID == uuid.Nil {
		return EventResult{}, fmt.Errorf("%w: user and event id are required", apperr.ErrValidation)
	}
	updated, err := a.events.Update(ctx, req.UserID, req.EventID, business.UpdateDonnaEventInput{
		Title:                 req.Title,
		Description:           req.Description,
		StartAt:               req.StartAt,
		EndAt:                 req.EndAt,
		Timezone:              req.Timezone,
		AllDay:                req.AllDay,
		Location:              req.Location,
		ReminderOffsetMinutes: req.ReminderOffsetMinutes,
		RecurrenceRule:        req.RecurrenceRule,
		Status:                req.Status,
		Color:                 req.Color,
	})
	if err != nil {
		return EventResult{}, err
	}
	result := eventFromEntity(updated)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "event.updated", Payload: result})
	return result, nil
}
