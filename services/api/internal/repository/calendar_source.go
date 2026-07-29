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
	calendarSourceColumns = `
	id, public_id, user_id, connected_account_id, provider_calendar_id, name, color,
	is_primary_on_provider, is_writable, access_role, sync_enabled, sync_cursor, last_synced_at, timezone,
	provider_metadata, created_at, updated_at, deleted_at`

	calendarSourceColumnsAliased = `
	s.id, s.public_id, s.user_id, s.connected_account_id, s.provider_calendar_id, s.name, s.color,
	s.is_primary_on_provider, s.is_writable, s.access_role, s.sync_enabled, s.sync_cursor, s.last_synced_at, s.timezone,
	s.provider_metadata, s.created_at, s.updated_at, s.deleted_at`

	sqlInsertCalendarSource = `
INSERT INTO calendar_sources (
	id, public_id, user_id, connected_account_id, provider_calendar_id, name, color,
	is_primary_on_provider, is_writable, access_role, sync_enabled, sync_cursor, last_synced_at, timezone,
	provider_metadata, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::jsonb,$16,$17
)
RETURNING` + calendarSourceColumns

	sqlSelectCalendarSourceByAccountProvider = `
SELECT` + calendarSourceColumns + `
FROM calendar_sources
WHERE connected_account_id = $1 AND provider_calendar_id = $2
ORDER BY deleted_at NULLS FIRST
LIMIT 1`

	sqlSelectCalendarSourcesByUser = `
SELECT` + calendarSourceColumnsAliased + `
FROM calendar_sources s
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE s.user_id = $1
  AND s.deleted_at IS NULL
  AND ca.deleted_at IS NULL
ORDER BY s.is_primary_on_provider DESC, s.name ASC`

	sqlSelectCalendarSourcesByAccount = `
SELECT` + calendarSourceColumns + `
FROM calendar_sources
WHERE connected_account_id = $1 AND deleted_at IS NULL`

	sqlUpdateCalendarSourceFromSync = `
UPDATE calendar_sources SET
	name = $2,
	color = $3,
	is_primary_on_provider = $4,
	is_writable = $5,
	access_role = $6,
	sync_cursor = $7,
	last_synced_at = $8,
	timezone = $9,
	provider_metadata = $10::jsonb,
	deleted_at = NULL,
	updated_at = $11
WHERE id = $1
RETURNING` + calendarSourceColumns

	sqlSoftDeleteCalendarSourcesMissing = `
UPDATE calendar_sources SET
	deleted_at = $3,
	updated_at = $3
WHERE connected_account_id = $1
  AND deleted_at IS NULL
  AND NOT (provider_calendar_id = ANY($2::text[]))`

	sqlSoftDeleteCalendarSourcesByProviderIDs = `
UPDATE calendar_sources SET
	deleted_at = $3,
	updated_at = $3
WHERE connected_account_id = $1
  AND deleted_at IS NULL
  AND provider_calendar_id = ANY($2::text[])`

	sqlDeleteCalendarSourcesByAccount = `
DELETE FROM calendar_sources
WHERE connected_account_id = $1`

	sqlDeleteOrphanCalendarSourcesForUser = `
DELETE FROM calendar_sources s
USING connected_accounts ca
WHERE s.connected_account_id = ca.id
  AND s.user_id = $1
  AND ca.deleted_at IS NOT NULL`

	sqlUpdateCalendarSourceEventSyncState = `
UPDATE calendar_sources SET
	sync_cursor = $2,
	last_synced_at = $3,
	updated_at = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + calendarSourceColumns

	sqlClearCalendarSourceEventSyncCursor = `
UPDATE calendar_sources SET
	sync_cursor = NULL,
	updated_at = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + calendarSourceColumns

	sqlUpdateCalendarSourcesSyncEnabledByAccount = `
UPDATE calendar_sources SET
	sync_enabled = $2,
	updated_at = $3
WHERE connected_account_id = $1 AND deleted_at IS NULL`
)

