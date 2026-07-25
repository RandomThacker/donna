package entity

import (
	"time"

	"github.com/google/uuid"
)

// User is the Identity aggregate root (Donna account).
type User struct {
	ID            uuid.UUID
	PublicID      string
	Email         string
	EmailVerified bool
	DisplayName   *string
	AvatarURL     *string
	Timezone      string
	Locale        *string
	Status        string
	LastLoginAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
