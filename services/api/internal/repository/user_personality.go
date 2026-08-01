package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const userPersonalityColumns = `
	id, user_id, personality_id, display_name, nickname,
	emoji_level, humor_level, greeting_style, encouragement_level, response_style,
	created_at, updated_at`

const (
	sqlSelectUserPersonality = `
SELECT` + userPersonalityColumns + `
FROM user_personality
WHERE user_id = $1`

	sqlUpsertUserPersonality = `
INSERT INTO user_personality (
	id, user_id, personality_id, display_name, nickname,
	emoji_level, humor_level, greeting_style, encouragement_level, response_style,
	created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
)
ON CONFLICT (user_id) DO UPDATE SET
	personality_id = EXCLUDED.personality_id,
	display_name = EXCLUDED.display_name,
	nickname = EXCLUDED.nickname,
	emoji_level = EXCLUDED.emoji_level,
	humor_level = EXCLUDED.humor_level,
	greeting_style = EXCLUDED.greeting_style,
	encouragement_level = EXCLUDED.encouragement_level,
	response_style = EXCLUDED.response_style,
	updated_at = EXCLUDED.updated_at
RETURNING` + userPersonalityColumns
)

// UserPersonalityRepository persists per-user personality preferences.
type UserPersonalityRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (entity.UserPersonality, error)
	Upsert(ctx context.Context, row entity.UserPersonality) (entity.UserPersonality, error)
	WithTx(tx pgx.Tx) UserPersonalityRepository
}

type userPersonalityRepository struct {
	q Querier
}

// NewUserPersonalityRepository constructs a UserPersonalityRepository.
func NewUserPersonalityRepository(pool *pgxpool.Pool) UserPersonalityRepository {
	return &userPersonalityRepository{q: pool}
}

func (r *userPersonalityRepository) WithTx(tx pgx.Tx) UserPersonalityRepository {
	return &userPersonalityRepository{q: tx}
}

func (r *userPersonalityRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (entity.UserPersonality, error) {
	return scanUserPersonality(r.q.QueryRow(ctx, sqlSelectUserPersonality, userID))
}

func (r *userPersonalityRepository) Upsert(ctx context.Context, row entity.UserPersonality) (entity.UserPersonality, error) {
	return scanUserPersonality(r.q.QueryRow(ctx, sqlUpsertUserPersonality,
		row.ID, row.UserID, row.PersonalityID, row.DisplayName, row.Nickname,
		row.EmojiLevel, row.HumorLevel, row.GreetingStyle, row.EncouragementLevel, row.ResponseStyle,
		row.CreatedAt, row.UpdatedAt,
	))
}

func scanUserPersonality(row pgx.Row) (entity.UserPersonality, error) {
	var out entity.UserPersonality
	err := row.Scan(
		&out.ID, &out.UserID, &out.PersonalityID, &out.DisplayName, &out.Nickname,
		&out.EmojiLevel, &out.HumorLevel, &out.GreetingStyle, &out.EncouragementLevel, &out.ResponseStyle,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.UserPersonality{}, apperr.ErrNotFound
		}
		return entity.UserPersonality{}, fmt.Errorf("scan user_personality: %w", err)
	}
	return out, nil
}