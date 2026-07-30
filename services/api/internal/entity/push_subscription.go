package entity

import (
	"time"

	"github.com/google/uuid"
)

// PushSubscription is a browser Web Push endpoint for one device.
type PushSubscription struct {
	ID        uuid.UUID
	PublicID  string
	UserID    uuid.UUID
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
