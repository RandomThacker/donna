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

// Bare columns for INSERT/UPDATE RETURNING (no table alias).
const occurrenceColumns = `
	id, public_id, task_id, user_id, date, sort_order,
	completed, completed_at, carried_forward, source,
	created_at, updated_at`

// Aliased columns for SELECTs that join as "o".
const occurrenceColumnsAliased = `
	o.id, o.public_id, o.task_id, o.user_id, o.date, o.sort_order,
	o.completed, o.completed_at, o.carried_forward, o.source,
	o.created_at, o.updated_at`

const sqlInsertTaskOccurrence = `
INSERT INTO task_occurrences (
	id, public_id, task_id, user_id, date, sort_order,
	completed, completed_at, carried_forward, source, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING` + occurrenceColumns

const sqlSelectOccurrencesByUserDate = `
SELECT` + occurrenceColumnsAliased + `,
	t.title, t.description, t.priority, t.project, t.labels, t.recurrence_rule
FROM task_occurrences o
JOIN tasks t ON t.id = o.task_id AND t.deleted_at IS NULL
WHERE o.user_id = $1 AND o.date = $2
ORDER BY o.completed ASC, o.sort_order ASC, o.created_at ASC`

const sqlSelectOccurrenceByID = `
SELECT` + occurrenceColumnsAliased + `,
	t.title, t.description, t.priority, t.project, t.labels, t.recurrence_rule
FROM task_occurrences o
JOIN tasks t ON t.id = o.task_id AND t.deleted_at IS NULL
WHERE o.id = $1`

const sqlCountOccurrencesByUserDate = `
SELECT COUNT(*) FROM task_occurrences WHERE user_id = $1 AND date = $2`

const sqlListIncompleteByUserDate = `
SELECT` + occurrenceColumnsAliased + `
FROM task_occurrences o
JOIN tasks t ON t.id = o.task_id AND t.deleted_at IS NULL
WHERE o.user_id = $1 AND o.date = $2 AND o.completed = false
ORDER BY o.sort_order ASC`

const sqlMaxSortOrder = `
SELECT COALESCE(MAX(sort_order), -1) FROM task_occurrences WHERE user_id = $1 AND date = $2`

const sqlUpdateOccurrenceCompletion = `
UPDATE task_occurrences SET
	completed = $2,
	completed_at = $3,
	updated_at = $4
WHERE id = $1 AND user_id = $5
RETURNING` + occurrenceColumns

const sqlUpdateOccurrenceSortOrder = `
UPDATE task_occurrences SET sort_order = $3, updated_at = $4
WHERE id = $1 AND user_id = $2 AND date = $5`

const sqlBumpOccurrenceSortOrders = `
UPDATE task_occurrences SET sort_order = sort_order + $3, updated_at = $4
WHERE user_id = $1 AND date = $2`

const sqlUpdateOccurrenceDateAndSort = `
UPDATE task_occurrences SET date = $2, sort_order = $3, updated_at = $4
WHERE id = $1 AND user_id = $5
RETURNING` + occurrenceColumns

const sqlSummariesByUserDateRange = `
SELECT
	o.date,
	COUNT(*)::int AS total,
	COUNT(*) FILTER (WHERE o.completed)::int AS completed,
	COUNT(*) FILTER (WHERE NOT o.completed)::int AS pending,
	COUNT(*) FILTER (WHERE o.carried_forward)::int AS carried
FROM task_occurrences o
JOIN tasks t ON t.id = o.task_id AND t.deleted_at IS NULL
WHERE o.user_id = $1 AND o.date BETWEEN $2 AND $3
GROUP BY o.date
ORDER BY o.date ASC`

