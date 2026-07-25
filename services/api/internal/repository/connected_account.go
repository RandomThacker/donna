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

const (
	sqlInsertConnectedAccount = `
INSERT INTO connected_accounts (
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, provider_metadata, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13
)
RETURNING
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, last_synced_at, provider_metadata, created_at, updated_at, deleted_at`

	sqlSelectConnectedAccountByProviderAccount = `
SELECT
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, last_synced_at, provider_metadata, created_at, updated_at, deleted_at
FROM connected_accounts
WHERE provider = $1 AND provider_account_id = $2 AND deleted_at IS NULL`

	sqlUpdateConnectedAccountCredentials = `
UPDATE connected_accounts SET
	credentials_ref = $2,
	token_expires_at = $3,
	scopes = $4,
	status = $5,
	updated_at = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, last_synced_at, provider_metadata, created_at, updated_at, deleted_at`
)

// ConnectedAccountRepository persists integration accounts.
type ConnectedAccountRepository interface {
	Create(ctx context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, error)
	GetByProviderAccount(ctx context.Context, provider, providerAccountID string) (entity.ConnectedAccount, error)
	UpdateCredentials(ctx context.Context, id uuid.UUID, credentialsRef string, tokenExpiresAt *time.Time, scopes []string, status string, updatedAt time.Time) (entity.ConnectedAccount, error)
	WithTx(tx pgx.Tx) ConnectedAccountRepository
}

type connectedAccountRepository struct {
	q Querier
}

// NewConnectedAccountRepository constructs a ConnectedAccountRepository.
func NewConnectedAccountRepository(pool *pgxpool.Pool) ConnectedAccountRepository {
	return &connectedAccountRepository{q: pool}
}

func (r *connectedAccountRepository) WithTx(tx pgx.Tx) ConnectedAccountRepository {
	return &connectedAccountRepository{q: tx}
}

func (r *connectedAccountRepository) Create(ctx context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, error) {
	meta := account.ProviderMetadata
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	scopes := account.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	row := r.q.QueryRow(ctx, sqlInsertConnectedAccount,
		account.ID,
		account.PublicID,
		account.UserID,
		account.Provider,
		account.ProviderAccountID,
		account.DisplayName,
		account.Status,
		scopes,
		account.CredentialsRef,
		account.TokenExpiresAt,
		meta,
		account.CreatedAt,
		account.UpdatedAt,
	)
	created, err := scanConnectedAccount(row)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.ConnectedAccount{}, apperr.ErrConflict
		}
		return entity.ConnectedAccount{}, fmt.Errorf("insert connected account: %w", err)
	}
	return created, nil
}

func (r *connectedAccountRepository) GetByProviderAccount(ctx context.Context, provider, providerAccountID string) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlSelectConnectedAccountByProviderAccount, provider, providerAccountID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("get connected account: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) UpdateCredentials(
	ctx context.Context,
	id uuid.UUID,
	credentialsRef string,
	tokenExpiresAt *time.Time,
	scopes []string,
	status string,
	updatedAt time.Time,
) (entity.ConnectedAccount, error) {
	if scopes == nil {
		scopes = []string{}
	}
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlUpdateConnectedAccountCredentials,
		id,
		credentialsRef,
		tokenExpiresAt,
		scopes,
		status,
		updatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("update connected account credentials: %w", err)
	}
	return account, nil
}

func scanConnectedAccount(row scannable) (entity.ConnectedAccount, error) {
	var a entity.ConnectedAccount
	err := row.Scan(
		&a.ID,
		&a.PublicID,
		&a.UserID,
		&a.Provider,
		&a.ProviderAccountID,
		&a.DisplayName,
		&a.Status,
		&a.Scopes,
		&a.CredentialsRef,
		&a.TokenExpiresAt,
		&a.LastSyncedAt,
		&a.ProviderMetadata,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	)
	return a, err
}
