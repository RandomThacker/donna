package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ConversationService owns primary web chat history.
type ConversationService struct {
	conversations repository.ConversationRepository
	messages      repository.MessageRepository
	tx            txRunner
	now           func() time.Time
}

type txRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// NewConversationService constructs a ConversationService.
func NewConversationService(
	conversations repository.ConversationRepository,
	messages repository.MessageRepository,
	tx txRunner,
) *ConversationService {
	return &ConversationService{
		conversations: conversations,
		messages:      messages,
		tx:            tx,
		now:           time.Now,
	}
}

// ChatHistory is the primary web thread and its messages.
type ChatHistory struct {
	Conversation entity.Conversation
	Messages     []entity.Message
}

// ChatSummary is list-row metadata for the primary web thread.
type ChatSummary struct {
	Conversation entity.Conversation
	LastMessage  *entity.Message
}

// AppendTurnResult is the persisted user + assistant pair.
type AppendTurnResult struct {
	Conversation     entity.Conversation
	UserMessage      entity.Message
	AssistantMessage entity.Message
}

// GetPrimaryHistory returns the primary web conversation messages (creates thread if missing).
// Loading history marks the thread as read.
func (s *ConversationService) GetPrimaryHistory(ctx context.Context, userID uuid.UUID) (ChatHistory, error) {
	if userID == uuid.Nil {
		return ChatHistory{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	conv, err := s.getOrCreatePrimary(ctx, s.conversations, userID)
	if err != nil {
		return ChatHistory{}, err
	}
	msgs, err := s.messages.ListByConversation(ctx, conv.ID, constant.ChatHistoryDefaultLimit)
	if err != nil {
		return ChatHistory{}, err
	}
	if conv.UnreadCount > 0 {
		cleared, clearErr := s.conversations.ClearUnread(ctx, conv.ID, s.now().UTC())
		if clearErr == nil {
			conv = cleared
		}
	}
	return ChatHistory{Conversation: conv, Messages: msgs}, nil
}

// GetPrimarySummary returns preview + unread for the messages list / FAB badge.
func (s *ConversationService) GetPrimarySummary(ctx context.Context, userID uuid.UUID) (ChatSummary, error) {
	if userID == uuid.Nil {
		return ChatSummary{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	conv, err := s.getOrCreatePrimary(ctx, s.conversations, userID)
	if err != nil {
		return ChatSummary{}, err
	}
	last, err := s.messages.LatestByConversation(ctx, conv.ID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return ChatSummary{Conversation: conv}, nil
		}
		return ChatSummary{}, err
	}
	return ChatSummary{Conversation: conv, LastMessage: &last}, nil
}

// PostAssistantNotice appends a Donna-only message (notifications, system prompts).
// clientMessageID makes the write idempotent when set.
// created is false when the notice was already present (unique client_message_id).
func (s *ConversationService) PostAssistantNotice(
	ctx context.Context,
	userID uuid.UUID,
	content string,
	clientMessageID string,
) (entity.Message, bool, error) {
	if userID == uuid.Nil {
		return entity.Message{}, false, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return entity.Message{}, false, fmt.Errorf("%w: content is required", apperr.ErrValidation)
	}

	var out entity.Message
	var created bool
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		convs := s.conversations.WithTx(tx)
		msgs := s.messages.WithTx(tx)

		conv, err := s.getOrCreatePrimary(ctx, convs, userID)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		var clientID *string
		if trimmed := strings.TrimSpace(clientMessageID); trimmed != "" {
			clientID = &trimmed
			exists, err := msgs.ExistsByClientMessageID(ctx, conv.ID, trimmed)
			if err != nil {
				return err
			}
			if exists {
				created = false
				out = entity.Message{}
				return nil
			}
		}
		msg, err := s.createMessage(ctx, msgs, userID, conv.ID, constant.MessageRoleAssistant, content, clientID, now)
		if err != nil {
			return err
		}
		if _, err := convs.BumpUnread(ctx, conv.ID, 1, now); err != nil {
			return err
		}
		out = msg
		created = true
		return nil
	})
	if err != nil {
		return entity.Message{}, false, err
	}
	return out, created, nil
}

// AppendTurn persists a user message and Donna reply on the primary web thread.
func (s *ConversationService) AppendTurn(
	ctx context.Context,
	userID uuid.UUID,
	userText string,
	assistantText string,
	clientMessageID *string,
) (AppendTurnResult, error) {
	if userID == uuid.Nil {
		return AppendTurnResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	userText = strings.TrimSpace(userText)
	assistantText = strings.TrimSpace(assistantText)
	if userText == "" {
		return AppendTurnResult{}, fmt.Errorf("%w: message is required", apperr.ErrValidation)
	}
	if assistantText == "" {
		assistantText = "…"
	}

	var out AppendTurnResult
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		convs := s.conversations.WithTx(tx)
		msgs := s.messages.WithTx(tx)

		conv, err := s.getOrCreatePrimary(ctx, convs, userID)
		if err != nil {
			return err
		}

		now := s.now().UTC()
		userMsg, err := s.createMessage(ctx, msgs, userID, conv.ID, constant.MessageRoleUser, userText, clientMessageID, now)
		if err != nil {
			return err
		}
		assistantMsg, err := s.createMessage(ctx, msgs, userID, conv.ID, constant.MessageRoleAssistant, assistantText, nil, now.Add(time.Millisecond))
		if err != nil {
			return err
		}
		touched, err := convs.TouchLastMessage(ctx, conv.ID, now.Add(time.Millisecond))
		if err != nil {
			return err
		}
		out = AppendTurnResult{
			Conversation:     touched,
			UserMessage:      userMsg,
			AssistantMessage: assistantMsg,
		}
		return nil
	})
	if err != nil {
		return AppendTurnResult{}, err
	}
	return out, nil
}

