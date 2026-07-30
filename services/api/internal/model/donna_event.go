package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateDonnaEventRequest is POST /donna/events.
type CreateDonnaEventRequest struct {
	Title                 string  `json:"title"`
	Description           *string `json:"description"`
	StartAt               string  `json:"start_at"`
	EndAt                 string  `json:"end_at"`
	Timezone              string  `json:"timezone"`
	AllDay                bool    `json:"all_day"`
	Location              *string `json:"location"`
	ReminderOffsetMinutes *int    `json:"reminder_offset_minutes"`
	RecurrenceRule        *string `json:"recurrence_rule"`
	Color                 *string `json:"color"`
}

// UpdateDonnaEventRequest is PATCH /donna/events/:id.
type UpdateDonnaEventRequest struct {
	Title                 *string `json:"title"`
	Description           *string `json:"description"`
	StartAt               *string `json:"start_at"`
	EndAt                 *string `json:"end_at"`
	Timezone              *string `json:"timezone"`
	AllDay                *bool   `json:"all_day"`
	Location              *string `json:"location"`
	ReminderOffsetMinutes *int    `json:"reminder_offset_minutes"`
	RecurrenceRule        *string `json:"recurrence_rule"`
	Status                *string `json:"status"`
	Color                 *string `json:"color"`
}

// DonnaEventResponse is the API shape for a Donna event.
type DonnaEventResponse struct {
	ID                    string  `json:"id"`
	PublicID              string  `json:"public_id"`
	Title                 string  `json:"title"`
	Description           *string `json:"description,omitempty"`
	StartAt               string  `json:"start_at"`
	EndAt                 string  `json:"end_at"`
	Timezone              string  `json:"timezone"`
	AllDay                bool    `json:"all_day"`
	Location              *string `json:"location,omitempty"`
	ReminderOffsetMinutes *int    `json:"reminder_offset_minutes,omitempty"`
	RecurrenceRule        *string `json:"recurrence_rule,omitempty"`
	Status                string  `json:"status"`
	Color                 *string `json:"color,omitempty"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

// DonnaEventFromEntity maps an entity to the API response.
func DonnaEventFromEntity(e entity.DonnaEvent) DonnaEventResponse {
	return DonnaEventResponse{
		ID:                    e.ID.String(),
		PublicID:              e.PublicID,
		Title:                 e.Title,
		Description:           e.Description,
		StartAt:               e.StartAt.UTC().Format(time.RFC3339Nano),
		EndAt:                 e.EndAt.UTC().Format(time.RFC3339Nano),
		Timezone:              e.Timezone,
		AllDay:                e.AllDay,
		Location:              e.Location,
		ReminderOffsetMinutes: e.ReminderOffsetMinutes,
		RecurrenceRule:        e.RecurrenceRule,
		Status:                e.Status,
		Color:                 e.Color,
		CreatedAt:             e.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:             e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// DonnaEventsFromEntities maps a slice of Donna events.
func DonnaEventsFromEntities(events []entity.DonnaEvent) []DonnaEventResponse {
	out := make([]DonnaEventResponse, 0, len(events))
	for _, e := range events {
		out = append(out, DonnaEventFromEntity(e))
	}
	return out
}
