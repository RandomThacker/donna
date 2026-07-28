package entity

import (
	"time"

	"github.com/google/uuid"
)

// CalendarSyncRun is one orchestration attempt of sources + events sync.
type CalendarSyncRun struct {
	ID                 uuid.UUID
	PublicID           string
	UserID             uuid.UUID
	ConnectedAccountID uuid.UUID
	Trigger            string
	Status             string
	StartedAt          time.Time
	FinishedAt         *time.Time
	DurationMs         *int
	CalendarsProcessed int
	SourcesCreated     int
	SourcesUpdated     int
	SourcesDeleted     int
	EventsCreated      int
	EventsUpdated      int
	EventsDeleted      int
	Failures           []byte // jsonb array
	CreatedAt          time.Time
}
