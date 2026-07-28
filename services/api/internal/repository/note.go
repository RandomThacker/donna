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

const noteColumns = `
	id, public_id, user_id, title, content, color, pinned, archived,
	created_at, updated_at, deleted_at`

const sqlInsertNote = `
INSERT INTO notes (
	id, public_id, user_id, title, content, color, pinned, archived,
	created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING` + noteColumns

const sqlSelectNoteByID = `
SELECT` + noteColumns + `
FROM notes
WHERE id = $1 AND deleted_at IS NULL`

const sqlListNotesByUser = `
SELECT` + noteColumns + `
FROM notes
WHERE user_id = $1 AND deleted_at IS NULL AND archived = false
ORDER BY pinned DESC, updated_at DESC`

const sqlUpdateNote = `
UPDATE notes SET
	title = COALESCE($2, title),
	content = COALESCE($3, content),
	color = COALESCE($4, color),
	pinned = COALESCE($5, pinned),
	archived = COALESCE($6, archived),
	updated_at = $7
WHERE id = $1 AND user_id = $8 AND deleted_at IS NULL
RETURNING` + noteColumns

const sqlSoftDeleteNote = `
UPDATE notes SET deleted_at = $3, updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

// NoteRepository persists Keep-style notes.
type NoteRepository interface {
	Create(ctx context.Context, note entity.Note) (entity.Note, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Note, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Note, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields NoteUpdateFields, updatedAt time.Time) (entity.Note, error)
	SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error
	WithTx(tx pgx.Tx) NoteRepository
}

// NoteUpdateFields are optional patches for a note.
type NoteUpdateFields struct {
	Title    *string
	Content  *string
	Color    *string
	Pinned   *bool
	Archived *bool
}

type noteRepository struct {
	q Querier
}

func NewNoteRepository(pool *pgxpool.Pool) NoteRepository {
	return &noteRepository{q: pool}
}

func (r *noteRepository) WithTx(tx pgx.Tx) NoteRepository {
	return &noteRepository{q: tx}
}

func (r *noteRepository) Create(ctx context.Context, note entity.Note) (entity.Note, error) {
	row := r.q.QueryRow(ctx, sqlInsertNote,
		note.ID, note.PublicID, note.UserID, note.Title, note.Content,
		note.Color, note.Pinned, note.Archived, note.CreatedAt, note.UpdatedAt,
	)
	return scanNote(row)
}

func (r *noteRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Note, error) {
	return scanNote(r.q.QueryRow(ctx, sqlSelectNoteByID, id))
}

func (r *noteRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Note, error) {
	rows, err := r.q.Query(ctx, sqlListNotesByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	return out, rows.Err()
}

func (r *noteRepository) Update(ctx context.Context, id, userID uuid.UUID, fields NoteUpdateFields, updatedAt time.Time) (entity.Note, error) {
	return scanNote(r.q.QueryRow(ctx, sqlUpdateNote,
		id, fields.Title, fields.Content, fields.Color, fields.Pinned, fields.Archived,
		updatedAt, userID,
	))
}

func (r *noteRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	tag, err := r.q.Exec(ctx, sqlSoftDeleteNote, id, userID, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func scanNote(row pgx.Row) (entity.Note, error) {
	var n entity.Note
	err := row.Scan(
		&n.ID, &n.PublicID, &n.UserID, &n.Title, &n.Content, &n.Color,
		&n.Pinned, &n.Archived, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Note{}, apperr.ErrNotFound
		}
		return entity.Note{}, fmt.Errorf("scan note: %w", err)
	}
	return n, nil
}
