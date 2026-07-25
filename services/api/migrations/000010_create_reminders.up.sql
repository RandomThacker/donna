-- Domain completion: reminders.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § reminders
-- remind_at is schedule source of truth (R-03). Write ownership (R-02): reminder service.

CREATE TABLE IF NOT EXISTS reminders (
    id                   uuid        NOT NULL,
    public_id            text        NOT NULL,
    user_id              uuid        NOT NULL,
    task_id              uuid        NULL,
    calendar_event_id    uuid        NULL,
    title                text        NOT NULL,
    remind_at            timestamptz NOT NULL,
    status               text        NOT NULL DEFAULT 'scheduled',
    channel_preference   text        NULL,
    last_error           text        NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,

    CONSTRAINT reminders_pkey PRIMARY KEY (id),
    CONSTRAINT reminders_public_id_key UNIQUE (public_id),
    CONSTRAINT reminders_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT reminders_task_id_fkey
        FOREIGN KEY (task_id) REFERENCES tasks (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT reminders_calendar_event_id_fkey
        FOREIGN KEY (calendar_event_id) REFERENCES calendar_events (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT reminders_public_id_prefix_check CHECK (public_id LIKE 'rem_%'),
    CONSTRAINT reminders_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT reminders_status_check
        CHECK (status IN ('scheduled', 'sent', 'cancelled', 'failed'))
);

CREATE INDEX IF NOT EXISTS reminders_status_remind_at_live_idx
    ON reminders (status, remind_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS reminders_user_id_status_remind_at_live_idx
    ON reminders (user_id, status, remind_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS reminders_task_id_live_idx
    ON reminders (task_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS reminders_calendar_event_id_live_idx
    ON reminders (calendar_event_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE reminders IS 'Fire-at-time nudges; remind_at is schedule source of truth.';
COMMENT ON COLUMN reminders.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN reminders.public_id IS 'Stable API identifier with rem_ prefix.';
COMMENT ON COLUMN reminders.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN reminders.task_id IS 'Optional associated task; app soft-cascades on task soft-delete.';
COMMENT ON COLUMN reminders.calendar_event_id IS 'Optional associated event; SET NULL on hard event removal.';
COMMENT ON COLUMN reminders.title IS 'Reminder text.';
COMMENT ON COLUMN reminders.remind_at IS 'Fire time (UTC); source of truth for scheduler_jobs.run_at.';
COMMENT ON COLUMN reminders.status IS 'scheduled | sent | cancelled | failed.';
COMMENT ON COLUMN reminders.channel_preference IS 'Optional channel override.';
COMMENT ON COLUMN reminders.last_error IS 'Last delivery failure detail.';
COMMENT ON COLUMN reminders.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN reminders.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN reminders.deleted_at IS 'Soft-delete marker; NULL means live.';
