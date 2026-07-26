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

const calendarEventColumns = `
	id, public_id, user_id, calendar_source_id, title, description, location,
	starts_at, ends_at, is_all_day, status, visibility, timezone, organizer_summary,
	attendees_summary, recurrence_rule, recurring_event_id, provider_recurring_event_id,
	provider_event_id, provider_etag, provider_updated_at, provider_payload, origin,
	created_at, updated_at, deleted_at`

const (
	sqlInsertCalendarEvent = `
INSERT INTO calendar_events (
	id, public_id, user_id, calendar_source_id, title, description, location,
	starts_at, ends_at, is_all_day, status, visibility, timezone, organizer_summary,
	attendees_summary, recurrence_rule, recurring_event_id, provider_recurring_event_id,
	provider_event_id, provider_etag, provider_updated_at, provider_payload, origin,
	created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,$15::jsonb,$16,$17,$18,
	$19,$20,$21,$22::jsonb,$23,$24,$25
)
RETURNING` + calendarEventColumns

	sqlSelectCalendarEventBySourceProvider = `
SELECT` + calendarEventColumns + `
FROM calendar_events
WHERE calendar_source_id = $1 AND provider_event_id = $2
ORDER BY deleted_at NULLS FIRST
LIMIT 1`

	sqlSelectCalendarEventsByUserRange = `
SELECT` + calendarEventColumns + `
FROM calendar_events
WHERE user_id = $1
  AND deleted_at IS NULL
  AND starts_at < $3
  AND ends_at > $2
ORDER BY starts_at ASC`

	sqlUpdateCalendarEventFromSync = `
UPDATE calendar_events SET
	title = $2,
	description = $3,
	location = $4,
	starts_at = $5,
	ends_at = $6,
	is_all_day = $7,
	status = $8,
	visibility = $9,
	timezone = $10,
	organizer_summary = $11::jsonb,
	attendees_summary = $12::jsonb,
	recurrence_rule = $13,
	recurring_event_id = $14,
	provider_recurring_event_id = $15,
	provider_etag = $16,
	provider_updated_at = $17,
	provider_payload = $18::jsonb,
	origin = $19,
	deleted_at = NULL,
	updated_at = $20
WHERE id = $1
RETURNING` + calendarEventColumns

	sqlSoftDeleteCalendarEventBySourceProvider = `
UPDATE calendar_events SET
	status = 'cancelled',
	deleted_at = $3,
	updated_at = $3
WHERE calendar_source_id = $1
  AND provider_event_id = $2
  AND deleted_at IS NULL
RETURNING` + calendarEventColumns

	sqlSoftDeleteCalendarEventsMissing = `
UPDATE calendar_events SET
	status = 'cancelled',
	deleted_at = $3,
	updated_at = $3
WHERE calendar_source_id = $1
  AND deleted_at IS NULL
  AND provider_event_id IS NOT NULL
  AND NOT (provider_event_id = ANY($2::text[]))`
)

// CalendarEventRepository persists calendar events.
type CalendarEventRepository interface {
	Create(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error)
	GetBySourceAndProviderEvent(ctx context.Context, sourceID uuid.UUID, providerEventID string) (entity.CalendarEvent, error)
	ListByUserInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.CalendarEvent, error)
	UpdateFromSync(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error)
	SoftDeleteByProviderEventID(ctx context.Context, sourceID uuid.UUID, providerEventID string, deletedAt time.Time) (entity.CalendarEvent, error)
	SoftDeleteMissing(ctx context.Context, sourceID uuid.UUID, keepProviderIDs []string, deletedAt time.Time) (int64, error)
	WithTx(tx pgx.Tx) CalendarEventRepository
}

type calendarEventRepository struct {
	q Querier
}

// NewCalendarEventRepository constructs a CalendarEventRepository.
func NewCalendarEventRepository(pool *pgxpool.Pool) CalendarEventRepository {
	return &calendarEventRepository{q: pool}
}

func (r *calendarEventRepository) WithTx(tx pgx.Tx) CalendarEventRepository {
	return &calendarEventRepository{q: tx}
}

func (r *calendarEventRepository) Create(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error) {
	attendees := event.AttendeesSummary
	if len(attendees) == 0 {
		attendees = []byte("[]")
	}
	var organizer any
	if len(event.OrganizerSummary) > 0 {
		organizer = event.OrganizerSummary
	}
	var payload any
	if len(event.ProviderPayload) > 0 {
		payload = event.ProviderPayload
	}
	created, err := scanCalendarEvent(r.q.QueryRow(ctx, sqlInsertCalendarEvent,
		event.ID,
		event.PublicID,
		event.UserID,
		event.CalendarSourceID,
		event.Title,
		event.Description,
		event.Location,
		event.StartsAt,
		event.EndsAt,
		event.IsAllDay,
		event.Status,
		event.Visibility,
		event.Timezone,
		organizer,
		attendees,
		event.RecurrenceRule,
		event.RecurringEventID,
		event.ProviderRecurringEventID,
		event.ProviderEventID,
		event.ProviderETag,
		event.ProviderUpdatedAt,
		payload,
		event.Origin,
		event.CreatedAt,
		event.UpdatedAt,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return entity.CalendarEvent{}, apperr.ErrConflict
		}
		return entity.CalendarEvent{}, fmt.Errorf("insert calendar event: %w", err)
	}
	return created, nil
}

