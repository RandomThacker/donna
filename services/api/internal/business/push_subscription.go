package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// SubscribePushInput is the browser PushSubscriptionJSON keys.
type SubscribePushInput struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent *string
}

// PushSubscriptionService manages per-user Web Push endpoints.
type PushSubscriptionService struct {
	subs repository.PushSubscriptionRepository
	now  func() time.Time
}

// NewPushSubscriptionService constructs a PushSubscriptionService.
func NewPushSubscriptionService(subs repository.PushSubscriptionRepository) *PushSubscriptionService {
	return &PushSubscriptionService{subs: subs, now: time.Now}
}

// Subscribe upserts a device subscription for the user.
func (s *PushSubscriptionService) Subscribe(ctx context.Context, userID uuid.UUID, in SubscribePushInput) (entity.PushSubscription, error) {
	if userID == uuid.Nil {
		return entity.PushSubscription{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	endpoint := strings.TrimSpace(in.Endpoint)
	p256dh := strings.TrimSpace(in.P256dh)
	auth := strings.TrimSpace(in.Auth)
	if endpoint == "" || p256dh == "" || auth == "" {
		return entity.PushSubscription{}, fmt.Errorf("%w: endpoint, p256dh, and auth are required", apperr.ErrValidation)
	}

	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.PushSubscription{}, err
	}
	now := s.now().UTC()
	var ua *string
	if in.UserAgent != nil {
		trimmed := strings.TrimSpace(*in.UserAgent)
		if trimmed != "" {
			ua = &trimmed
		}
	}
	return s.subs.Upsert(ctx, entity.PushSubscription{
		ID:        id,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixPushSubscription, id),
		UserID:    userID,
		Endpoint:  endpoint,
		P256dh:    p256dh,
		Auth:      auth,
		UserAgent: ua,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// Unsubscribe soft-deletes a subscription by endpoint for the user.
func (s *PushSubscriptionService) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return fmt.Errorf("%w: endpoint is required", apperr.ErrValidation)
	}
	return s.subs.SoftDeleteByEndpoint(ctx, userID, endpoint, s.now().UTC())
}

// List returns live subscriptions for a user.
func (s *PushSubscriptionService) List(ctx context.Context, userID uuid.UUID) ([]entity.PushSubscription, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.subs.ListByUser(ctx, userID)
}
