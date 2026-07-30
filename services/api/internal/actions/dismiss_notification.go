package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// DismissNotificationRequest is the input for DismissNotificationAction.
type DismissNotificationRequest struct {
	UserID         uuid.UUID
	NotificationID uuid.UUID
}

// DismissNotificationAction marks a notification as DISMISSED.
type DismissNotificationAction struct {
	notifications NotificationQueryService
	publisher     DomainEventPublisher
}

// NewDismissNotificationAction constructs DismissNotificationAction.
func NewDismissNotificationAction(
	notifications NotificationQueryService,
	publisher DomainEventPublisher,
) *DismissNotificationAction {
	if publisher == nil {
		publisher = NoopPublisher{}
	}
	return &DismissNotificationAction{notifications: notifications, publisher: publisher}
}

// Execute runs the dismiss workflow.
func (a *DismissNotificationAction) Execute(ctx context.Context, req DismissNotificationRequest) (NotificationResult, error) {
	if a.notifications == nil {
		return NotificationResult{}, fmt.Errorf("%w: notification service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.NotificationID == uuid.Nil {
		return NotificationResult{}, fmt.Errorf("%w: user and notification id are required", apperr.ErrValidation)
	}
	n, err := a.notifications.MarkDismissed(ctx, req.UserID, req.NotificationID)
	if err != nil {
		return NotificationResult{}, err
	}
	result := notificationFromEntity(n)
	_ = a.publisher.Publish(ctx, DomainEvent{Name: "notification.dismissed", Payload: result})
	return result, nil
}