// TaskOccurrenceRepository persists daily journal rows.
type TaskOccurrenceRepository interface {
	Create(ctx context.Context, occ entity.TaskOccurrence) (entity.TaskOccurrence, error)
	CountByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) (int, error)
	ListByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]entity.TaskOccurrenceWithTask, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.TaskOccurrenceWithTask, error)
	ListIncompleteByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]entity.TaskOccurrence, error)
	MaxSortOrder(ctx context.Context, userID uuid.UUID, date time.Time) (int, error)
	UpdateCompletion(ctx context.Context, id, userID uuid.UUID, completed bool, completedAt *time.Time, updatedAt time.Time) (entity.TaskOccurrence, error)
	UpdateSortOrder(ctx context.Context, id, userID uuid.UUID, sortOrder int, date time.Time, updatedAt time.Time) error
	BumpSortOrders(ctx context.Context, userID uuid.UUID, date time.Time, delta int, updatedAt time.Time) error
	UpdateDateAndSort(ctx context.Context, id, userID uuid.UUID, date time.Time, sortOrder int, updatedAt time.Time) (entity.TaskOccurrence, error)
	SummariesByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TaskDaySummary, error)
	ExistsForTaskDate(ctx context.Context, taskID uuid.UUID, date time.Time) (bool, error)
	DeleteCarryForwardAfter(ctx context.Context, userID uuid.UUID, after time.Time) (int64, error)
	DeleteByTaskID(ctx context.Context, taskID, userID uuid.UUID) (int64, error)
	WithTx(tx pgx.Tx) TaskOccurrenceRepository
}

type taskOccurrenceRepository struct {
	q Querier
}

func NewTaskOccurrenceRepository(pool *pgxpool.Pool) TaskOccurrenceRepository {
	return &taskOccurrenceRepository{q: pool}
}

func (r *taskOccurrenceRepository) WithTx(tx pgx.Tx) TaskOccurrenceRepository {
	return &taskOccurrenceRepository{q: tx}
}

func (r *taskOccurrenceRepository) Create(ctx context.Context, occ entity.TaskOccurrence) (entity.TaskOccurrence, error) {
	row := r.q.QueryRow(ctx, sqlInsertTaskOccurrence,
		occ.ID, occ.PublicID, occ.TaskID, occ.UserID, civilDate(occ.Date),
		occ.SortOrder, occ.Completed, occ.CompletedAt, occ.CarriedForward,
		occ.Source, occ.CreatedAt, occ.UpdatedAt,
	)
	return scanOccurrence(row)
}

func (r *taskOccurrenceRepository) CountByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) (int, error) {
	var n int
	err := r.q.QueryRow(ctx, sqlCountOccurrencesByUserDate, userID, civilDate(date)).Scan(&n)
	return n, err
}

