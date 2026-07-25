-- Domain completion: notifications.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § notifications
-- Delivery vs read separated (R-04): delivery_status + read_at / dismissed_at.

CREATE TABLE IF NOT EXISTS notifications (
    id                   uuid        NOT NULL,
    public_id            text        NOT NULL,
    user_id              uuid        NOT NULL,
    channel              text        NOT NULL DEFAULT 'browser_push',
    title                text        NOT NULL,
    body                 text        NOT NULL,
    priority             text        NOT NULL DEFAULT 'normal',
    delivery_status      text        NOT NULL DEFAULT 'pending',
    reminder_id          uuid        NULL,
    calendar_event_id    uuid        NULL,
    scheduler_job_id     uuid        NULL,
    conversation_id      uuid        NULL,
    payload              jsonb       NOT NULL DEFAULT '{}'::jsonb,
    sent_at              timestamptz NULL,
    read_at              timestamptz NULL,
    dismissed_at         timestamptz NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,

    CONSTRAINT notifications_pkey PRIMARY KEY (id),
    CONSTRAINT notifications_public_id_key UNIQUE (public_id),
    CONSTRAINT notifications_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT notifications_reminder_id_fkey
        FOREIGN KEY (reminder_id) REFERENCES reminders (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT notifications_calendar_event_id_fkey
        FOREIGN KEY (calendar_event_id) REFERENCES calendar_events (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT notifications_scheduler_job_id_fkey
        FOREIGN KEY (scheduler_job_id) REFERENCES scheduler_jobs (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT notifications_conversation_id_fkey
        FOREIGN KEY (conversation_id) REFERENCES conversations (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT notifications_public_id_prefix_check CHECK (public_id LIKE 'ntf_%'),
    CONSTRAINT notifications_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT notifications_body_not_empty_check CHECK (body <> ''),
    CONSTRAINT notifications_channel_check
        CHECK (channel IN ('browser_push', 'email', 'telegram')),
    CONSTRAINT notifications_priority_check
        CHECK (priority IN ('low', 'normal', 'high')),
    CONSTRAINT notifications_delivery_status_check
        CHECK (delivery_status IN ('pending', 'sent', 'failed')),
    CONSTRAINT notifications_sent_at_check
        CHECK ((delivery_status <> 'sent') OR (sent_at IS NOT NULL)),
    CONSTRAINT notifications_payload_object_check
        CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX IF NOT EXISTS notifications_user_id_created_at_live_idx
    ON notifications (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS notifications_user_id_unread_live_idx
    ON notifications (user_id, created_at DESC)
    WHERE deleted_at IS NULL AND read_at IS NULL;

CREATE INDEX IF NOT EXISTS notifications_user_id_delivery_status_live_idx
    ON notifications (user_id, delivery_status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS notifications_reminder_id_idx
    ON notifications (reminder_id);

COMMENT ON TABLE notifications IS 'User-visible alerts; delivery_status separate from read_at/dismissed_at (R-04).';
COMMENT ON COLUMN notifications.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN notifications.public_id IS 'Stable API identifier with ntf_ prefix.';
COMMENT ON COLUMN notifications.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN notifications.channel IS 'browser_push | email | telegram.';
COMMENT ON COLUMN notifications.title IS 'Notification title.';
COMMENT ON COLUMN notifications.body IS 'Notification body.';
COMMENT ON COLUMN notifications.priority IS 'low | normal | high.';
COMMENT ON COLUMN notifications.delivery_status IS 'pending | sent | failed.';
COMMENT ON COLUMN notifications.reminder_id IS 'Optional reminder correlation.';
COMMENT ON COLUMN notifications.calendar_event_id IS 'Optional calendar event correlation.';
COMMENT ON COLUMN notifications.scheduler_job_id IS 'Optional scheduler job correlation.';
COMMENT ON COLUMN notifications.conversation_id IS 'Optional conversation correlation.';
COMMENT ON COLUMN notifications.payload IS 'Non-secret channel extras (jsonb object); not a substitute for FKs.';
COMMENT ON COLUMN notifications.sent_at IS 'Required when delivery_status = sent.';
COMMENT ON COLUMN notifications.read_at IS 'User read marker.';
COMMENT ON COLUMN notifications.dismissed_at IS 'User dismiss marker.';
COMMENT ON COLUMN notifications.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN notifications.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN notifications.deleted_at IS 'Soft-delete marker; NULL means live.';