// CalendarSourceRepository persists calendar feeds.
type CalendarSourceRepository interface {
	Create(ctx context.Context, source entity.CalendarSource) (entity.CalendarSource, error)
	GetByAccountAndProviderCalendar(ctx context.Context, accountID uuid.UUID, providerCalendarID string) (entity.CalendarSource, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.CalendarSource, error)
	ListByConnectedAccountID(ctx context.Context, accountID uuid.UUID) ([]entity.CalendarSource, error)
	UpdateFromSync(ctx context.Context, source entity.CalendarSource) (entity.CalendarSource, error)
	SoftDeleteMissing(ctx context.Context, accountID uuid.UUID, keepProviderIDs []string, deletedAt time.Time) (int64, error)
	SoftDeleteByProviderIDs(ctx context.Context, accountID uuid.UUID, providerIDs []string, deletedAt time.Time) (int64, error)
	DeleteByConnectedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
	DeleteOrphansForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateEventSyncState(ctx context.Context, id uuid.UUID, syncCursor *string, lastSyncedAt, updatedAt time.Time) (entity.CalendarSource, error)
	ClearEventSyncCursor(ctx context.Context, id uuid.UUID, updatedAt time.Time) (entity.CalendarSource, error)
	UpdateSyncEnabledByAccount(ctx context.Context, accountID uuid.UUID, syncEnabled bool, updatedAt time.Time) (int64, error)
	WithTx(tx pgx.Tx) CalendarSourceRepository
}

type calendarSourceRepository struct {
	q Querier
}

// NewCalendarSourceRepository constructs a CalendarSourceRepository.
func NewCalendarSourceRepository(pool *pgxpool.Pool) CalendarSourceRepository {
	return &calendarSourceRepository{q: pool}
}

func (r *calendarSourceRepository) WithTx(tx pgx.Tx) CalendarSourceRepository {
	return &calendarSourceRepository{q: tx}
}

func (r *calendarSourceRepository) Create(ctx context.Context, source entity.CalendarSource) (entity.CalendarSource, error) {
	meta := source.ProviderMetadata
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	row := r.q.QueryRow(ctx, sqlInsertCalendarSource,
		source.ID,
		source.PublicID,
		source.UserID,
		source.ConnectedAccountID,
		source.ProviderCalendarID,
		source.Name,
		source.Color,
		source.IsPrimaryOnProvider,
		source.IsWritable,
		source.AccessRole,
		source.SyncEnabled,
		source.SyncCursor,
		source.LastSyncedAt,
		source.Timezone,
		meta,
		source.CreatedAt,
		source.UpdatedAt,
	)
	created, err := scanCalendarSource(row)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.CalendarSource{}, apperr.ErrConflict
		}
		return entity.CalendarSource{}, fmt.Errorf("insert calendar source: %w", err)
	}
	return created, nil
}

func (r *calendarSourceRepository) GetByAccountAndProviderCalendar(
	ctx context.Context,
	accountID uuid.UUID,
	providerCalendarID string,
) (entity.CalendarSource, error) {
	source, err := scanCalendarSource(r.q.QueryRow(ctx, sqlSelectCalendarSourceByAccountProvider, accountID, providerCalendarID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarSource{}, apperr.ErrNotFound
		}
		return entity.CalendarSource{}, fmt.Errorf("get calendar source: %w", err)
	}
	return source, nil
}

func (r *calendarSourceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.CalendarSource, error) {
	rows, err := r.q.Query(ctx, sqlSelectCalendarSourcesByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("list calendar sources by user: %w", err)
	}
	defer rows.Close()
	return collectCalendarSources(rows)
}

func (r *calendarSourceRepository) ListByConnectedAccountID(ctx context.Context, accountID uuid.UUID) ([]entity.CalendarSource, error) {
	rows, err := r.q.Query(ctx, sqlSelectCalendarSourcesByAccount, accountID)
	if err != nil {
		return nil, fmt.Errorf("list calendar sources by account: %w", err)
	}
	defer rows.Close()
	return collectCalendarSources(rows)
}

func (r *calendarSourceRepository) UpdateFromSync(ctx context.Context, source entity.CalendarSource) (entity.CalendarSource, error) {
	meta := source.ProviderMetadata
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	updated, err := scanCalendarSource(r.q.QueryRow(ctx, sqlUpdateCalendarSourceFromSync,
		source.ID,
		source.Name,
		source.Color,
		source.IsPrimaryOnProvider,
		source.IsWritable,
		source.AccessRole,
		source.SyncCursor,
		source.LastSyncedAt,
		source.Timezone,
		meta,
		source.UpdatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarSource{}, apperr.ErrNotFound
		}
		return entity.CalendarSource{}, fmt.Errorf("update calendar source: %w", err)
	}
	return updated, nil
}

