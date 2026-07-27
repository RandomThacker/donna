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

const taskColumns = `
	id, public_id, user_id, title, description, status, priority, project, labels,
	due_at, completed_at, is_backlog, recurrence_rule, provider, provider_task_id,
	provider_payload, created_at, updated_at, deleted_at`

const sqlInsertTask = `
INSERT INTO tasks (
	id, public_id, user_id, title, description, status, priority, project, labels,
	due_at, completed_at, is_backlog, recurrence_rule, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
)
RETURNING` + taskColumns

const sqlSelectTaskByID = `
SELECT` + taskColumns + `
FROM tasks
WHERE id = $1 AND deleted_at IS NULL`

const sqlUpdateTask = `
UPDATE tasks SET
	title = COALESCE($2, title),
	description = COALESCE($3, description),
	priority = COALESCE($4, priority),
	project = COALESCE($5, project),
	labels = COALESCE($6, labels),
	recurrence_rule = COALESCE($7, recurrence_rule),
	updated_at = $8
WHERE id = $1 AND user_id = $9 AND deleted_at IS NULL
RETURNING` + taskColumns

const sqlSoftDeleteTask = `
UPDATE tasks SET deleted_at = $3, updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

// TaskRepository persists permanent tasks.
type TaskRepository interface {
	Create(ctx context.Context, task entity.Task) (entity.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Task, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields TaskUpdateFields, updatedAt time.Time) (entity.Task, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error
	WithTx(tx pgx.Tx) TaskRepository
}

type TaskUpdateFields struct {
	Title          *string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	RecurrenceRule *string
}

type taskRepository struct {
	q Querier
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &taskRepository{q: pool}
}

func (r *taskRepository) WithTx(tx pgx.Tx) TaskRepository {
	return &taskRepository{q: tx}
}

func (r *taskRepository) Create(ctx context.Context, task entity.Task) (entity.Task, error) {
	labels := task.Labels
	if labels == nil {
		labels = []string{}
	}
	row := r.q.QueryRow(ctx, sqlInsertTask,
		task.ID, task.PublicID, task.UserID, task.Title, task.Description,
		task.Status, task.Priority, task.Project, labels,
		task.DueAt, task.CompletedAt, task.IsBacklog, task.RecurrenceRule,
		task.CreatedAt, task.UpdatedAt,
	)
	return scanTask(row)
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Task, error) {
	row := r.q.QueryRow(ctx, sqlSelectTaskByID, id)
	task, err := scanTask(row)
	if errors.Is(err, apperr.ErrNotFound) {
		return entity.Task{}, err
	}
	return task, err
}

func (r *taskRepository) Update(ctx context.Context, id, userID uuid.UUID, fields TaskUpdateFields, updatedAt time.Time) (entity.Task, error) {
	row := r.q.QueryRow(ctx, sqlUpdateTask,
		id, fields.Title, fields.Description, fields.Priority, fields.Project,
		fields.Labels, fields.RecurrenceRule, updatedAt, userID,
	)
	return scanTask(row)
}

func (r *taskRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteTask, id, userID, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func scanTask(row pgx.Row) (entity.Task, error) {
	var t entity.Task
	var labels []string
	err := row.Scan(
		&t.ID, &t.PublicID, &t.UserID, &t.Title, &t.Description, &t.Status,
		&t.Priority, &t.Project, &labels, &t.DueAt, &t.CompletedAt,
		&t.IsBacklog, &t.RecurrenceRule, &t.Provider, &t.ProviderTaskID,
		&t.ProviderPayload, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Task{}, apperr.ErrNotFound
		}
		return entity.Task{}, fmt.Errorf("scan task: %w", err)
	}
	t.Labels = labels
	return t, nil
}
