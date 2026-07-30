package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// ChatMessageResponse is one persisted chat turn.
type ChatMessageResponse struct {
	ID        string    `json:"id"`
	PublicID  string    `json:"public_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatHistoryResponse is GET /chat/messages.
type ChatHistoryResponse struct {
	ConversationID string                `json:"conversation_id"`
	PublicID       string                `json:"conversation_public_id"`
	Messages       []ChatMessageResponse `json:"messages"`
}

// ChatCommandResponse is POST /chat/command with optional persisted ids.
type ChatCommandResponse struct {
	Reply                string `json:"reply"`
	Intent               string `json:"intent"`
	ConversationPublicID string `json:"conversation_public_id,omitempty"`
	UserMessagePublicID  string `json:"user_message_public_id,omitempty"`
	ReplyMessagePublicID string `json:"reply_message_public_id,omitempty"`
}

// ChatMessageFromEntity maps a message entity to the API shape.
func ChatMessageFromEntity(m entity.Message) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        m.ID.String(),
		PublicID:  m.PublicID,
		Role:      m.Role,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

// ChatMessagesFromEntities maps a slice of messages.
func ChatMessagesFromEntities(messages []entity.Message) []ChatMessageResponse {
	out := make([]ChatMessageResponse, 0, len(messages))
	for _, m := range messages {
		out = append(out, ChatMessageFromEntity(m))
	}
	return out
}

// ChatHistoryFromEntities builds the history response.
func ChatHistoryFromEntities(conv entity.Conversation, messages []entity.Message) ChatHistoryResponse {
	return ChatHistoryResponse{
		ConversationID: conv.ID.String(),
		PublicID:       conv.PublicID,
		Messages:       ChatMessagesFromEntities(messages),
	}
}