func (r *calendarEventRepository) GetBySourceAndProviderEvent(
	ctx context.Context,
	sourceID uuid.UUID,
	providerEventID string,
) (entity.CalendarEvent, error) {
	event, err := scanCalendarEvent(r.q.QueryRow(ctx, sqlSelectCalendarEventBySourceProvider, sourceID, providerEventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarEvent{}, apperr.ErrNotFound
		}
		return entity.CalendarEvent{}, fmt.Errorf("get calendar event: %w", err)
	}
	return event, nil
}

func (r *calendarEventRepository) ListByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.CalendarEvent, error) {
	rows, err := r.q.Query(ctx, sqlSelectCalendarEventsByUserRange, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	defer rows.Close()
	return collectCalendarEvents(rows)
}

func (r *calendarEventRepository) UpdateFromSync(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error) {
	attendees := event.AttendeesSummary
	if len(attendees) == 0 {
		attendees = []byte("[]")
	}
	var organizer any
	if len(event.OrganizerSummary) > 0 {
		organizer = event.OrganizerSummary
	}
	var payload any
	if len(event.ProviderPayload) > 0 {
		payload = event.ProviderPayload
	}
	updated, err := scanCalendarEvent(r.q.QueryRow(ctx, sqlUpdateCalendarEventFromSync,
		event.ID,
		event.Title,
		event.Description,
		event.Location,
		event.StartsAt,
		event.EndsAt,
		event.IsAllDay,
		event.Status,
		event.Visibility,
		event.Timezone,
		organizer,
		attendees,
		event.RecurrenceRule,
		event.RecurringEventID,
		event.ProviderRecurringEventID,
		event.ProviderETag,
		event.ProviderUpdatedAt,
		payload,
		event.Origin,
		event.UpdatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarEvent{}, apperr.ErrNotFound
		}
		return entity.CalendarEvent{}, fmt.Errorf("update calendar event from sync: %w", err)
	}
	return updated, nil
}

func (r *calendarEventRepository) SoftDeleteByProviderEventID(
	ctx context.Context,
	sourceID uuid.UUID,
	providerEventID string,
	deletedAt time.Time,
) (entity.CalendarEvent, error) {
	event, err := scanCalendarEvent(r.q.QueryRow(ctx, sqlSoftDeleteCalendarEventBySourceProvider, sourceID, providerEventID, deletedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarEvent{}, apperr.ErrNotFound
		}
		return entity.CalendarEvent{}, fmt.Errorf("soft-delete calendar event: %w", err)
	}
	return event, nil
}

func (r *calendarEventRepository) SoftDeleteMissing(
	ctx context.Context,
	sourceID uuid.UUID,
	keepProviderIDs []string,
	deletedAt time.Time,
) (int64, error) {
	if keepProviderIDs == nil {
		keepProviderIDs = []string{}
	}
	tag, err := r.q.Exec(ctx, sqlSoftDeleteCalendarEventsMissing, sourceID, keepProviderIDs, deletedAt)
	if err != nil {
		return 0, fmt.Errorf("soft-delete missing calendar events: %w", err)
	}
	return tag.RowsAffected(), nil
}

func collectCalendarEvents(rows pgx.Rows) ([]entity.CalendarEvent, error) {
	out := make([]entity.CalendarEvent, 0)
	for rows.Next() {
		event, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func scanCalendarEvent(row scannable) (entity.CalendarEvent, error) {
	var e entity.CalendarEvent
	err := row.Scan(
		&e.ID,
		&e.PublicID,
		&e.UserID,
		&e.CalendarSourceID,
		&e.Title,
		&e.Description,
		&e.Location,
		&e.StartsAt,
		&e.EndsAt,
		&e.IsAllDay,
		&e.Status,
		&e.Visibility,
		&e.Timezone,
		&e.OrganizerSummary,
		&e.AttendeesSummary,
		&e.RecurrenceRule,
		&e.RecurringEventID,
		&e.ProviderRecurringEventID,
		&e.ProviderEventID,
		&e.ProviderETag,
		&e.ProviderUpdatedAt,
		&e.ProviderPayload,
		&e.Origin,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	)
	return e, err
}
