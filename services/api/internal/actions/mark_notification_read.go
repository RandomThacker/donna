package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// MarkNotificationReadRequest is the input for MarkNotificationReadAction.
type MarkNotificationReadRequest struct {
	UserID         uuid.UUID
	NotificationID uuid.UUID
}

// MarkNotificationReadAction marks a notification as READ.
type MarkNotificationReadAction struct {
	notifications NotificationQueryService
	publisher     DomainEventPublisher
}

// NewMarkNotificationReadAction constructs MarkNotificationReadAction.
func NewMarkNotificationReadAction(
	notifications NotificationQueryService,
	publisher DomainEventPublisher,
) *MarkNotificationReadAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &MarkNotificationReadAction{notifications: notifications, publisher: publisher}
}

// Execute runs the mark-read workflow.
func (a *MarkNotificationReadAction) Execute(ctx context.Context, req MarkNotificationReadRequest) (NotificationResult, error) {
	if a.notifications == nil {
		return NotificationResult{}, fmt.Errorf("%w: notification service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.NotificationID == uuid.Nil {
		return NotificationResult{}, fmt.Errorf("%w: user and notification id are required", apperr.ErrValidation)
	}
	n, err := a.notifications.MarkRead(ctx, req.UserID, req.NotificationID)
	if err != nil {
		return NotificationResult{}, err
	}
	result := notificationFromEntity(n)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "notification.read", Payload: result})
	return result, nil
}
