-- Domain completion: conversations.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § conversations

CREATE TABLE IF NOT EXISTS conversations (
    id                   uuid        NOT NULL,
    public_id            text        NOT NULL,
    user_id              uuid        NOT NULL,
    title                text        NULL,
    purpose              text        NULL,
    status               text        NOT NULL DEFAULT 'active',
    unread_count         integer     NOT NULL DEFAULT 0,
    last_message_at      timestamptz NULL,
    channel              text        NOT NULL DEFAULT 'web',
    channel_thread_id    text        NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,

    CONSTRAINT conversations_pkey PRIMARY KEY (id),
    CONSTRAINT conversations_public_id_key UNIQUE (public_id),
    CONSTRAINT conversations_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT conversations_public_id_prefix_check CHECK (public_id LIKE 'conv_%'),
    CONSTRAINT conversations_status_check
        CHECK (status IN ('active', 'archived')),
    CONSTRAINT conversations_channel_check
        CHECK (channel IN ('web', 'telegram', 'whatsapp')),
    CONSTRAINT conversations_purpose_check
        CHECK (purpose IS NULL OR purpose IN ('general', 'morning', 'midday', 'evening', 'system')),
    CONSTRAINT conversations_unread_count_check CHECK (unread_count >= 0)
);

CREATE INDEX IF NOT EXISTS conversations_user_id_last_message_at_live_idx
    ON conversations (user_id, last_message_at DESC NULLS LAST)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS conversations_user_id_status_live_idx
    ON conversations (user_id, status)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE conversations IS 'Chat threads with Donna.';
COMMENT ON COLUMN conversations.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN conversations.public_id IS 'Stable API identifier with conv_ prefix.';
COMMENT ON COLUMN conversations.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN conversations.title IS 'Optional thread title.';
COMMENT ON COLUMN conversations.purpose IS 'Optional purpose: general | morning | midday | evening | system.';
COMMENT ON COLUMN conversations.status IS 'active | archived.';
COMMENT ON COLUMN conversations.unread_count IS 'Denormalized unread count; maintained by application.';
COMMENT ON COLUMN conversations.last_message_at IS 'Timestamp of latest message for list sorting.';
COMMENT ON COLUMN conversations.channel IS 'Surface: web | telegram | whatsapp (Phase 1 uses web).';
COMMENT ON COLUMN conversations.channel_thread_id IS 'External thread identifier when applicable.';
COMMENT ON COLUMN conversations.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN conversations.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN conversations.deleted_at IS 'Soft-delete marker; NULL means live.';
