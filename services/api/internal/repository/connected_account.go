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

const connectedAccountColumns = `
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, last_synced_at,
	calendar_list_sync_token, calendar_sync_status, last_failed_sync_at, last_sync_duration_ms,
	last_sync_created_count, last_sync_updated_count, last_sync_deleted_count,
	provider_metadata, created_at, updated_at, deleted_at`

const (
	sqlInsertConnectedAccount = `
INSERT INTO connected_accounts (
	id, public_id, user_id, provider, provider_account_id, display_name, status, scopes,
	credentials_ref, token_expires_at, provider_metadata, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13
)
RETURNING` + connectedAccountColumns

	sqlSelectConnectedAccountByProviderAccount = `
SELECT` + connectedAccountColumns + `
FROM connected_accounts
WHERE provider = $1 AND provider_account_id = $2 AND deleted_at IS NULL`

	sqlSelectConnectedAccountByUserProvider = `
SELECT` + connectedAccountColumns + `
FROM connected_accounts
WHERE user_id = $1 AND provider = $2 AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 1`

	sqlSelectConnectedAccountsByUser = `
SELECT` + connectedAccountColumns + `
FROM connected_accounts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC`

	sqlSelectConnectedAccountByID = `
SELECT` + connectedAccountColumns + `
FROM connected_accounts
WHERE id = $1 AND deleted_at IS NULL`

	sqlUpdateConnectedAccountCredentials = `
UPDATE connected_accounts SET
	credentials_ref = $2,
	token_expires_at = $3,
	scopes = $4,
	status = $5,
	updated_at = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns

	sqlUpdateConnectedAccountProfile = `
UPDATE connected_accounts SET
	display_name = $2,
	provider_metadata = $3::jsonb,
	updated_at = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns

	sqlMarkCalendarSyncRunning = `
UPDATE connected_accounts SET
	calendar_sync_status = $2,
	updated_at = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns

	sqlRecordCalendarSyncResult = `
UPDATE connected_accounts SET
	calendar_sync_status = $2,
	last_synced_at = COALESCE($3, last_synced_at),
	last_failed_sync_at = COALESCE($4, last_failed_sync_at),
	last_sync_duration_ms = $5,
	last_sync_created_count = CASE WHEN $11 THEN $6 ELSE last_sync_created_count END,
	last_sync_updated_count = CASE WHEN $11 THEN $7 ELSE last_sync_updated_count END,
	last_sync_deleted_count = CASE WHEN $11 THEN $8 ELSE last_sync_deleted_count END,
	calendar_list_sync_token = CASE
		WHEN $12 THEN NULL
		WHEN $9::text IS NOT NULL THEN $9
		ELSE calendar_list_sync_token
	END,
	updated_at = $10
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns

	sqlClearCalendarListSyncToken = `
UPDATE connected_accounts SET
	calendar_list_sync_token = NULL,
	updated_at = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns

	sqlSoftDeleteConnectedAccount = `
UPDATE connected_accounts SET
	status = $2,
	deleted_at = $3,
	updated_at = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + connectedAccountColumns
)

// CalendarSyncRecord is persistence input for sync observability.
type CalendarSyncRecord struct {
	Status        string
	SuccessfulAt  *time.Time
	FailedAt      *time.Time
	DurationMs    int
	CreatedCount  int
	UpdatedCount  int
	DeletedCount  int
	ListSyncToken *string
	UpdatedAt     time.Time
	// UpdateCounts is true for successful syncs; failures preserve last success counts.
	UpdateCounts bool
	// ClearListSyncToken forces calendar_list_sync_token to NULL when ListSyncToken is nil.
	ClearListSyncToken bool
}

// ConnectedAccountRepository persists integration accounts.
type ConnectedAccountRepository interface {
	Create(ctx context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.ConnectedAccount, error)
	GetByProviderAccount(ctx context.Context, provider, providerAccountID string) (entity.ConnectedAccount, error)
	GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (entity.ConnectedAccount, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.ConnectedAccount, error)
	UpdateCredentials(ctx context.Context, id uuid.UUID, credentialsRef string, tokenExpiresAt *time.Time, scopes []string, status string, updatedAt time.Time) (entity.ConnectedAccount, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, displayName *string, providerMetadata []byte, updatedAt time.Time) (entity.ConnectedAccount, error)
	MarkCalendarSyncRunning(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) (entity.ConnectedAccount, error)
	RecordCalendarSync(ctx context.Context, id uuid.UUID, record CalendarSyncRecord) (entity.ConnectedAccount, error)
	ClearCalendarListSyncToken(ctx context.Context, id uuid.UUID, updatedAt time.Time) (entity.ConnectedAccount, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedAt time.Time) (entity.ConnectedAccount, error)
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

func (r *connectedAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlSelectConnectedAccountByID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("get connected account by id: %w", err)
	}
	return account, nil
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

func (r *connectedAccountRepository) GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlSelectConnectedAccountByUserProvider, userID, provider))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("get connected account by user: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.ConnectedAccount, error) {
	rows, err := r.q.Query(ctx, sqlSelectConnectedAccountsByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("list connected accounts by user: %w", err)
	}
	defer rows.Close()

	out := make([]entity.ConnectedAccount, 0)
	for rows.Next() {
		account, scanErr := scanConnectedAccount(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan connected account: %w", scanErr)
		}
		out = append(out, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list connected accounts by user: %w", err)
	}
	return out, nil
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

func (r *connectedAccountRepository) UpdateProfile(
	ctx context.Context,
	id uuid.UUID,
	displayName *string,
	providerMetadata []byte,
	updatedAt time.Time,
) (entity.ConnectedAccount, error) {
	meta := providerMetadata
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlUpdateConnectedAccountProfile,
		id,
		displayName,
		meta,
		updatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("update connected account profile: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) MarkCalendarSyncRunning(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlMarkCalendarSyncRunning, id, status, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("mark calendar sync running: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) RecordCalendarSync(ctx context.Context, id uuid.UUID, record CalendarSyncRecord) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlRecordCalendarSyncResult,
		id,
		record.Status,
		record.SuccessfulAt,
		record.FailedAt,
		record.DurationMs,
		record.CreatedCount,
		record.UpdatedCount,
		record.DeletedCount,
		record.ListSyncToken,
		record.UpdatedAt,
		record.UpdateCounts,
		record.ClearListSyncToken,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("record calendar sync: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) ClearCalendarListSyncToken(ctx context.Context, id uuid.UUID, updatedAt time.Time) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlClearCalendarListSyncToken, id, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("clear calendar list sync token: %w", err)
	}
	return account, nil
}

func (r *connectedAccountRepository) SoftDelete(ctx context.Context, id uuid.UUID, deletedAt time.Time) (entity.ConnectedAccount, error) {
	account, err := scanConnectedAccount(r.q.QueryRow(ctx, sqlSoftDeleteConnectedAccount,
		id,
		constant.ConnectedAccountStatusDisconnected,
		deletedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ConnectedAccount{}, apperr.ErrNotFound
		}
		return entity.ConnectedAccount{}, fmt.Errorf("soft delete connected account: %w", err)
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
		&a.CalendarListSyncToken,
		&a.CalendarSyncStatus,
		&a.LastFailedSyncAt,
		&a.LastSyncDurationMs,
		&a.LastSyncCreatedCount,
		&a.LastSyncUpdatedCount,
		&a.LastSyncDeletedCount,
		&a.ProviderMetadata,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.DeletedAt,
	)
	return a, err
}
