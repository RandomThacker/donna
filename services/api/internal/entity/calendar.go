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
	SyncCursor          *string
	LastSyncedAt        *time.Time
	Timezone            *string
	ProviderMetadata    []byte
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}