func (r *calendarSourceRepository) SoftDeleteMissing(
	ctx context.Context,
	accountID uuid.UUID,
	keepProviderIDs []string,
	deletedAt time.Time,
) (int64, error) {
	if keepProviderIDs == nil {
		keepProviderIDs = []string{}
	}
	tag, err := r.q.Exec(ctx, sqlSoftDeleteCalendarSourcesMissing, accountID, keepProviderIDs, deletedAt)
	if err != nil {
		return 0, fmt.Errorf("soft-delete missing calendar sources: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarSourceRepository) SoftDeleteByProviderIDs(
	ctx context.Context,
	accountID uuid.UUID,
	providerIDs []string,
	deletedAt time.Time,
) (int64, error) {
	if len(providerIDs) == 0 {
		return 0, nil
	}
	tag, err := r.q.Exec(ctx, sqlSoftDeleteCalendarSourcesByProviderIDs, accountID, providerIDs, deletedAt)
	if err != nil {
		return 0, fmt.Errorf("soft-delete calendar sources by provider ids: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarSourceRepository) DeleteByConnectedAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) (int64, error) {
	tag, err := r.q.Exec(ctx, sqlDeleteCalendarSourcesByAccount, accountID)
	if err != nil {
		return 0, fmt.Errorf("delete calendar sources by account: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarSourceRepository) DeleteOrphansForUser(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	tag, err := r.q.Exec(ctx, sqlDeleteOrphanCalendarSourcesForUser, userID)
	if err != nil {
		return 0, fmt.Errorf("delete orphan calendar sources: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarSourceRepository) UpdateEventSyncState(
	ctx context.Context,
	id uuid.UUID,
	syncCursor *string,
	lastSyncedAt, updatedAt time.Time,
) (entity.CalendarSource, error) {
	source, err := scanCalendarSource(r.q.QueryRow(ctx, sqlUpdateCalendarSourceEventSyncState, id, syncCursor, lastSyncedAt, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarSource{}, apperr.ErrNotFound
		}
		return entity.CalendarSource{}, fmt.Errorf("update calendar source event sync state: %w", err)
	}
	return source, nil
}

func (r *calendarSourceRepository) ClearEventSyncCursor(ctx context.Context, id uuid.UUID, updatedAt time.Time) (entity.CalendarSource, error) {
	source, err := scanCalendarSource(r.q.QueryRow(ctx, sqlClearCalendarSourceEventSyncCursor, id, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarSource{}, apperr.ErrNotFound
		}
		return entity.CalendarSource{}, fmt.Errorf("clear calendar source event sync cursor: %w", err)
	}
	return source, nil
}

func (r *calendarSourceRepository) UpdateSyncEnabledByAccount(
	ctx context.Context,
	accountID uuid.UUID,
	syncEnabled bool,
	updatedAt time.Time,
) (int64, error) {
	tag, err := r.q.Exec(ctx, sqlUpdateCalendarSourcesSyncEnabledByAccount, accountID, syncEnabled, updatedAt)
	if err != nil {
		return 0, fmt.Errorf("update calendar source sync_enabled: %w", err)
	}
	return tag.RowsAffected(), nil
}

func collectCalendarSources(rows pgx.Rows) ([]entity.CalendarSource, error) {
	out := make([]entity.CalendarSource, 0)
	for rows.Next() {
		source, err := scanCalendarSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanCalendarSource(row scannable) (entity.CalendarSource, error) {
	var s entity.CalendarSource
	err := row.Scan(
		&s.ID,
		&s.PublicID,
		&s.UserID,
		&s.ConnectedAccountID,
		&s.ProviderCalendarID,
		&s.Name,
		&s.Color,
		&s.IsPrimaryOnProvider,
		&s.IsWritable,
		&s.AccessRole,
		&s.SyncEnabled,
		&s.SyncCursor,
		&s.LastSyncedAt,
		&s.Timezone,
		&s.ProviderMetadata,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
	)
	return s, err
}
