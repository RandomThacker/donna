-- Append-only history of calendar sync pipeline attempts (sources + events).
-- connected_accounts keeps last-state snapshot; this table is per-run observability.

CREATE TABLE IF NOT EXISTS calendar_sync_runs (
    id                      uuid        NOT NULL,
    public_id               text        NOT NULL,
    user_id                 uuid        NOT NULL,
    connected_account_id    uuid        NOT NULL,
    trigger                 text        NOT NULL,
    status                  text        NOT NULL,
    started_at              timestamptz NOT NULL,
    finished_at             timestamptz NULL,
    duration_ms             integer     NULL,
    calendars_processed     integer     NOT NULL DEFAULT 0,
    sources_created         integer     NOT NULL DEFAULT 0,
    sources_updated         integer     NOT NULL DEFAULT 0,
    sources_deleted         integer     NOT NULL DEFAULT 0,
    events_created          integer     NOT NULL DEFAULT 0,
    events_updated          integer     NOT NULL DEFAULT 0,
    events_deleted          integer     NOT NULL DEFAULT 0,
    failures                jsonb       NOT NULL DEFAULT '[]'::jsonb,
    created_at              timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT calendar_sync_runs_pkey PRIMARY KEY (id),
    CONSTRAINT calendar_sync_runs_public_id_key UNIQUE (public_id),
    CONSTRAINT calendar_sync_runs_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_sync_runs_connected_account_id_fkey
        FOREIGN KEY (connected_account_id) REFERENCES connected_accounts (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_sync_runs_public_id_prefix_check CHECK (public_id LIKE 'csync_%'),
    CONSTRAINT calendar_sync_runs_trigger_check
        CHECK (trigger IN ('manual', 'ensure', 'scheduler')),
    CONSTRAINT calendar_sync_runs_status_check
        CHECK (status IN ('running', 'succeeded', 'partial', 'failed', 'skipped')),
    CONSTRAINT calendar_sync_runs_failures_array_check
        CHECK (jsonb_typeof(failures) = 'array'),
    CONSTRAINT calendar_sync_runs_duration_nonneg_check
        CHECK (duration_ms IS NULL OR duration_ms >= 0),
    CONSTRAINT calendar_sync_runs_counts_nonneg_check
        CHECK (
            calendars_processed >= 0
            AND sources_created >= 0 AND sources_updated >= 0 AND sources_deleted >= 0
            AND events_created >= 0 AND events_updated >= 0 AND events_deleted >= 0
        )
);

CREATE INDEX IF NOT EXISTS calendar_sync_runs_account_started_idx
    ON calendar_sync_runs (connected_account_id, started_at DESC);

CREATE INDEX IF NOT EXISTS calendar_sync_runs_user_started_idx
    ON calendar_sync_runs (user_id, started_at DESC);

COMMENT ON TABLE calendar_sync_runs IS 'Per-attempt calendar sync pipeline observability (sources + events).';
COMMENT ON COLUMN calendar_sync_runs.trigger IS 'manual | ensure | scheduler.';
COMMENT ON COLUMN calendar_sync_runs.status IS 'running | succeeded | partial | failed | skipped.';
COMMENT ON COLUMN calendar_sync_runs.failures IS 'Per-calendar failure objects (jsonb array).';
COMMENT ON COLUMN calendar_sync_runs.calendars_processed IS 'Enabled calendar_sources attempted for events sync.';
COMMENT ON COLUMN calendar_sync_runs.events_created IS 'Events created this run.';
COMMENT ON COLUMN calendar_sync_runs.events_updated IS 'Events updated this run.';
COMMENT ON COLUMN calendar_sync_runs.events_deleted IS 'Events soft-deleted this run.';
