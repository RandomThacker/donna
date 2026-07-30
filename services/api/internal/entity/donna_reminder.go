package entity

import (
	"time"

	"github.com/google/uuid"
)

// DonnaReminder is a user-owned standalone reminder that lives only in Donna.
type DonnaReminder struct {
	ID             uuid.UUID
	PublicID       string
	UserID         uuid.UUID
	Title          string
	Description    *string
	TriggerAt      time.Time
	Timezone       string
	RecurrenceRule *string
	Status         string
	Color          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
