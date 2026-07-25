package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schedulerJobColumns = `
	id, public_id, user_id, job_type, status, run_at, attempt_count, max_attempts, payload,
	reminder_id, connected_account_id, last_error, started_at, finished_at, created_at, updated_at`

const (
	sqlInsertSchedulerJob = `
INSERT INTO scheduler_jobs (
	id, public_id, user_id, job_type, status, run_at, attempt_count, max_attempts, payload,
	reminder_id, connected_account_id, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13
)
RETURNING` + schedulerJobColumns

	sqlSelectDueJobs = `
SELECT` + schedulerJobColumns + `
FROM scheduler_jobs
WHERE job_type = ANY($1::text[])
  AND status = 'pending'
  AND run_at <= $2
ORDER BY run_at ASC
LIMIT $3`

	sqlSelectPendingByTypeAndAccount = `
SELECT` + schedulerJobColumns + `
FROM scheduler_jobs
WHERE job_type = $1
  AND status = 'pending'
  AND connected_account_id = $2
ORDER BY run_at ASC
LIMIT 1`

	sqlClaimSchedulerJob = `
UPDATE scheduler_jobs SET
	status = 'running',
	attempt_count = attempt_count + 1,
	started_at = $2,
	updated_at = $2
WHERE id = $1 AND status = 'pending'
RETURNING` + schedulerJobColumns

	sqlFinishSchedulerJob = `
UPDATE scheduler_jobs SET
	status = $2,
	last_error = $3,
	finished_at = $4,
	updated_at = $4
WHERE id = $1
RETURNING` + schedulerJobColumns

	sqlReschedulePendingJob = `
UPDATE scheduler_jobs SET
	run_at = $2,
	updated_at = $3
WHERE id = $1 AND status = 'pending'
RETURNING` + schedulerJobColumns
)

// SchedulerJobRepository persists durable jobs for the platform scheduler.
type SchedulerJobRepository interface {
	Create(ctx context.Context, job entity.SchedulerJob) (entity.SchedulerJob, error)
	ListDue(ctx context.Context, jobTypes []string, asOf time.Time, limit int) ([]entity.SchedulerJob, error)
	GetPendingByTypeAndAccount(ctx context.Context, jobType string, accountID uuid.UUID) (entity.SchedulerJob, error)
	Claim(ctx context.Context, id uuid.UUID, startedAt time.Time) (entity.SchedulerJob, error)
	Finish(ctx context.Context, id uuid.UUID, status string, lastError *string, finishedAt time.Time) (entity.SchedulerJob, error)
	ReschedulePending(ctx context.Context, id uuid.UUID, runAt, updatedAt time.Time) (entity.SchedulerJob, error)
	WithTx(tx pgx.Tx) SchedulerJobRepository
}

type schedulerJobRepository struct {
	q Querier
}

// NewSchedulerJobRepository constructs a SchedulerJobRepository.
func NewSchedulerJobRepository(pool *pgxpool.Pool) SchedulerJobRepository {
	return &schedulerJobRepository{q: pool}
}

func (r *schedulerJobRepository) WithTx(tx pgx.Tx) SchedulerJobRepository {
	return &schedulerJobRepository{q: tx}
}

func (r *schedulerJobRepository) Create(ctx context.Context, job entity.SchedulerJob) (entity.SchedulerJob, error) {
	payload := job.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	created, err := scanSchedulerJob(r.q.QueryRow(ctx, sqlInsertSchedulerJob,
		job.ID,
		job.PublicID,
		job.UserID,
		job.JobType,
		job.Status,
		job.RunAt,
		job.AttemptCount,
		job.MaxAttempts,
		payload,
		job.ReminderID,
		job.ConnectedAccountID,
		job.CreatedAt,
		job.UpdatedAt,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return entity.SchedulerJob{}, apperr.ErrConflict
		}
		return entity.SchedulerJob{}, fmt.Errorf("insert scheduler job: %w", err)
	}
	return created, nil
}

