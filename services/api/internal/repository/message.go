package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const messageColumns = `
	id, public_id, user_id, conversation_id, role, content, content_format,
	client_message_id, citations, ai_session_id, created_at, updated_at, deleted_at`

const sqlInsertMessage = `
INSERT INTO messages (
	id, public_id, user_id, conversation_id, role, content, content_format,
	client_message_id, citations, ai_session_id, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING` + messageColumns

const sqlListMessagesByConversation = `
SELECT` + messageColumns + `
FROM messages
WHERE conversation_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $2`

// MessageRepository persists conversation turns.
type MessageRepository interface {
	Create(ctx context.Context, m entity.Message) (entity.Message, error)
	ListByConversation(ctx context.Context, conversationID uuid.UUID, limit int) ([]entity.Message, error)
	WithTx(tx pgx.Tx) MessageRepository
}

type messageRepository struct {
	q Querier
}

// NewMessageRepository constructs a MessageRepository.
func NewMessageRepository(pool *pgxpool.Pool) MessageRepository {
	return &messageRepository{q: pool}
}

func (r *messageRepository) WithTx(tx pgx.Tx) MessageRepository {
	return &messageRepository{q: tx}
}

func (r *messageRepository) Create(ctx context.Context, m entity.Message) (entity.Message, error) {
	citations := m.Citations
	if len(citations) == 0 {
		citations = json.RawMessage(`{}`)
	}
	row := r.q.QueryRow(ctx, sqlInsertMessage,
		m.ID, m.PublicID, m.UserID, m.ConversationID, m.Role, m.Content, m.ContentFormat,
		m.ClientMessageID, citations, m.AISessionID, m.CreatedAt, m.UpdatedAt,
	)
	return scanMessage(row)
}

func (r *messageRepository) ListByConversation(ctx context.Context, conversationID uuid.UUID, limit int) ([]entity.Message, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.q.Query(ctx, sqlListMessagesByConversation, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Message, 0)
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func scanMessage(row pgx.Row) (entity.Message, error) {
	var m entity.Message
	err := row.Scan(
		&m.ID, &m.PublicID, &m.UserID, &m.ConversationID, &m.Role, &m.Content, &m.ContentFormat,
		&m.ClientMessageID, &m.Citations, &m.AISessionID, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Message{}, apperr.ErrNotFound
		}
		return entity.Message{}, fmt.Errorf("scan message: %w", err)
	}
	return m, nil
}