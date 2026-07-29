package entity

import (
	"time"

	"github.com/google/uuid"
)

// TaskTag is a user-defined colored label for tasks.
type TaskTag struct {
	ID        uuid.UUID
	PublicID  string
	UserID    uuid.UUID
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
