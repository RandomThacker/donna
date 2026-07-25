-- Domain completion: scheduler_jobs.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § scheduler_jobs
-- No soft delete — retention via hard delete. For reminder_fire, run_at copies remind_at (R-03).

CREATE TABLE IF NOT EXISTS scheduler_jobs (
    id                     uuid        NOT NULL,
    public_id              text        NOT NULL,
    user_id                uuid        NOT NULL,
    job_type               text        NOT NULL,
    status                 text        NOT NULL DEFAULT 'pending',
    run_at                 timestamptz NOT NULL,
    attempt_count          integer     NOT NULL DEFAULT 0,
    max_attempts           integer     NOT NULL DEFAULT 5,
    payload                jsonb       NOT NULL DEFAULT '{}'::jsonb,
    reminder_id            uuid        NULL,
    connected_account_id   uuid        NULL,
    last_error             text        NULL,
    started_at             timestamptz NULL,
    finished_at            timestamptz NULL,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT scheduler_jobs_pkey PRIMARY KEY (id),
    CONSTRAINT scheduler_jobs_public_id_key UNIQUE (public_id),
    CONSTRAINT scheduler_jobs_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT scheduler_jobs_reminder_id_fkey
        FOREIGN KEY (reminder_id) REFERENCES reminders (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT scheduler_jobs_connected_account_id_fkey
        FOREIGN KEY (connected_account_id) REFERENCES connected_accounts (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT scheduler_jobs_public_id_prefix_check CHECK (public_id LIKE 'job_%'),
    CONSTRAINT scheduler_jobs_job_type_check
        CHECK (job_type IN (
            'morning_briefing',
            'midday_checkin',
            'evening_reflection',
            'reminder_fire',
            'calendar_sync'
        )),
    CONSTRAINT scheduler_jobs_status_check
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    CONSTRAINT scheduler_jobs_attempt_count_check CHECK (attempt_count >= 0),
    CONSTRAINT scheduler_jobs_max_attempts_check CHECK (max_attempts >= 1),
    CONSTRAINT scheduler_jobs_payload_object_check
        CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX IF NOT EXISTS scheduler_jobs_pending_run_at_idx
    ON scheduler_jobs (run_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS scheduler_jobs_user_id_job_type_status_idx
    ON scheduler_jobs (user_id, job_type, status);

CREATE INDEX IF NOT EXISTS scheduler_jobs_reminder_id_idx
    ON scheduler_jobs (reminder_id);

COMMENT ON TABLE scheduler_jobs IS 'Durable job ledger; no soft delete — purge by retention policy.';
COMMENT ON COLUMN scheduler_jobs.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN scheduler_jobs.public_id IS 'Stable API identifier with job_ prefix.';
COMMENT ON COLUMN scheduler_jobs.user_id IS 'Owning Donna user scope.';
COMMENT ON COLUMN scheduler_jobs.job_type IS 'morning_briefing | midday_checkin | evening_reflection | reminder_fire | calendar_sync.';
COMMENT ON COLUMN scheduler_jobs.status IS 'pending | running | succeeded | failed | cancelled.';
COMMENT ON COLUMN scheduler_jobs.run_at IS 'Due time (UTC); for reminder_fire equals reminders.remind_at at creation.';
COMMENT ON COLUMN scheduler_jobs.attempt_count IS 'Number of execution attempts so far.';
COMMENT ON COLUMN scheduler_jobs.max_attempts IS 'Maximum attempts before terminal failure.';
COMMENT ON COLUMN scheduler_jobs.payload IS 'Extras not covered by FKs (jsonb object); never the only copy of reminder_id.';
COMMENT ON COLUMN scheduler_jobs.reminder_id IS 'Optional reminder correlation.';
COMMENT ON COLUMN scheduler_jobs.connected_account_id IS 'Optional account for sync jobs.';
COMMENT ON COLUMN scheduler_jobs.last_error IS 'Last failure detail.';
COMMENT ON COLUMN scheduler_jobs.started_at IS 'When the current/last attempt started.';
COMMENT ON COLUMN scheduler_jobs.finished_at IS 'When the job reached a terminal state.';
COMMENT ON COLUMN scheduler_jobs.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN scheduler_jobs.updated_at IS 'Last mutation time; maintained by application.';
