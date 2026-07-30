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
	UnreadCount    int                   `json:"unread_count"`
	Messages       []ChatMessageResponse `json:"messages"`
}

// ChatSummaryResponse is GET /chat/summary — list preview + unread.
type ChatSummaryResponse struct {
	ConversationID       string     `json:"conversation_id"`
	ConversationPublicID string     `json:"conversation_public_id"`
	UnreadCount          int        `json:"unread_count"`
	Preview              string     `json:"preview"`
	LastMessageAt        *time.Time `json:"last_message_at,omitempty"`
	LastMessageRole      string     `json:"last_message_role,omitempty"`
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
		UnreadCount:    conv.UnreadCount,
		Messages:       ChatMessagesFromEntities(messages),
	}
}

// ChatSummaryFromEntities builds the list summary response.
func ChatSummaryFromEntities(conv entity.Conversation, last *entity.Message) ChatSummaryResponse {
	out := ChatSummaryResponse{
		ConversationID:       conv.ID.String(),
		ConversationPublicID: conv.PublicID,
		UnreadCount:          conv.UnreadCount,
		Preview:              "",
	}
	if last != nil {
		out.Preview = last.Content
		out.LastMessageRole = last.Role
		at := last.CreatedAt
		out.LastMessageAt = &at
	} else if conv.LastMessageAt != nil {
		out.LastMessageAt = conv.LastMessageAt
	}
	return out
}
