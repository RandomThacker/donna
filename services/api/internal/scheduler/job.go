package scheduler

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// Job is a pluggable unit of background work backed by scheduler_jobs.
// Calendar, Gmail, Contacts, Tasks, and reminders each implement this once
// and register with the platform Runner — the scheduler stays integration-agnostic.
type Job interface {
	// Type returns the scheduler_jobs.job_type this handler owns.
	Type() string
	// Run executes one claimed job. The Runner records success/failure.
	Run(ctx context.Context, job entity.SchedulerJob) error
}

// RecurringJob optionally enqueues the next run after a job finishes
// (success or failure). Cadence belongs in scheduler_jobs.payload, not cron code.
type RecurringJob interface {
	Job
	ScheduleNext(ctx context.Context, job entity.SchedulerJob) error
}
