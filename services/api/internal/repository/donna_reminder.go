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

const donnaReminderColumns = `
	id, public_id, user_id, title, description, trigger_at, timezone,
	recurrence_rule, status, color, created_at, updated_at, deleted_at`

const (
	sqlInsertDonnaReminder = `
INSERT INTO donna_reminders (
	id, public_id, user_id, title, description, trigger_at, timezone,
	recurrence_rule, status, color, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
)
RETURNING` + donnaReminderColumns

	sqlSelectDonnaReminderByID = `
SELECT` + donnaReminderColumns + `
FROM donna_reminders
WHERE id = $1 AND deleted_at IS NULL`

	sqlListDonnaRemindersByUserRange = `
SELECT` + donnaReminderColumns + `
FROM donna_reminders
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status <> 'cancelled'
  AND (
    (
      (recurrence_rule IS NULL OR btrim(recurrence_rule) = '')
      AND trigger_at >= $2
      AND trigger_at < $3
    )
    OR (
      recurrence_rule IS NOT NULL AND btrim(recurrence_rule) <> ''
      AND trigger_at < $3
    )
  )
ORDER BY trigger_at ASC`

	sqlListDonnaRemindersByUser = `
SELECT` + donnaReminderColumns + `
FROM donna_reminders
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY trigger_at ASC`

	sqlUpdateDonnaReminder = `
UPDATE donna_reminders SET
	title = COALESCE($2, title),
	description = COALESCE($3, description),
	trigger_at = COALESCE($4, trigger_at),
	timezone = COALESCE($5, timezone),
	recurrence_rule = COALESCE($6, recurrence_rule),
	status = COALESCE($7, status),
	color = COALESCE($8, color),
	updated_at = $9
WHERE id = $1 AND user_id = $10 AND deleted_at IS NULL
RETURNING` + donnaReminderColumns

	sqlSoftDeleteDonnaReminder = `
UPDATE donna_reminders SET deleted_at = $3, updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
)

// DonnaReminderRepository persists Donna-owned timeline reminders.
type DonnaReminderRepository interface {
	Create(ctx context.Context, reminder entity.DonnaReminder) (entity.DonnaReminder, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.DonnaReminder, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DonnaReminder, error)
	ListByUserInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.DonnaReminder, error)
	// ListForSchedulerByUserInRange returns a narrow projection for Occurrence scheduling.
	ListForSchedulerByUserInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.DonnaReminder, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields DonnaReminderUpdateFields, updatedAt time.Time) (entity.DonnaReminder, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error
	WithTx(tx pgx.Tx) DonnaReminderRepository
}

// DonnaReminderUpdateFields are optional patches for a Donna reminder.
type DonnaReminderUpdateFields struct {
	Title          *string
	Description    *string
	TriggerAt      *time.Time
	Timezone       *string
	RecurrenceRule *string
	Status         *string
	Color          *string
}

type donnaReminderRepository struct {
	q Querier
}

// NewDonnaReminderRepository constructs a DonnaReminderRepository.
func NewDonnaReminderRepository(pool *pgxpool.Pool) DonnaReminderRepository {
	return &donnaReminderRepository{q: pool}
}

func (r *donnaReminderRepository) WithTx(tx pgx.Tx) DonnaReminderRepository {
	return &donnaReminderRepository{q: tx}
}

func (r *donnaReminderRepository) Create(ctx context.Context, reminder entity.DonnaReminder) (entity.DonnaReminder, error) {
	return scanDonnaReminder(r.q.QueryRow(ctx, sqlInsertDonnaReminder,
		reminder.ID, reminder.PublicID, reminder.UserID, reminder.Title, reminder.Description,
		reminder.TriggerAt, reminder.Timezone, reminder.RecurrenceRule, reminder.Status,
		reminder.Color, reminder.CreatedAt, reminder.UpdatedAt,
	))
}

func (r *donnaReminderRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.DonnaReminder, error) {
	return scanDonnaReminder(r.q.QueryRow(ctx, sqlSelectDonnaReminderByID, id))
}

func (r *donnaReminderRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.DonnaReminder, error) {
	rows, err := r.q.Query(ctx, sqlListDonnaRemindersByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectDonnaReminders(rows)
}

func (r *donnaReminderRepository) ListByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.DonnaReminder, error) {
	rows, err := r.q.Query(ctx, sqlListDonnaRemindersByUserRange, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectDonnaReminders(rows)
}

func (r *donnaReminderRepository) ListForSchedulerByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.DonnaReminder, error) {
	rows, err := r.q.Query(ctx, sqlListDonnaRemindersForScheduler, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectDonnaRemindersForScheduler(rows)
}

func (r *donnaReminderRepository) Update(
	ctx context.Context,
	id, userID uuid.UUID,
	fields DonnaReminderUpdateFields,
	updatedAt time.Time,
) (entity.DonnaReminder, error) {
	return scanDonnaReminder(r.q.QueryRow(ctx, sqlUpdateDonnaReminder,
		id, fields.Title, fields.Description, fields.TriggerAt, fields.Timezone,
		fields.RecurrenceRule, fields.Status, fields.Color, updatedAt, userID,
	))
}

func (r *donnaReminderRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteDonnaReminder, id, userID, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func collectDonnaReminders(rows pgx.Rows) ([]entity.DonnaReminder, error) {
	out := make([]entity.DonnaReminder, 0)
	for rows.Next() {
		reminder, err := scanDonnaReminder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, reminder)
	}
	return out, rows.Err()
}

func collectDonnaRemindersForScheduler(rows pgx.Rows) ([]entity.DonnaReminder, error) {
	out := make([]entity.DonnaReminder, 0)
	for rows.Next() {
		reminder, err := scanDonnaReminderForScheduler(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, reminder)
	}
	return out, rows.Err()
}

func scanDonnaReminder(row pgx.Row) (entity.DonnaReminder, error) {
	var rem entity.DonnaReminder
	err := row.Scan(
		&rem.ID, &rem.PublicID, &rem.UserID, &rem.Title, &rem.Description,
		&rem.TriggerAt, &rem.Timezone, &rem.RecurrenceRule, &rem.Status, &rem.Color,
		&rem.CreatedAt, &rem.UpdatedAt, &rem.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.DonnaReminder{}, apperr.ErrNotFound
		}
		return entity.DonnaReminder{}, fmt.Errorf("scan donna reminder: %w", err)
	}
	return rem, nil
}

func scanDonnaReminderForScheduler(row pgx.Row) (entity.DonnaReminder, error) {
	var rem entity.DonnaReminder
	err := row.Scan(
		&rem.ID, &rem.PublicID, &rem.UserID, &rem.Title, &rem.Description,
		&rem.TriggerAt, &rem.Timezone, &rem.RecurrenceRule, &rem.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.DonnaReminder{}, apperr.ErrNotFound
		}
		return entity.DonnaReminder{}, fmt.Errorf("scan donna reminder for scheduler: %w", err)
	}
	return rem, nil
}
