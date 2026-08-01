package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserPersonality is the per-user Personality Engine preference (1:1 with User).
type UserPersonality struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	PersonalityID      string
	DisplayName        *string
	Nickname           *string
	EmojiLevel         string
	HumorLevel         string
	GreetingStyle      string
	EncouragementLevel string
	ResponseStyle      string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
