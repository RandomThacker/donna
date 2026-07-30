package entity

import (
	"time"

	"github.com/google/uuid"
)

// DonnaEvent is a user-owned calendar event that lives only in Donna.
type DonnaEvent struct {
	ID                     uuid.UUID
	PublicID               string
	UserID                 uuid.UUID
	Title                  string
	Description            *string
	StartAt                time.Time
	EndAt                  time.Time
	Timezone               string
	AllDay                 bool
	Location               *string
	ReminderOffsetMinutes  *int
	RecurrenceRule         *string
	Status                 string
	Color                  *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time
}
