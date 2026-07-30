package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Notification is a timeline-derived alert queued for later delivery.
type Notification struct {
	ID                     uuid.UUID
	PublicID               string
	UserID                 uuid.UUID
	TimelineItemParentID   *string
	OccurrenceID           *string
	Title                  string
	Body                   string
	NotificationType       *string
	ScheduledFor           *time.Time
	Status                 string
	DeliveryChannels       []string
	ChannelDeliveryStatus  json.RawMessage // per-channel map
	Payload                json.RawMessage
	ReadAt                 *time.Time
	DismissedAt            *time.Time
	SentAt                 *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time

	// Legacy single-channel column retained for schema compatibility.
	Channel  string
	Priority string
}