func (r *taskOccurrenceRepository) ListByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]entity.TaskOccurrenceWithTask, error) {
	rows, err := r.q.Query(ctx, sqlSelectOccurrencesByUserDate, userID, civilDate(date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOccurrencesWithTask(rows)
}

func (r *taskOccurrenceRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.TaskOccurrenceWithTask, error) {
	row := r.q.QueryRow(ctx, sqlSelectOccurrenceByID, id)
	return scanOccurrenceWithTask(row)
}

func (r *taskOccurrenceRepository) ListIncompleteByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]entity.TaskOccurrence, error) {
	rows, err := r.q.Query(ctx, sqlListIncompleteByUserDate, userID, civilDate(date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.TaskOccurrence, 0)
	for rows.Next() {
		occ, err := scanOccurrence(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, occ)
	}
	return out, rows.Err()
}

func (r *taskOccurrenceRepository) MaxSortOrder(ctx context.Context, userID uuid.UUID, date time.Time) (int, error) {
	var n int
	err := r.q.QueryRow(ctx, sqlMaxSortOrder, userID, civilDate(date)).Scan(&n)
	return n, err
}

func (r *taskOccurrenceRepository) UpdateCompletion(ctx context.Context, id, userID uuid.UUID, completed bool, completedAt *time.Time, updatedAt time.Time) (entity.TaskOccurrence, error) {
	row := r.q.QueryRow(ctx, sqlUpdateOccurrenceCompletion, id, completed, completedAt, updatedAt, userID)
	return scanOccurrence(row)
}

func (r *taskOccurrenceRepository) UpdateSortOrder(ctx context.Context, id, userID uuid.UUID, sortOrder int, date time.Time, updatedAt time.Time) error {
	_, err := r.q.Exec(ctx, sqlUpdateOccurrenceSortOrder, id, userID, sortOrder, updatedAt, civilDate(date))
	return err
}

func (r *taskOccurrenceRepository) BumpSortOrders(ctx context.Context, userID uuid.UUID, date time.Time, delta int, updatedAt time.Time) error {
	_, err := r.q.Exec(ctx, sqlBumpOccurrenceSortOrders, userID, civilDate(date), delta, updatedAt)
	return err
}

func (r *taskOccurrenceRepository) UpdateDateAndSort(ctx context.Context, id, userID uuid.UUID, date time.Time, sortOrder int, updatedAt time.Time) (entity.TaskOccurrence, error) {
	row := r.q.QueryRow(ctx, sqlUpdateOccurrenceDateAndSort, id, civilDate(date), sortOrder, updatedAt, userID)
	return scanOccurrence(row)
}

func (r *taskOccurrenceRepository) SummariesByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TaskDaySummary, error) {
	rows, err := r.q.Query(ctx, sqlSummariesByUserDateRange, userID, civilDate(from), civilDate(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.TaskDaySummary, 0)
	for rows.Next() {
		var s entity.TaskDaySummary
		if err := rows.Scan(&s.Date, &s.Total, &s.Completed, &s.Pending, &s.Carried); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *taskOccurrenceRepository) ExistsForTaskDate(ctx context.Context, taskID uuid.UUID, date time.Time) (bool, error) {
	var exists bool
	err := r.q.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM task_occurrences WHERE task_id = $1 AND date = $2)`,
		taskID, civilDate(date)).Scan(&exists)
	return exists, err
}

func (r *taskOccurrenceRepository) DeleteCarryForwardAfter(ctx context.Context, userID uuid.UUID, after time.Time) (int64, error) {
	tag, err := r.q.Exec(ctx, `
DELETE FROM task_occurrences
WHERE user_id = $1
  AND date > $2
  AND source = 'carry_forward'`,
		userID, civilDate(after))
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *taskOccurrenceRepository) DeleteByTaskID(ctx context.Context, taskID, userID uuid.UUID) (int64, error) {
	tag, err := r.q.Exec(ctx, `
DELETE FROM task_occurrences
WHERE task_id = $1 AND user_id = $2`,
		taskID, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func civilDate(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func scanOccurrence(row pgx.Row) (entity.TaskOccurrence, error) {
	var o entity.TaskOccurrence
	var date time.Time
	err := row.Scan(
		&o.ID, &o.PublicID, &o.TaskID, &o.UserID, &date, &o.SortOrder,
		&o.Completed, &o.CompletedAt, &o.CarriedForward, &o.Source,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TaskOccurrence{}, apperr.ErrNotFound
		}
		return entity.TaskOccurrence{}, fmt.Errorf("scan occurrence: %w", err)
	}
	o.Date = date
	return o, nil
}

func scanOccurrenceWithTask(row pgx.Row) (entity.TaskOccurrenceWithTask, error) {
	var o entity.TaskOccurrenceWithTask
	var date time.Time
	var labels []string
	err := row.Scan(
		&o.ID, &o.PublicID, &o.TaskID, &o.UserID, &date, &o.SortOrder,
		&o.Completed, &o.CompletedAt, &o.CarriedForward, &o.Source,
		&o.CreatedAt, &o.UpdatedAt,
		&o.Title, &o.Description, &o.Priority, &o.Project, &labels, &o.RecurrenceRule,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TaskOccurrenceWithTask{}, apperr.ErrNotFound
		}
		return entity.TaskOccurrenceWithTask{}, fmt.Errorf("scan occurrence with task: %w", err)
	}
	o.Date = date
	o.Labels = labels
	return o, nil
}

func collectOccurrencesWithTask(rows pgx.Rows) ([]entity.TaskOccurrenceWithTask, error) {
	out := make([]entity.TaskOccurrenceWithTask, 0)
	for rows.Next() {
		var o entity.TaskOccurrenceWithTask
		var date time.Time
		var labels []string
		if err := rows.Scan(
			&o.ID, &o.PublicID, &o.TaskID, &o.UserID, &date, &o.SortOrder,
			&o.Completed, &o.CompletedAt, &o.CarriedForward, &o.Source,
			&o.CreatedAt, &o.UpdatedAt,
			&o.Title, &o.Description, &o.Priority, &o.Project, &labels, &o.RecurrenceRule,
		); err != nil {
			return nil, err
		}
		o.Date = date
		o.Labels = labels
		out = append(out, o)
	}
	return out, rows.Err()
}
