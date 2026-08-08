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

const calendarEventColumnsAliased = `
	e.id, e.public_id, e.user_id, e.calendar_source_id, e.title, e.description, e.location,
	e.starts_at, e.ends_at, e.is_all_day, e.status, e.visibility, e.timezone, e.organizer_summary,
	e.attendees_summary, e.recurrence_rule, e.recurring_event_id, e.provider_recurring_event_id,
	e.provider_event_id, e.provider_etag, e.provider_updated_at, e.provider_payload, e.origin,
	e.created_at, e.updated_at, e.deleted_at`

// Sync decision lookup: fields required by shouldSkipEventUpdate / content hash /
// resurrection / identity preserve — deliberately omits provider_payload.
const calendarEventSyncDecisionColumns = `
	id, public_id, calendar_source_id, title, description, location,
	starts_at, ends_at, is_all_day, status, visibility, timezone, organizer_summary,
	attendees_summary, recurrence_rule, recurring_event_id, provider_event_id,
	provider_etag, provider_updated_at, created_at, deleted_at`

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

	sqlSelectCalendarEventSyncDecisionBySourceProvider = `
SELECT` + calendarEventSyncDecisionColumns + `
FROM calendar_events
WHERE calendar_source_id = $1 AND provider_event_id = $2
ORDER BY deleted_at NULLS FIRST
LIMIT 1`

	// DISTINCT ON picks the live row first (deleted_at NULLS FIRST), matching the
	// single-row GetForSyncDecision ORDER BY deleted_at NULLS FIRST LIMIT 1.
	sqlSelectCalendarEventSyncDecisionByProviderEventIDs = `
SELECT DISTINCT ON (provider_event_id)` + calendarEventSyncDecisionColumns + `
FROM calendar_events
WHERE calendar_source_id = $1
  AND provider_event_id = ANY($2::text[])
ORDER BY provider_event_id, deleted_at NULLS FIRST`

	sqlSelectCalendarEventsByUserRange = `
SELECT` + calendarEventColumnsAliased + `
FROM calendar_events e
JOIN calendar_sources s ON s.id = e.calendar_source_id
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE e.user_id = $1
  AND e.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND ca.deleted_at IS NULL
  AND e.starts_at < $3
  AND e.ends_at > $2
ORDER BY e.starts_at ASC`

	sqlSelectCalendarEventsByUserRangeWithProvider = `
SELECT` + calendarEventColumnsAliased + `, ca.provider, s.color
FROM calendar_events e
JOIN calendar_sources s ON s.id = e.calendar_source_id
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE e.user_id = $1
  AND e.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND ca.deleted_at IS NULL
  AND e.starts_at < $3
  AND e.ends_at > $2
ORDER BY e.starts_at ASC`

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

	sqlDeleteCalendarEventsByAccount = `
DELETE FROM calendar_events e
USING calendar_sources s
WHERE e.calendar_source_id = s.id
  AND s.connected_account_id = $1`

	sqlDeleteOrphanCalendarEventsForUser = `
DELETE FROM calendar_events e
USING calendar_sources s
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE e.calendar_source_id = s.id
  AND e.user_id = $1
  AND (s.deleted_at IS NOT NULL OR ca.deleted_at IS NOT NULL)`

	sqlCountLiveCalendarEventsByAccount = `
SELECT COUNT(*)
FROM calendar_events e
JOIN calendar_sources s ON s.id = e.calendar_source_id
WHERE s.connected_account_id = $1
  AND e.deleted_at IS NULL
  AND s.deleted_at IS NULL`
)

// CalendarEventRepository persists calendar events.
type CalendarEventRepository interface {
	Create(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error)
	GetBySourceAndProviderEvent(ctx context.Context, sourceID uuid.UUID, providerEventID string) (entity.CalendarEvent, error)
	// GetForSyncDecision returns a narrow projection for calendar sync skip/update
	// decisions. ProviderPayload is never loaded.
	GetForSyncDecision(ctx context.Context, sourceID uuid.UUID, providerEventID string) (entity.CalendarEvent, error)
	// GetForSyncDecisionByProviderEventIDs batch-loads narrow sync-decision rows for
	// one source. For each provider_event_id, prefers the live row (deleted_at IS NULL)
	// over soft-deleted duplicates — same semantics as GetForSyncDecision.
	// ProviderPayload is never loaded. Empty ids returns an empty map without querying.
	GetForSyncDecisionByProviderEventIDs(
		ctx context.Context,
		sourceID uuid.UUID,
		providerEventIDs []string,
	) (map[string]entity.CalendarEvent, error)
	ListByUserInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.CalendarEvent, error)
	ListByUserInRangeWithProvider(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.CalendarEventWithProvider, error)
	// ListForSchedulerByUserInRange returns a narrow projection for Occurrence scheduling.
	// providers filters connected_accounts.provider (e.g. ["google"]).
	ListForSchedulerByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
		providers []string,
	) ([]entity.CalendarEventWithProvider, error)
	// ListCalendarOccurrences is the Sprint 1B alias for the shared Calendar Occurrence query.
	ListCalendarOccurrences(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
		providers []string,
	) ([]entity.CalendarEventWithProvider, error)
	UpdateFromSync(ctx context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error)
	SoftDeleteByProviderEventID(ctx context.Context, sourceID uuid.UUID, providerEventID string, deletedAt time.Time) (entity.CalendarEvent, error)
	SoftDeleteMissing(ctx context.Context, sourceID uuid.UUID, keepProviderIDs []string, deletedAt time.Time) (int64, error)
	DeleteByConnectedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
	DeleteOrphansForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountLiveByConnectedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
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

