package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pushSubscriptionColumns = `
	id, public_id, user_id, endpoint, p256dh, auth, user_agent,
	created_at, updated_at, deleted_at`

const (
	sqlInsertPushSubscription = `
INSERT INTO push_subscriptions (
	id, public_id, user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING` + pushSubscriptionColumns

	sqlSelectPushSubscriptionByUserEndpoint = `
SELECT` + pushSubscriptionColumns + `
FROM push_subscriptions
WHERE user_id = $1 AND endpoint = $2 AND deleted_at IS NULL
LIMIT 1`

	sqlUpdatePushSubscriptionKeys = `
UPDATE push_subscriptions SET
	p256dh = $3,
	auth = $4,
	user_agent = $5,
	updated_at = $6
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING` + pushSubscriptionColumns

	sqlListPushSubscriptionsByUser = `
SELECT` + pushSubscriptionColumns + `
FROM push_subscriptions
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC`

	sqlSoftDeletePushSubscriptionByEndpoint = `
UPDATE push_subscriptions SET deleted_at = $3, updated_at = $3
WHERE user_id = $1 AND endpoint = $2 AND deleted_at IS NULL`
)

// PushSubscriptionRepository persists browser push endpoints.
type PushSubscriptionRepository interface {
	Upsert(ctx context.Context, sub entity.PushSubscription) (entity.PushSubscription, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.PushSubscription, error)
	SoftDeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string, deletedAt time.Time) error
	WithTx(tx pgx.Tx) PushSubscriptionRepository
}

type pushSubscriptionRepository struct {
	q Querier
}

// NewPushSubscriptionRepository constructs a PushSubscriptionRepository.
func NewPushSubscriptionRepository(pool *pgxpool.Pool) PushSubscriptionRepository {
	return &pushSubscriptionRepository{q: pool}
}

func (r *pushSubscriptionRepository) WithTx(tx pgx.Tx) PushSubscriptionRepository {
	return &pushSubscriptionRepository{q: tx}
}

func (r *pushSubscriptionRepository) Upsert(ctx context.Context, sub entity.PushSubscription) (entity.PushSubscription, error) {
	existing, err := scanPushSubscription(r.q.QueryRow(ctx, sqlSelectPushSubscriptionByUserEndpoint, sub.UserID, sub.Endpoint))
	if err == nil {
		return scanPushSubscription(r.q.QueryRow(ctx, sqlUpdatePushSubscriptionKeys,
			existing.ID, sub.UserID, sub.P256dh, sub.Auth, sub.UserAgent, sub.UpdatedAt,
		))
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		return entity.PushSubscription{}, err
	}
	return scanPushSubscription(r.q.QueryRow(ctx, sqlInsertPushSubscription,
		sub.ID, sub.PublicID, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth,
		sub.UserAgent, sub.CreatedAt, sub.UpdatedAt,
	))
}

func (r *pushSubscriptionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.PushSubscription, error) {
	rows, err := r.q.Query(ctx, sqlListPushSubscriptionsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.PushSubscription, 0)
	for rows.Next() {
		sub, err := scanPushSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (r *pushSubscriptionRepository) SoftDeleteByEndpoint(
	ctx context.Context,
	userID uuid.UUID,
	endpoint string,
	deletedAt time.Time,
) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeletePushSubscriptionByEndpoint, userID, endpoint, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func scanPushSubscription(row pgx.Row) (entity.PushSubscription, error) {
	var s entity.PushSubscription
	err := row.Scan(
		&s.ID, &s.PublicID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.UserAgent,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.PushSubscription{}, apperr.ErrNotFound
		}
		return entity.PushSubscription{}, fmt.Errorf("scan push subscription: %w", err)
	}
	return s, nil
}
