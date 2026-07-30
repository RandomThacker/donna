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

const donnaEventColumns = `
	id, public_id, user_id, title, description, start_at, end_at, timezone, all_day,
	location, reminder_offset_minutes, recurrence_rule, status, color,
	created_at, updated_at, deleted_at`

const (
	sqlInsertDonnaEvent = `
INSERT INTO donna_events (
	id, public_id, user_id, title, description, start_at, end_at, timezone, all_day,
	location, reminder_offset_minutes, recurrence_rule, status, color,
	created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
)
RETURNING` + donnaEventColumns

	sqlSelectDonnaEventByID = `
SELECT` + donnaEventColumns + `
FROM donna_events
WHERE id = $1 AND deleted_at IS NULL`

	sqlListDonnaEventsByUserRange = `
SELECT` + donnaEventColumns + `
FROM donna_events
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status <> 'cancelled'
  AND (
    (
      (recurrence_rule IS NULL OR btrim(recurrence_rule) = '')
      AND start_at < $3
      AND end_at > $2
    )
    OR (
      recurrence_rule IS NOT NULL AND btrim(recurrence_rule) <> ''
      AND start_at < $3
    )
  )
ORDER BY start_at ASC`

	sqlListDonnaEventsByUser = `
SELECT` + donnaEventColumns + `
FROM donna_events
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY start_at ASC`

	sqlUpdateDonnaEvent = `
UPDATE donna_events SET
	title = COALESCE($2, title),
	description = COALESCE($3, description),
	start_at = COALESCE($4, start_at),
	end_at = COALESCE($5, end_at),
	timezone = COALESCE($6, timezone),
	all_day = COALESCE($7, all_day),
	location = COALESCE($8, location),
	reminder_offset_minutes = COALESCE($9, reminder_offset_minutes),
	recurrence_rule = COALESCE($10, recurrence_rule),
	status = COALESCE($11, status),
	color = COALESCE($12, color),
	updated_at = $13
WHERE id = $1 AND user_id = $14 AND deleted_at IS NULL
RETURNING` + donnaEventColumns

	sqlSoftDeleteDonnaEvent = `
UPDATE donna_events SET deleted_at = $3, updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
)

// DonnaEventRepository persists Donna-owned timeline events.
type DonnaEventRepository interface {
	Create(ctx context.Context, event entity.DonnaEvent) (entity.DonnaEvent, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.DonnaEvent, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DonnaEvent, error)
	ListByUserInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.DonnaEvent, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields DonnaEventUpdateFields, updatedAt time.Time) (entity.DonnaEvent, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error
	WithTx(tx pgx.Tx) DonnaEventRepository
}

// DonnaEventUpdateFields are optional patches for a Donna event.
type DonnaEventUpdateFields struct {
	Title                 *string
	Description           *string
	StartAt               *time.Time
	EndAt                 *time.Time
	Timezone              *string
	AllDay                *bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Status                *string
	Color                 *string
}

type donnaEventRepository struct {
	q Querier
}

// NewDonnaEventRepository constructs a DonnaEventRepository.
func NewDonnaEventRepository(pool *pgxpool.Pool) DonnaEventRepository {
	return &donnaEventRepository{q: pool}
}

func (r *donnaEventRepository) WithTx(tx pgx.Tx) DonnaEventRepository {
	return &donnaEventRepository{q: tx}
}

func (r *donnaEventRepository) Create(ctx context.Context, event entity.DonnaEvent) (entity.DonnaEvent, error) {
	return scanDonnaEvent(r.q.QueryRow(ctx, sqlInsertDonnaEvent,
		event.ID, event.PublicID, event.UserID, event.Title, event.Description,
		event.StartAt, event.EndAt, event.Timezone, event.AllDay, event.Location,
		event.ReminderOffsetMinutes, event.RecurrenceRule, event.Status, event.Color,
		event.CreatedAt, event.UpdatedAt,
	))
}

func (r *donnaEventRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.DonnaEvent, error) {
	return scanDonnaEvent(r.q.QueryRow(ctx, sqlSelectDonnaEventByID, id))
}

func (r *donnaEventRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DonnaEvent, error) {
	rows, err := r.q.Query(ctx, sqlListDonnaEventsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectDonnaEvents(rows)
}

func (r *donnaEventRepository) ListByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.DonnaEvent, error) {
	rows, err := r.q.Query(ctx, sqlListDonnaEventsByUserRange, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectDonnaEvents(rows)
}

func (r *donnaEventRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	fields DonnaEventUpdateFields,
	updatedAt time.Time,
) (entity.DonnaEvent, error) {
	return scanDonnaEvent(r.q.QueryRow(ctx, sqlUpdateDonnaEvent,
		id, fields.Title, fields.Description, fields.StartAt, fields.EndAt,
		fields.Timezone, fields.AllDay, fields.Location, fields.ReminderOffsetMinutes,
		fields.RecurrenceRule, fields.Status, fields.Color, updatedAt, userID,
	))
}

func (r *donnaEventRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteDonnaEvent, id, userID, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func collectDonnaEvents(rows pgx.Rows) ([]entity.DonnaEvent, error) {
	out := make([]entity.DonnaEvent, 0)
	for rows.Next() {
		event, err := scanDonnaEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func scanDonnaEvent(row pgx.Row) (entity.DonnaEvent, error) {
	var e entity.DonnaEvent
	err := row.Scan(
		&e.ID, &e.PublicID, &e.UserID, &e.Title, &e.Description,
		&e.StartAt, &e.EndAt, &e.Timezone, &e.AllDay, &e.Location,
		&e.ReminderOffsetMinutes, &e.RecurrenceRule, &e.Status, &e.Color,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.DonnaEvent{}, apperr.ErrNotFound
		}
		return entity.DonnaEvent{}, fmt.Errorf("scan donna event: %w", err)
	}
	return e, nil
}
