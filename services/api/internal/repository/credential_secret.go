package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlInsertCredentialSecret = `
INSERT INTO credential_secrets (id, ref, ciphertext, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, ref, ciphertext, created_at, updated_at, deleted_at`

	sqlSelectCredentialSecretByRef = `
SELECT id, ref, ciphertext, created_at, updated_at, deleted_at
FROM credential_secrets
WHERE ref = $1 AND deleted_at IS NULL`

	sqlUpdateCredentialSecret = `
UPDATE credential_secrets SET ciphertext = $2, updated_at = $3
WHERE ref = $1 AND deleted_at IS NULL
RETURNING id, ref, ciphertext, created_at, updated_at, deleted_at`
)

// CredentialSecretRepository persists sealed OAuth tokens.
type CredentialSecretRepository interface {
	Create(ctx context.Context, secret entity.CredentialSecret) (entity.CredentialSecret, error)
	GetByRef(ctx context.Context, ref string) (entity.CredentialSecret, error)
	UpdateCiphertext(ctx context.Context, ref string, ciphertext []byte, updatedAt time.Time) (entity.CredentialSecret, error)
	WithTx(tx pgx.Tx) CredentialSecretRepository
}

type credentialSecretRepository struct {
	q Querier
}

// NewCredentialSecretRepository constructs a CredentialSecretRepository.
func NewCredentialSecretRepository(pool *pgxpool.Pool) CredentialSecretRepository {
	return &credentialSecretRepository{q: pool}
}

func (r *credentialSecretRepository) WithTx(tx pgx.Tx) CredentialSecretRepository {
	return &credentialSecretRepository{q: tx}
}

func (r *credentialSecretRepository) Create(ctx context.Context, secret entity.CredentialSecret) (entity.CredentialSecret, error) {
	row := r.q.QueryRow(ctx, sqlInsertCredentialSecret,
		secret.ID, secret.Ref, secret.Ciphertext, secret.CreatedAt, secret.UpdatedAt,
	)
	created, err := scanCredentialSecret(row)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.CredentialSecret{}, apperr.ErrConflict
		}
		return entity.CredentialSecret{}, fmt.Errorf("insert credential secret: %w", err)
	}
	return created, nil
}

func (r *credentialSecretRepository) GetByRef(ctx context.Context, ref string) (entity.CredentialSecret, error) {
	secret, err := scanCredentialSecret(r.q.QueryRow(ctx, sqlSelectCredentialSecretByRef, ref))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CredentialSecret{}, apperr.ErrNotFound
		}
		return entity.CredentialSecret{}, fmt.Errorf("get credential secret: %w", err)
	}
	return secret, nil
}

func (r *credentialSecretRepository) UpdateCiphertext(ctx context.Context, ref string, ciphertext []byte, updatedAt time.Time) (entity.CredentialSecret, error) {
	secret, err := scanCredentialSecret(r.q.QueryRow(ctx, sqlUpdateCredentialSecret, ref, ciphertext, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CredentialSecret{}, apperr.ErrNotFound
		}
		return entity.CredentialSecret{}, fmt.Errorf("update credential secret: %w", err)
	}
	return secret, nil
}

func scanCredentialSecret(row scannable) (entity.CredentialSecret, error) {
	var s entity.CredentialSecret
	err := row.Scan(&s.ID, &s.Ref, &s.Ciphertext, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	return s, err
}
