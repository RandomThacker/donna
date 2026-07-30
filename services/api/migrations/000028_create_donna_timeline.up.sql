-- Donna-owned timeline items (events + reminders).
-- Provider calendar events stay in calendar_events; these tables are Donna-native only.

CREATE TABLE IF NOT EXISTS donna_events (
    id                         uuid        NOT NULL,
    public_id                  text        NOT NULL,
    user_id                    uuid        NOT NULL,
    title                      text        NOT NULL,
    description                text        NULL,
    start_at                   timestamptz NOT NULL,
    end_at                     timestamptz NOT NULL,
    timezone                   text        NOT NULL DEFAULT 'UTC',
    all_day                    boolean     NOT NULL DEFAULT false,
    location                   text        NULL,
    reminder_offset_minutes    integer     NULL,
    recurrence_rule            text        NULL,
    status                     text        NOT NULL DEFAULT 'confirmed',
    color                      text        NULL,
    created_at                 timestamptz NOT NULL,
    updated_at                 timestamptz NOT NULL,
    deleted_at                 timestamptz NULL,

    CONSTRAINT donna_events_pkey PRIMARY KEY (id),
    CONSTRAINT donna_events_public_id_key UNIQUE (public_id),
    CONSTRAINT donna_events_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT donna_events_public_id_prefix_check CHECK (public_id LIKE 'dev_%'),
    CONSTRAINT donna_events_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT donna_events_range_check CHECK (end_at >= start_at),
    CONSTRAINT donna_events_status_check CHECK (status IN ('confirmed', 'cancelled')),
    CONSTRAINT donna_events_reminder_offset_check
        CHECK (reminder_offset_minutes IS NULL OR reminder_offset_minutes >= 0)
);

CREATE INDEX IF NOT EXISTS donna_events_user_id_start_at_live_idx
    ON donna_events (user_id, start_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS donna_events_user_id_range_live_idx
    ON donna_events (user_id, start_at, end_at)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE donna_events IS 'Donna-owned calendar events for the unified timeline.';
COMMENT ON COLUMN donna_events.public_id IS 'Stable API identifier with dev_ prefix.';
COMMENT ON COLUMN donna_events.recurrence_rule IS 'Stored RRULE; expansion is a future phase.';
COMMENT ON COLUMN donna_events.deleted_at IS 'Soft-delete marker; NULL means live.';

CREATE TABLE IF NOT EXISTS donna_reminders (
    id                uuid        NOT NULL,
    public_id         text        NOT NULL,
    user_id           uuid        NOT NULL,
    title             text        NOT NULL,
    description       text        NULL,
    trigger_at        timestamptz NOT NULL,
    timezone          text        NOT NULL DEFAULT 'UTC',
    recurrence_rule   text        NULL,
    status            text        NOT NULL DEFAULT 'scheduled',
    color             text        NULL,
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL,
    deleted_at        timestamptz NULL,

    CONSTRAINT donna_reminders_pkey PRIMARY KEY (id),
    CONSTRAINT donna_reminders_public_id_key UNIQUE (public_id),
    CONSTRAINT donna_reminders_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT donna_reminders_public_id_prefix_check CHECK (public_id LIKE 'drm_%'),
    CONSTRAINT donna_reminders_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT donna_reminders_status_check
        CHECK (status IN ('scheduled', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS donna_reminders_user_id_trigger_at_live_idx
    ON donna_reminders (user_id, trigger_at)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE donna_reminders IS 'Donna-owned standalone reminders for the unified timeline.';
COMMENT ON COLUMN donna_reminders.public_id IS 'Stable API identifier with drm_ prefix.';
COMMENT ON COLUMN donna_reminders.trigger_at IS 'When the reminder should surface on the timeline.';
COMMENT ON COLUMN donna_reminders.recurrence_rule IS 'Stored RRULE; expansion is a future phase.';
COMMENT ON COLUMN donna_reminders.deleted_at IS 'Soft-delete marker; NULL means live.';
