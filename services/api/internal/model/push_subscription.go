package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// SubscribePushRequest is POST /push/subscribe body.
type SubscribePushRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
	UserAgent *string `json:"user_agent,omitempty"`
}

// UnsubscribePushRequest is DELETE /push/unsubscribe body.
type UnsubscribePushRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

// PushSubscriptionResponse is the API shape for a stored subscription.
type PushSubscriptionResponse struct {
	ID        string  `json:"id"`
	PublicID  string  `json:"public_id"`
	Endpoint  string  `json:"endpoint"`
	UserAgent *string `json:"user_agent,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// PushSubscriptionFromEntity maps an entity to the API response (no secrets).
func PushSubscriptionFromEntity(s entity.PushSubscription) PushSubscriptionResponse {
	return PushSubscriptionResponse{
		ID:        s.ID.String(),
		PublicID:  s.PublicID,
		Endpoint:  s.Endpoint,
		UserAgent: s.UserAgent,
		CreatedAt: s.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: s.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// VAPIDPublicKeyResponse exposes the browser-facing VAPID key.
type VAPIDPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}