func (r *calendarEventRepository) GetForSyncDecision(
	ctx context.Context,
	sourceID uuid.UUID,
	providerEventID string,
) (entity.CalendarEvent, error) {
	event, err := scanCalendarEventSyncDecision(r.q.QueryRow(ctx, sqlSelectCalendarEventSyncDecisionBySourceProvider, sourceID, providerEventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.CalendarEvent{}, apperr.ErrNotFound
		}
		return entity.CalendarEvent{}, fmt.Errorf("get calendar event sync decision: %w", err)
	}
	return event, nil
}

func (r *calendarEventRepository) GetForSyncDecisionByProviderEventIDs(
	ctx context.Context,
	sourceID uuid.UUID,
	providerEventIDs []string,
) (map[string]entity.CalendarEvent, error) {
	out := make(map[string]entity.CalendarEvent)
	if len(providerEventIDs) == 0 {
		return out, nil
	}
	// Deduplicate while preserving a stable query size bound.
	seen := make(map[string]struct{}, len(providerEventIDs))
	ids := make([]string, 0, len(providerEventIDs))
	for _, id := range providerEventIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := r.q.Query(ctx, sqlSelectCalendarEventSyncDecisionByProviderEventIDs, sourceID, ids)
	if err != nil {
		return nil, fmt.Errorf("list calendar event sync decisions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		event, scanErr := scanCalendarEventSyncDecision(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if event.ProviderEventID == nil || *event.ProviderEventID == "" {
			continue
		}
		out[*event.ProviderEventID] = event
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list calendar event sync decisions: %w", err)
	}
	return out, nil
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

func (r *calendarEventRepository) ListByUserInRangeWithProvider(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.CalendarEventWithProvider, error) {
	rows, err := r.q.Query(ctx, sqlSelectCalendarEventsByUserRangeWithProvider, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list calendar events with provider: %w", err)
	}
	defer rows.Close()
	return collectCalendarEventsWithProvider(rows)
}

func (r *calendarEventRepository) ListForSchedulerByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	if providers == nil {
		providers = []string{}
	}
	rows, err := r.q.Query(ctx, sqlSelectCalendarEventsForScheduler, userID, from, to, providers)
	if err != nil {
		return nil, fmt.Errorf("list calendar events for scheduler: %w", err)
	}
	defer rows.Close()
	return collectCalendarEventsForScheduler(rows)
}

// ListCalendarOccurrences runs one narrow calendar_events query for the given
// provider filters (Sprint 1B shared Occurrence path).
func (r *calendarEventRepository) ListCalendarOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	return r.ListForSchedulerByUserInRange(ctx, userID, from, to, providers)
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

func (r *calendarEventRepository) DeleteByConnectedAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) (int64, error) {
	tag, err := r.q.Exec(ctx, sqlDeleteCalendarEventsByAccount, accountID)
	if err != nil {
		return 0, fmt.Errorf("delete calendar events by account: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarEventRepository) DeleteOrphansForUser(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	tag, err := r.q.Exec(ctx, sqlDeleteOrphanCalendarEventsForUser, userID)
	if err != nil {
		return 0, fmt.Errorf("delete orphan calendar events: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *calendarEventRepository) CountLiveByConnectedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var n int64
	if err := r.q.QueryRow(ctx, sqlCountLiveCalendarEventsByAccount, accountID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count live calendar events by account: %w", err)
	}
	return n, nil
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

func collectCalendarEventsWithProvider(rows pgx.Rows) ([]entity.CalendarEventWithProvider, error) {
	out := make([]entity.CalendarEventWithProvider, 0)
	for rows.Next() {
		item, err := scanCalendarEventWithProvider(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func collectCalendarEventsForScheduler(rows pgx.Rows) ([]entity.CalendarEventWithProvider, error) {
	out := make([]entity.CalendarEventWithProvider, 0)
	for rows.Next() {
		item, err := scanCalendarEventForScheduler(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
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

func scanCalendarEventSyncDecision(row scannable) (entity.CalendarEvent, error) {
	var e entity.CalendarEvent
	err := row.Scan(
		&e.ID,
		&e.PublicID,
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
		&e.ProviderEventID,
		&e.ProviderETag,
		&e.ProviderUpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
	)
	return e, err
}

func scanCalendarEventWithProvider(row scannable) (entity.CalendarEventWithProvider, error) {
	var e entity.CalendarEvent
	var provider string
	var sourceColor *string
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
		&provider,
		&sourceColor,
	)
	if err != nil {
		return entity.CalendarEventWithProvider{}, err
	}
	return entity.CalendarEventWithProvider{
		Event:       e,
		Provider:    provider,
		SourceColor: sourceColor,
	}, nil
}

func scanCalendarEventForScheduler(row scannable) (entity.CalendarEventWithProvider, error) {
	var e entity.CalendarEvent
	var provider string
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
		&e.Status,
		&e.Timezone,
		&e.ProviderEventID,
		&e.Origin,
		&provider,
	)
	if err != nil {
		return entity.CalendarEventWithProvider{}, err
	}
	return entity.CalendarEventWithProvider{
		Event:    e,
		Provider: provider,
	}, nil
}
