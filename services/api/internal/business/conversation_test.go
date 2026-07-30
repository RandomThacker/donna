package business

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type memConversationRepo struct {
	byID   map[uuid.UUID]entity.Conversation
	byUser map[uuid.UUID][]uuid.UUID
}

func newMemConversationRepo() *memConversationRepo {
	return &memConversationRepo{
		byID:   map[uuid.UUID]entity.Conversation{},
		byUser: map[uuid.UUID][]uuid.UUID{},
	}
}

func (m *memConversationRepo) Create(_ context.Context, c entity.Conversation) (entity.Conversation, error) {
	m.byID[c.ID] = c
	m.byUser[c.UserID] = append(m.byUser[c.UserID], c.ID)
	return c, nil
}

func (m *memConversationRepo) FindPrimaryWeb(_ context.Context, userID uuid.UUID) (entity.Conversation, error) {
	ids := m.byUser[userID]
	var oldest *entity.Conversation
	for _, id := range ids {
		c := m.byID[id]
		if c.DeletedAt != nil || c.Channel != constant.ConversationChannelWeb || c.Status != constant.ConversationStatusActive {
			continue
		}
		if c.Purpose == nil || *c.Purpose != constant.ConversationPurposeGeneral {
			continue
		}
		if oldest == nil || c.CreatedAt.Before(oldest.CreatedAt) {
			cp := c
			oldest = &cp
		}
	}
	if oldest == nil {
		return entity.Conversation{}, apperr.ErrNotFound
	}
	return *oldest, nil
}

func (m *memConversationRepo) TouchLastMessage(_ context.Context, id uuid.UUID, at time.Time) (entity.Conversation, error) {
	c, ok := m.byID[id]
	if !ok || c.DeletedAt != nil {
		return entity.Conversation{}, apperr.ErrNotFound
	}
	c.LastMessageAt = &at
	c.UpdatedAt = at
	m.byID[id] = c
	return c, nil
}

func (m *memConversationRepo) WithTx(pgx.Tx) repository.ConversationRepository { return m }

type memMessageRepo struct {
	byConv map[uuid.UUID][]entity.Message
}

func newMemMessageRepo() *memMessageRepo {
	return &memMessageRepo{byConv: map[uuid.UUID][]entity.Message{}}
}

func (m *memMessageRepo) Create(_ context.Context, msg entity.Message) (entity.Message, error) {
	m.byConv[msg.ConversationID] = append(m.byConv[msg.ConversationID], msg)
	return msg, nil
}

func (m *memMessageRepo) ListByConversation(_ context.Context, conversationID uuid.UUID, limit int) ([]entity.Message, error) {
	all := m.byConv[conversationID]
	if limit > 0 && len(all) > limit {
		return append([]entity.Message(nil), all[:limit]...), nil
	}
	return append([]entity.Message(nil), all...), nil
}

func (m *memMessageRepo) WithTx(pgx.Tx) repository.MessageRepository { return m }

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

func TestConversationAppendAndHistory(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000301")
	convs := newMemConversationRepo()
	msgs := newMemMessageRepo()
	svc := NewConversationService(convs, msgs, memTx{})
	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	_, err := svc.AppendTurn(context.Background(), userID, "Add task Ship chat", "Done — added “Ship chat”.", nil)
	if err != nil {
		t.Fatal(err)
	}
	history, err := svc.GetPrimaryHistory(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history.Messages))
	}
	if history.Messages[0].Role != constant.MessageRoleUser || history.Messages[0].Content != "Add task Ship chat" {
		t.Fatalf("user message = %+v", history.Messages[0])
	}
	if history.Messages[1].Role != constant.MessageRoleAssistant {
		t.Fatalf("assistant role = %s", history.Messages[1].Role)
	}
	if history.Conversation.PublicID == "" {
		t.Fatal("expected conversation public id")
	}
}

func TestConversationRejectsEmptyMessage(t *testing.T) {
	t.Parallel()
	svc := NewConversationService(newMemConversationRepo(), newMemMessageRepo(), memTx{})
	_, err := svc.AppendTurn(context.Background(), uuid.MustParse("018f0000-0000-7000-8000-000000000302"), "  ", "x", nil)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}
