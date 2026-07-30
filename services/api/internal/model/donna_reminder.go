package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// CreateDonnaReminderRequest is POST /donna/reminders.
type CreateDonnaReminderRequest struct {
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	TriggerAt      string  `json:"trigger_at"`
	Timezone       string  `json:"timezone"`
	RecurrenceRule *string `json:"recurrence_rule"`
	Color          *string `json:"color"`
}

// UpdateDonnaReminderRequest is PATCH /donna/reminders/:id.
type UpdateDonnaReminderRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	TriggerAt      *string `json:"trigger_at"`
	Timezone       *string `json:"timezone"`
	RecurrenceRule *string `json:"recurrence_rule"`
	Status         *string `json:"status"`
	Color          *string `json:"color"`
}

// DonnaReminderResponse is the API shape for a Donna reminder.
type DonnaReminderResponse struct {
	ID             string  `json:"id"`
	PublicID       string  `json:"public_id"`
	Title          string  `json:"title"`
	Description    *string `json:"description,omitempty"`
	TriggerAt      string  `json:"trigger_at"`
	Timezone       string  `json:"timezone"`
	RecurrenceRule *string `json:"recurrence_rule,omitempty"`
	Status         string  `json:"status"`
	Color          *string `json:"color,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// DonnaReminderFromEntity maps an entity to the API response.
func DonnaReminderFromEntity(r entity.DonnaReminder) DonnaReminderResponse {
	return DonnaReminderResponse{
		ID:             r.ID.String(),
		PublicID:       r.PublicID,
		Title:          r.Title,
		Description:    r.Description,
		TriggerAt:      r.TriggerAt.UTC().Format(time.RFC3339Nano),
		Timezone:       r.Timezone,
		RecurrenceRule: r.RecurrenceRule,
		Status:         r.Status,
		Color:          r.Color,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// DonnaRemindersFromEntities maps a slice of Donna reminders.
func DonnaRemindersFromEntities(reminders []entity.DonnaReminder) []DonnaReminderResponse {
	out := make([]DonnaReminderResponse, 0, len(reminders))
	for _, r := range reminders {
		out = append(out, DonnaReminderFromEntity(r))
	}
	return out
}
