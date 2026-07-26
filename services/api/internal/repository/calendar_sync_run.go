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

const calendarSyncRunColumns = `
	id, public_id, user_id, connected_account_id, trigger, status,
	started_at, finished_at, duration_ms, calendars_processed,
	sources_created, sources_updated, sources_deleted,
	events_created, events_updated, events_deleted, failures, created_at`

const (
	sqlInsertCalendarSyncRun = `
INSERT INTO calendar_sync_runs (
	id, public_id, user_id, connected_account_id, trigger, status,
	started_at, finished_at, duration_ms, calendars_processed,
	sources_created, sources_updated, sources_deleted,
	events_created, events_updated, events_deleted, failures, created_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17::jsonb,$18
)
RETURNING` + calendarSyncRunColumns

	sqlUpdateCalendarSyncRunFinish = `
UPDATE calendar_sync_runs SET
	status = $2,
	finished_at = $3,
	duration_ms = $4,
	calendars_processed = $5,
	sources_created = $6,
	sources_updated = $7,
	sources_deleted = $8,
	events_created = $9,
	events_updated = $10,
	events_deleted = $11,
	failures = $12::jsonb
WHERE id = $1
RETURNING` + calendarSyncRunColumns
)

// CalendarSyncRunRepository persists sync pipeline attempts.
type CalendarSyncRunRepository interface {
	Create(ctx context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error)
	Finish(ctx context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error)
	WithTx(tx pgx.Tx) CalendarSyncRunRepository
}

type calendarSyncRunRepository struct {
	q Querier
}

// NewCalendarSyncRunRepository constructs a CalendarSyncRunRepository.
func NewCalendarSyncRunRepository(pool *pgxpool.Pool) CalendarSyncRunRepository {
	return &calendarSyncRunRepository{q: pool}
}

func (r *calendarSyncRunRepository) WithTx(tx pgx.Tx) CalendarSyncRunRepository {
	return &calendarSyncRunRepository{q: tx}
}

func (r *calendarSyncRunRepository) Create(ctx context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error) {
	failures := run.Failures
	if len(failures) == 0 {
		failures = []byte("[]")
	}
	created, err := scanCalendarSyncRun(r.q.QueryRow(ctx, sqlInsertCalendarSyncRun,
		run.ID,
		run.PublicID,
		run.UserID,
		run.ConnectedAccountID,
		run.Trigger,
		run.Status,
		run.StartedAt,
		run.FinishedAt,
		run.DurationMs,
		run.CalendarsProcessed,
		run.SourcesCreated,
		run.SourcesUpdated,
		run.SourcesDeleted,
		run.EventsCreated,
		run.EventsUpdated,
		run.EventsDeleted,
		failures,
		run.CreatedAt,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return entity.CalendarSyncRun{}, apperr.ErrConflict
		}
		return entity.CalendarSyncRun{}, fmt.Errorf("insert calendar sync run: %w", err)
	}
	return created, nil
}

func (r *calendarSyncRunRepository) Finish(ctx context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error) {
	failures := run.Failures
	if len(failures) == 0 {
		failures = []byte("[]")
	}
	updated, err := scanCalendarSyncRun(r.q.QueryRow(ctx, sqlUpdateCalendarSyncRunFinish,
		run.ID,
		run.Status,
		run.FinishedAt,
		run.DurationMs,
		run.CalendarsProcessed,
		run.SourcesCreated,
		run.SourcesUpdated,
		run.SourcesDeleted,
		run.EventsCreated,
		run.EventsUpdated,
		run.EventsDeleted,
		failures,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarSyncRun{}, apperr.ErrNotFound
		}
		return entity.CalendarSyncRun{}, fmt.Errorf("finish calendar sync run: %w", err)
	}
	return updated, nil
}

func scanCalendarSyncRun(row scannable) (entity.CalendarSyncRun, error) {
	var run entity.CalendarSyncRun
	err := row.Scan(
		&run.ID,
		&run.PublicID,
		&run.UserID,
		&run.ConnectedAccountID,
		&run.Trigger,
		&run.Status,
		&run.StartedAt,
		&run.FinishedAt,
		&run.DurationMs,
		&run.CalendarsProcessed,
		&run.SourcesCreated,
		&run.SourcesUpdated,
		&run.SourcesDeleted,
		&run.EventsCreated,
		&run.EventsUpdated,
		&run.EventsDeleted,
		&run.Failures,
		&run.CreatedAt,
	)
	return run, err
}