func (s *ConversationService) getOrCreatePrimary(
	ctx context.Context,
	convs repository.ConversationRepository,
	userID uuid.UUID,
) (entity.Conversation, error) {
	existing, err := convs.FindPrimaryWeb(ctx, userID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		return entity.Conversation{}, err
	}

	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.Conversation{}, err
	}
	now := s.now().UTC()
	purpose := constant.ConversationPurposeGeneral
	created, err := convs.Create(ctx, entity.Conversation{
		ID:        id,
		PublicID:  idgen.PublicID(constant.PublicIDPrefixConversation, id),
		UserID:    userID,
		Purpose:   &purpose,
		Status:    constant.ConversationStatusActive,
		Channel:   constant.ConversationChannelWeb,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		// Concurrent create: re-select the winner.
		existing, findErr := convs.FindPrimaryWeb(ctx, userID)
		if findErr == nil {
			return existing, nil
		}
		return entity.Conversation{}, err
	}
	return created, nil
}

func (s *ConversationService) createMessage(
	ctx context.Context,
	msgs repository.MessageRepository,
	userID, conversationID uuid.UUID,
	role, content string,
	clientMessageID *string,
	at time.Time,
) (entity.Message, error) {
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.Message{}, err
	}
	var clientID *string
	if clientMessageID != nil {
		trimmed := strings.TrimSpace(*clientMessageID)
		if trimmed != "" {
			clientID = &trimmed
		}
	}
	return msgs.Create(ctx, entity.Message{
		ID:              id,
		PublicID:        idgen.PublicID(constant.PublicIDPrefixMessage, id),
		UserID:          userID,
		ConversationID:  conversationID,
		Role:            role,
		Content:         content,
		ContentFormat:   constant.MessageContentFormatPlain,
		ClientMessageID: clientID,
		Citations:       json.RawMessage(`{}`),
		CreatedAt:       at,
		UpdatedAt:       at,
	})
}
