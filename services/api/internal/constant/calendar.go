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

	// CalendarEventSyncLookback / Lookahead bound the initial full events.list window.
	CalendarEventSyncLookback  = 365 * 24 * time.Hour
	CalendarEventSyncLookahead = 730 * 24 * time.Hour

	CalendarEventOriginDonna        = "donna"
	CalendarEventOriginProviderSync = "provider_sync"
	CalendarEventStatusConfirmed    = "confirmed"
	CalendarEventStatusTentative    = "tentative"
	CalendarEventStatusCancelled    = "cancelled"

	// CalendarProviderCalendarIDDonna is the virtual source for Donna-owned events.
	CalendarProviderCalendarIDDonna = "donna_local"
	CalendarDonnaSourceName         = "Donna"

	CalendarSyncTriggerManual    = "manual"
	CalendarSyncTriggerEnsure    = "ensure"
	CalendarSyncTriggerScheduler = "scheduler"

	CalendarSyncRunStatusRunning   = "running"
	CalendarSyncRunStatusSucceeded = "succeeded"
	CalendarSyncRunStatusPartial   = "partial"
	CalendarSyncRunStatusFailed    = "failed"
	CalendarSyncRunStatusSkipped   = "skipped"
)
