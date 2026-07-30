package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Conversation is a chat thread with Donna.
type Conversation struct {
	ID              uuid.UUID
	PublicID        string
	UserID          uuid.UUID
	Title           *string
	Purpose         *string
	Status          string
	UnreadCount     int
	LastMessageAt   *time.Time
	Channel         string
	ChannelThreadID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// Message is one turn in a conversation.
type Message struct {
	ID              uuid.UUID
	PublicID        string
	UserID          uuid.UUID
	ConversationID  uuid.UUID
	Role            string
	Content         string
	ContentFormat   string
	ClientMessageID *string
	Citations       json.RawMessage
	AISessionID     *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
