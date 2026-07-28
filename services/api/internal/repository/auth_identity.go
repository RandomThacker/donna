package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlInsertAuthIdentity = `
INSERT INTO auth_identities (
	id, public_id, user_id, provider, provider_subject, email, email_verified, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING
	id, public_id, user_id, provider, provider_subject, email, email_verified, created_at, updated_at, deleted_at`

	sqlSelectAuthIdentityByProviderSubject = `
SELECT
	id, public_id, user_id, provider, provider_subject, email, email_verified, created_at, updated_at, deleted_at
FROM auth_identities
WHERE provider = $1 AND provider_subject = $2 AND deleted_at IS NULL`
)

// AuthIdentityRepository persists login IdP bindings.
type AuthIdentityRepository interface {
	Create(ctx context.Context, identity entity.AuthIdentity) (entity.AuthIdentity, error)
	GetByProviderSubject(ctx context.Context, provider, subject string) (entity.AuthIdentity, error)
	WithTx(tx pgx.Tx) AuthIdentityRepository
}

type authIdentityRepository struct {
	q Querier
}

// NewAuthIdentityRepository constructs an AuthIdentityRepository.
func NewAuthIdentityRepository(pool *pgxpool.Pool) AuthIdentityRepository {
	return &authIdentityRepository{q: pool}
}

func (r *authIdentityRepository) WithTx(tx pgx.Tx) AuthIdentityRepository {
	return &authIdentityRepository{q: tx}
}

func (r *authIdentityRepository) Create(ctx context.Context, identity entity.AuthIdentity) (entity.AuthIdentity, error) {
	row := r.q.QueryRow(ctx, sqlInsertAuthIdentity,
		identity.ID,
		identity.PublicID,
		identity.UserID,
		identity.Provider,
		identity.ProviderSubject,
		identity.Email,
		identity.EmailVerified,
		identity.CreatedAt,
		identity.UpdatedAt,
	)
	created, err := scanAuthIdentity(row)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.AuthIdentity{}, apperr.ErrConflict
		}
		return entity.AuthIdentity{}, fmt.Errorf("insert auth identity: %w", err)
	}
	return created, nil
}

func (r *authIdentityRepository) GetByProviderSubject(ctx context.Context, provider, subject string) (entity.AuthIdentity, error) {
	identity, err := scanAuthIdentity(r.q.QueryRow(ctx, sqlSelectAuthIdentityByProviderSubject, provider, subject))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AuthIdentity{}, apperr.ErrNotFound
		}
		return entity.AuthIdentity{}, fmt.Errorf("get auth identity: %w", err)
	}
	return identity, nil
}

func scanAuthIdentity(row scannable) (entity.AuthIdentity, error) {
	var a entity.AuthIdentity
	err := row.Scan(
		&a.ID,
		&a.PublicID,
		&a.UserID,
		&a.Provider,
		&a.ProviderSubject,
		&a.Email,
		&a.EmailVerified,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	)
	return a, err
}
