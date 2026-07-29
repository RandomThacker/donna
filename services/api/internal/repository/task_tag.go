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

const taskTagColumns = `
	id, public_id, user_id, name, color, created_at, updated_at`

const sqlInsertTaskTag = `
INSERT INTO task_tags (
	id, public_id, user_id, name, color, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING` + taskTagColumns

const sqlSelectTaskTagByID = `
SELECT` + taskTagColumns + `
FROM task_tags
WHERE id = $1`

const sqlListTaskTagsByUser = `
SELECT` + taskTagColumns + `
FROM task_tags
WHERE user_id = $1
ORDER BY lower(name) ASC`

const sqlUpdateTaskTag = `
UPDATE task_tags SET
	name = COALESCE($2, name),
	color = COALESCE($3, color),
	updated_at = $4
WHERE id = $1 AND user_id = $5
RETURNING` + taskTagColumns

const sqlDeleteTaskTag = `
DELETE FROM task_tags
WHERE id = $1 AND user_id = $2`

const sqlDeleteTaskTagAssignmentsForTask = `
DELETE FROM task_tag_assignments
WHERE task_id = $1`

const sqlReplaceTaskTagAssignments = `
DELETE FROM task_tag_assignments
WHERE task_id = $1`

const sqlInsertTaskTagAssignment = `
INSERT INTO task_tag_assignments (task_id, tag_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`

const sqlListTaskTagsByTaskIDs = `
SELECT tta.task_id, tt.id, tt.public_id, tt.user_id, tt.name, tt.color, tt.created_at, tt.updated_at
FROM task_tag_assignments tta
JOIN task_tags tt ON tt.id = tta.tag_id
WHERE tta.task_id = ANY($1) AND tt.user_id = $2
ORDER BY lower(tt.name) ASC`

// TaskTagRepository persists user-defined task tags.
type TaskTagRepository interface {
	Create(ctx context.Context, tag entity.TaskTag) (entity.TaskTag, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.TaskTag, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.TaskTag, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields TaskTagUpdateFields, updatedAt time.Time) (entity.TaskTag, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ReplaceTaskTags(ctx context.Context, taskID uuid.UUID, tagIDs []uuid.UUID, now time.Time) error
	RemoveAllForTask(ctx context.Context, taskID uuid.UUID) error
	ListByTaskIDs(ctx context.Context, userID uuid.UUID, taskIDs []uuid.UUID) (map[uuid.UUID][]entity.TaskTag, error)
	WithTx(tx pgx.Tx) TaskTagRepository
}

// TaskTagUpdateFields are optional patches for a tag.
type TaskTagUpdateFields struct {
	Name  *string
	Color *string
}

type taskTagRepository struct {
	q Querier
}

func NewTaskTagRepository(pool *pgxpool.Pool) TaskTagRepository {
	return &taskTagRepository{q: pool}
}

func (r *taskTagRepository) WithTx(tx pgx.Tx) TaskTagRepository {
	return &taskTagRepository{q: tx}
}

func (r *taskTagRepository) Create(ctx context.Context, tag entity.TaskTag) (entity.TaskTag, error) {
	row := r.q.QueryRow(ctx, sqlInsertTaskTag,
		tag.ID, tag.PublicID, tag.UserID, tag.Name, tag.Color, tag.CreatedAt, tag.UpdatedAt,
	)
	return scanTaskTag(row)
}

func (r *taskTagRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.TaskTag, error) {
	row := r.q.QueryRow(ctx, sqlSelectTaskTagByID, id)
	return scanTaskTag(row)
}

func (r *taskTagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.TaskTag, error) {
	rows, err := r.q.Query(ctx, sqlListTaskTagsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.TaskTag, 0)
	for rows.Next() {
		tag, err := scanTaskTag(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, tag)
	}
	return out, rows.Err()
}

func (r *taskTagRepository) Update(ctx context.Context, id, userID uuid.UUID, fields TaskTagUpdateFields, updatedAt time.Time) (entity.TaskTag, error) {
	row := r.q.QueryRow(ctx, sqlUpdateTaskTag, id, fields.Name, fields.Color, updatedAt, userID)
	return scanTaskTag(row)
}

func (r *taskTagRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tag.UserID != userID {
		return apperr.ErrForbidden
	}
	ct, err := r.q.Exec(ctx, sqlDeleteTaskTag, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *taskTagRepository) RemoveAllForTask(ctx context.Context, taskID uuid.UUID) error {
	_, err := r.q.Exec(ctx, sqlDeleteTaskTagAssignmentsForTask, taskID)
	return err
}

func (r *taskTagRepository) ReplaceTaskTags(ctx context.Context, taskID uuid.UUID, tagIDs []uuid.UUID, now time.Time) error {
	if _, err := r.q.Exec(ctx, sqlReplaceTaskTagAssignments, taskID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := r.q.Exec(ctx, sqlInsertTaskTagAssignment, taskID, tagID, now); err != nil {
			return err
		}
	}
	return nil
}

func (r *taskTagRepository) ListByTaskIDs(ctx context.Context, userID uuid.UUID, taskIDs []uuid.UUID) (map[uuid.UUID][]entity.TaskTag, error) {
	out := make(map[uuid.UUID][]entity.TaskTag)
	if len(taskIDs) == 0 {
		return out, nil
	}
	rows, err := r.q.Query(ctx, sqlListTaskTagsByTaskIDs, taskIDs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID uuid.UUID
		var tag entity.TaskTag
		if err := rows.Scan(
			&taskID,
			&tag.ID, &tag.PublicID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out[taskID] = append(out[taskID], tag)
	}
	return out, rows.Err()
}

func scanTaskTag(row pgx.Row) (entity.TaskTag, error) {
	var tag entity.TaskTag
	err := row.Scan(
		&tag.ID, &tag.PublicID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TaskTag{}, apperr.ErrNotFound
		}
		return entity.TaskTag{}, fmt.Errorf("scan task tag: %w", err)
	}
	return tag, nil
}
