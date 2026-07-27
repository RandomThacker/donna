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

const dailyNoteColumns = `id, public_id, user_id, date, content, created_at, updated_at`

const sqlUpsertDailyNote = `
INSERT INTO daily_notes (id, public_id, user_id, date, content, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id, date) DO UPDATE SET
	content = EXCLUDED.content,
	updated_at = EXCLUDED.updated_at
RETURNING ` + dailyNoteColumns

const sqlSelectDailyNote = `
SELECT ` + dailyNoteColumns + `
FROM daily_notes
WHERE user_id = $1 AND date = $2`

// DailyNoteRepository persists per-day markdown notes.
type DailyNoteRepository interface {
	GetByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) (entity.DailyNote, error)
	Upsert(ctx context.Context, note entity.DailyNote) (entity.DailyNote, error)
	WithTx(tx pgx.Tx) DailyNoteRepository
}

type dailyNoteRepository struct {
	q Querier
}

func NewDailyNoteRepository(pool *pgxpool.Pool) DailyNoteRepository {
	return &dailyNoteRepository{q: pool}
}

func (r *dailyNoteRepository) WithTx(tx pgx.Tx) DailyNoteRepository {
	return &dailyNoteRepository{q: tx}
}

func (r *dailyNoteRepository) GetByUserDate(ctx context.Context, userID uuid.UUID, date time.Time) (entity.DailyNote, error) {
	row := r.q.QueryRow(ctx, sqlSelectDailyNote, userID, civilDate(date))
	return scanDailyNote(row)
}

func (r *dailyNoteRepository) Upsert(ctx context.Context, note entity.DailyNote) (entity.DailyNote, error) {
	row := r.q.QueryRow(ctx, sqlUpsertDailyNote,
		note.ID, note.PublicID, note.UserID, civilDate(note.Date),
		note.Content, note.CreatedAt, note.UpdatedAt,
	)
	return scanDailyNote(row)
}

func scanDailyNote(row pgx.Row) (entity.DailyNote, error) {
	var n entity.DailyNote
	var date time.Time
	err := row.Scan(&n.ID, &n.PublicID, &n.UserID, &date, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.DailyNote{}, apperr.ErrNotFound
		}
		return entity.DailyNote{}, fmt.Errorf("scan daily note: %w", err)
	}
	n.Date = date
	return n, nil
}
