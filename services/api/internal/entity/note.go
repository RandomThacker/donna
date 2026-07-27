package entity

import (
	"time"

	"github.com/google/uuid"
)

// Note is a user-owned Keep-style quick note.
type Note struct {
	ID        uuid.UUID
	PublicID  string
	UserID    uuid.UUID
	Title     string
	Content   string
	Color     string
	Pinned    bool
	Archived  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
