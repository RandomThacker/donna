package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// GetNotificationsRequest is the input for GetNotificationsAction.
type GetNotificationsRequest struct {
	UserID   uuid.UUID
	Statuses []string
}

// GetNotificationsAction lists notifications for a user.
type GetNotificationsAction struct {
	notifications NotificationQueryService
}

// NewGetNotificationsAction constructs GetNotificationsAction.
func NewGetNotificationsAction(notifications NotificationQueryService) *GetNotificationsAction {
	return &GetNotificationsAction{notifications: notifications}
}

// Execute runs the list-notifications workflow.
func (a *GetNotificationsAction) Execute(ctx context.Context, req GetNotificationsRequest) ([]NotificationResult, error) {
	if a.notifications == nil {
		return nil, fmt.Errorf("%w: notification service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	items, err := a.notifications.List(ctx, req.UserID, req.Statuses)
	if err != nil {
		return nil, err
	}
	out := make([]NotificationResult, 0, len(items))
	for _, n := range items {
		out = append(out, notificationFromEntity(n))
	}
	return out, nil
}
