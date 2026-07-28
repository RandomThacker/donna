package calendarsync

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/scheduler"
)

// Job runs incremental calendar source sync for a connected account.
// Implements scheduler.Job and scheduler.RecurringJob so cadence stays in
// scheduler_jobs.payload and future integrations share the same Runner.
type Job struct {
	calendar *business.CalendarService
}

// NewJob constructs a CalendarSync Job handler.
func NewJob(calendar *business.CalendarService) *Job {
	return &Job{calendar: calendar}
}

var (
	_ scheduler.Job          = (*Job)(nil)
	_ scheduler.RecurringJob = (*Job)(nil)
)

// Type returns scheduler_jobs.job_type for calendar sync.
func (j *Job) Type() string {
	return constant.SchedulerJobTypeCalendarSync
}

// Run syncs calendar sources and events for the job's connected account.
func (j *Job) Run(ctx context.Context, job entity.SchedulerJob) error {
	if j.calendar == nil {
		return fmt.Errorf("calendar service is not configured")
	}
	if job.ConnectedAccountID == nil {
		return fmt.Errorf("missing connected_account_id")
	}
	result, err := j.calendar.SyncSourcesForAccount(ctx, *job.ConnectedAccountID)
	if err != nil {
		return err
	}
	// Partial success (some calendars failed) still completes the job; failures are persisted.
	if result.Status == constant.CalendarSyncRunStatusFailed {
		return fmt.Errorf("calendar sync failed")
	}
	return nil
}

// ScheduleNext enqueues the next pending calendar_sync job from payload cadence.
func (j *Job) ScheduleNext(ctx context.Context, job entity.SchedulerJob) error {
	if j.calendar == nil {
		return fmt.Errorf("calendar service is not configured")
	}
	if job.ConnectedAccountID == nil {
		return fmt.Errorf("missing connected_account_id")
	}
	return j.calendar.EnsureBackgroundSyncJobWithPayload(ctx, *job.ConnectedAccountID, job.Payload)
}
