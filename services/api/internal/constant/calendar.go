package constant

import "time"

// Calendar sync strategy knobs (also stored on scheduler_jobs.payload for background cadence).
const (
	// CalendarSyncStaleAfter triggers on-demand / startup incremental sync when last success is older.
	CalendarSyncStaleAfter = 2 * time.Minute

	// CalendarSyncIntervalMinutes is written into scheduler_jobs.payload (not a cron expression).
	CalendarSyncIntervalMinutes = 15

	CalendarSyncStatusIdle      = "idle"
	CalendarSyncStatusRunning   = "running"
	CalendarSyncStatusSucceeded = "succeeded"
	CalendarSyncStatusFailed    = "failed"
)
