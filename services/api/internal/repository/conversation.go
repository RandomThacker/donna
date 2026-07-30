package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const conversationColumns = `
	id, public_id, user_id, title, purpose, status, unread_count,
	last_message_at, channel, channel_thread_id, created_at, updated_at, deleted_at`

const sqlInsertConversation = `
INSERT INTO conversations (
	id, public_id, user_id, title, purpose, status, unread_count,
	last_message_at, channel, channel_thread_id, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING` + conversationColumns

const sqlSelectPrimaryWebConversation = `
SELECT` + conversationColumns + `
FROM conversations
WHERE user_id = $1
  AND channel = $2
  AND status = $3
  AND purpose = $4
  AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 1`

const sqlTouchConversationLastMessage = `
UPDATE conversations
SET last_message_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + conversationColumns

// ConversationRepository persists chat threads.
type ConversationRepository interface {
	Create(ctx context.Context, c entity.Conversation) (entity.Conversation, error)
	FindPrimaryWeb(ctx context.Context, userID uuid.UUID) (entity.Conversation, error)
	TouchLastMessage(ctx context.Context, id uuid.UUID, at time.Time) (entity.Conversation, error)
	WithTx(tx pgx.Tx) ConversationRepository
}

type conversationRepository struct {
	q Querier
}

// NewConversationRepository constructs a ConversationRepository.
func NewConversationRepository(pool *pgxpool.Pool) ConversationRepository {
	return &conversationRepository{q: pool}
}

func (r *conversationRepository) WithTx(tx pgx.Tx) ConversationRepository {
	return &conversationRepository{q: tx}
}

func (r *conversationRepository) Create(ctx context.Context, c entity.Conversation) (entity.Conversation, error) {
	row := r.q.QueryRow(ctx, sqlInsertConversation,
		c.ID, c.PublicID, c.UserID, c.Title, c.Purpose, c.Status, c.UnreadCount,
		c.LastMessageAt, c.Channel, c.ChannelThreadID, c.CreatedAt, c.UpdatedAt,
	)
	return scanConversation(row)
}

func (r *conversationRepository) FindPrimaryWeb(ctx context.Context, userID uuid.UUID) (entity.Conversation, error) {
	purpose := constant.ConversationPurposeGeneral
	return scanConversation(r.q.QueryRow(ctx, sqlSelectPrimaryWebConversation,
		userID,
		constant.ConversationChannelWeb,
		constant.ConversationStatusActive,
		purpose,
	))
}

func (r *conversationRepository) TouchLastMessage(ctx context.Context, id uuid.UUID, at time.Time) (entity.Conversation, error) {
	return scanConversation(r.q.QueryRow(ctx, sqlTouchConversationLastMessage, id, at))
}

func scanConversation(row pgx.Row) (entity.Conversation, error) {
	var c entity.Conversation
	err := row.Scan(
		&c.ID, &c.PublicID, &c.UserID, &c.Title, &c.Purpose, &c.Status, &c.UnreadCount,
		&c.LastMessageAt, &c.Channel, &c.ChannelThreadID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Conversation{}, apperr.ErrNotFound
		}
		return entity.Conversation{}, fmt.Errorf("scan conversation: %w", err)
	}
	return c, nil
}
