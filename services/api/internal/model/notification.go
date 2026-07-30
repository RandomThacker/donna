package model

import (
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// NotificationResponse is the API shape for a queued notification.
type NotificationResponse struct {
	ID                    string          `json:"id"`
	PublicID              string          `json:"public_id"`
	TimelineItemParentID  *string         `json:"timeline_item_parent_id,omitempty"`
	OccurrenceID          *string         `json:"occurrence_id,omitempty"`
	Title                 string          `json:"title"`
	Body                  string          `json:"body"`
	NotificationType      *string         `json:"notification_type,omitempty"`
	ScheduledFor          *string         `json:"scheduled_for,omitempty"`
	Status                string          `json:"status"`
	DeliveryChannels      []string        `json:"delivery_channels"`
	ChannelDeliveryStatus json.RawMessage `json:"channel_delivery_status,omitempty"`
	Payload               json.RawMessage `json:"payload,omitempty"`
	ReadAt                *string         `json:"read_at,omitempty"`
	DismissedAt           *string         `json:"dismissed_at,omitempty"`
	SentAt                *string         `json:"sent_at,omitempty"`
	CreatedAt             string          `json:"created_at"`
	UpdatedAt             string          `json:"updated_at"`
}

// NotificationFromEntity maps an entity notification to the API response.
func NotificationFromEntity(n entity.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:                    n.ID.String(),
		PublicID:              n.PublicID,
		TimelineItemParentID:  n.TimelineItemParentID,
		OccurrenceID:          n.OccurrenceID,
		Title:                 n.Title,
		Body:                  n.Body,
		NotificationType:      n.NotificationType,
		Status:                n.Status,
		DeliveryChannels:      n.DeliveryChannels,
		ChannelDeliveryStatus: n.ChannelDeliveryStatus,
		Payload:               n.Payload,
		CreatedAt:             n.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:             n.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if n.DeliveryChannels == nil {
		resp.DeliveryChannels = []string{}
	}
	if n.ScheduledFor != nil {
		v := n.ScheduledFor.UTC().Format(time.RFC3339Nano)
		resp.ScheduledFor = &v
	}
	if n.ReadAt != nil {
		v := n.ReadAt.UTC().Format(time.RFC3339Nano)
		resp.ReadAt = &v
	}
	if n.DismissedAt != nil {
		v := n.DismissedAt.UTC().Format(time.RFC3339Nano)
		resp.DismissedAt = &v
	}
	if n.SentAt != nil {
		v := n.SentAt.UTC().Format(time.RFC3339Nano)
		resp.SentAt = &v
	}
	return resp
}

// NotificationsFromEntities maps a slice of notifications.
func NotificationsFromEntities(items []entity.Notification) []NotificationResponse {
	out := make([]NotificationResponse, 0, len(items))
	for _, n := range items {
		out = append(out, NotificationFromEntity(n))
	}
	return out
}
