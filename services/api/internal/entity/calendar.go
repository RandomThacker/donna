package entity

import (
	"time"

	"github.com/google/uuid"
)

// CalendarSource is a synced calendar feed under a connected account.
type CalendarSource struct {
	ID                  uuid.UUID
	PublicID            string
	UserID              uuid.UUID
	ConnectedAccountID  uuid.UUID
	ProviderCalendarID  string
	Name                string
	Color               *string
	IsPrimaryOnProvider bool
	IsWritable          bool
	AccessRole          *string
	SyncEnabled         bool
	SyncCursor          *string // per-source Google events.list syncToken
	LastSyncedAt        *time.Time
	Timezone            *string
	ProviderMetadata    []byte
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

// CalendarEvent is a Donna-canonical calendar event (synced or local).
type CalendarEvent struct {
	ID                       uuid.UUID
	PublicID                 string
	UserID                   uuid.UUID
	CalendarSourceID         uuid.UUID
	Title                    string
	Description              *string
	Location                 *string
	StartsAt                 time.Time
	EndsAt                   time.Time
	IsAllDay                 bool
	Status                   string
	Visibility               *string
	Timezone                 *string
	OrganizerSummary         []byte // jsonb object
	AttendeesSummary         []byte // jsonb array
	RecurrenceRule           *string
	RecurringEventID         *uuid.UUID // Donna parent series row
	ProviderRecurringEventID *string    // Google recurringEventId
	ProviderEventID          *string
	ProviderETag             *string
	ProviderUpdatedAt        *time.Time
	ProviderPayload          []byte
	Origin                   string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                *time.Time
}
