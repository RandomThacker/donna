package constant

// Scheduler job ledger statuses and shared job-type identifiers.
const (
	SchedulerJobStatusPending   = "pending"
	SchedulerJobStatusRunning   = "running"
	SchedulerJobStatusSucceeded = "succeeded"
	SchedulerJobStatusFailed    = "failed"

	PublicIDPrefixSchedulerJob = "job_"

	// Job types registered with the platform scheduler (extend as integrations land).
	SchedulerJobTypeCalendarSync = "calendar_sync"
)