func (r *schedulerJobRepository) ListDue(ctx context.Context, jobTypes []string, asOf time.Time, limit int) ([]entity.SchedulerJob, error) {
	if len(jobTypes) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.q.Query(ctx, sqlSelectDueJobs, jobTypes, asOf, limit)
	if err != nil {
		return nil, fmt.Errorf("list due scheduler jobs: %w", err)
	}
	defer rows.Close()
	out := make([]entity.SchedulerJob, 0)
	for rows.Next() {
		job, err := scanSchedulerJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, job)
	}
	return out, rows.Err()
}

func (r *schedulerJobRepository) GetPendingByTypeAndAccount(ctx context.Context, jobType string, accountID uuid.UUID) (entity.SchedulerJob, error) {
	job, err := scanSchedulerJob(r.q.QueryRow(ctx, sqlSelectPendingByTypeAndAccount, jobType, accountID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.SchedulerJob{}, apperr.ErrNotFound
		}
		return entity.SchedulerJob{}, fmt.Errorf("get pending scheduler job: %w", err)
	}
	return job, nil
}

func (r *schedulerJobRepository) Claim(ctx context.Context, id uuid.UUID, startedAt time.Time) (entity.SchedulerJob, error) {
	job, err := scanSchedulerJob(r.q.QueryRow(ctx, sqlClaimSchedulerJob, id, startedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.SchedulerJob{}, apperr.ErrNotFound
		}
		return entity.SchedulerJob{}, fmt.Errorf("claim scheduler job: %w", err)
	}
	return job, nil
}

func (r *schedulerJobRepository) Finish(ctx context.Context, id uuid.UUID, status string, lastError *string, finishedAt time.Time) (entity.SchedulerJob, error) {
	job, err := scanSchedulerJob(r.q.QueryRow(ctx, sqlFinishSchedulerJob, id, status, lastError, finishedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.SchedulerJob{}, apperr.ErrNotFound
		}
		return entity.SchedulerJob{}, fmt.Errorf("finish scheduler job: %w", err)
	}
	return job, nil
}

func (r *schedulerJobRepository) ReschedulePending(ctx context.Context, id uuid.UUID, runAt, updatedAt time.Time) (entity.SchedulerJob, error) {
	job, err := scanSchedulerJob(r.q.QueryRow(ctx, sqlReschedulePendingJob, id, runAt, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.SchedulerJob{}, apperr.ErrNotFound
		}
		return entity.SchedulerJob{}, fmt.Errorf("reschedule scheduler job: %w", err)
	}
	return job, nil
}

// CalendarSyncJobPayload is stored in scheduler_jobs.payload.
type CalendarSyncJobPayload struct {
	IntervalMinutes int    `json:"interval_minutes"`
	SyncKind        string `json:"sync_kind"`
}

// EncodeCalendarSyncPayload marshals the background sync cadence config.
func EncodeCalendarSyncPayload(intervalMinutes int) ([]byte, error) {
	if intervalMinutes <= 0 {
		intervalMinutes = constant.CalendarSyncIntervalMinutes
	}
	return json.Marshal(CalendarSyncJobPayload{
		IntervalMinutes: intervalMinutes,
		SyncKind:        "calendar_sources",
	})
}

// DecodeCalendarSyncPayload reads cadence from a job payload.
func DecodeCalendarSyncPayload(raw []byte) CalendarSyncJobPayload {
	var p CalendarSyncJobPayload
	_ = json.Unmarshal(raw, &p)
	if p.IntervalMinutes <= 0 {
		p.IntervalMinutes = constant.CalendarSyncIntervalMinutes
	}
	if p.SyncKind == "" {
		p.SyncKind = "calendar_sources"
	}
	return p
}

func scanSchedulerJob(row scannable) (entity.SchedulerJob, error) {
	var j entity.SchedulerJob
	err := row.Scan(
		&j.ID,
		&j.PublicID,
		&j.UserID,
		&j.JobType,
		&j.Status,
		&j.RunAt,
		&j.AttemptCount,
		&j.MaxAttempts,
		&j.Payload,
		&j.ReminderID,
		&j.ConnectedAccountID,
		&j.LastError,
		&j.StartedAt,
		&j.FinishedAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	return j, err
}
