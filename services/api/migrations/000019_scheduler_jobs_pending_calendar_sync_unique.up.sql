-- One pending calendar_sync job per connected account (prevents concurrent ensure leaks).
-- Cadence remains in scheduler_jobs.payload (interval_minutes), not cron constants.

CREATE UNIQUE INDEX IF NOT EXISTS scheduler_jobs_pending_calendar_sync_account_uidx
    ON scheduler_jobs (connected_account_id)
    WHERE status = 'pending'
      AND job_type = 'calendar_sync'
      AND connected_account_id IS NOT NULL;

COMMENT ON INDEX scheduler_jobs_pending_calendar_sync_account_uidx IS
    'At most one pending calendar_sync job per connected account.';
