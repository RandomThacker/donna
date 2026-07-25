-- Domain completion: messages.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § messages
-- ai_session_id included nullable without FK until ai_sessions exists (wave 2 per physical design).

CREATE TABLE IF NOT EXISTS messages (
    id                  uuid        NOT NULL,
    public_id           text        NOT NULL,
    user_id             uuid        NOT NULL,
    conversation_id     uuid        NOT NULL,
    role                text        NOT NULL,
    content             text        NOT NULL,
    content_format      text        NOT NULL DEFAULT 'plain',
    client_message_id   text        NULL,
    citations           jsonb       NOT NULL DEFAULT '{}'::jsonb,
    ai_session_id       uuid        NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz NULL,

    CONSTRAINT messages_pkey PRIMARY KEY (id),
    CONSTRAINT messages_public_id_key UNIQUE (public_id),
    CONSTRAINT messages_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT messages_conversation_id_fkey
        FOREIGN KEY (conversation_id) REFERENCES conversations (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT messages_public_id_prefix_check CHECK (public_id LIKE 'msg_%'),
    CONSTRAINT messages_role_check
        CHECK (role IN ('user', 'assistant', 'system')),
    CONSTRAINT messages_content_format_check
        CHECK (content_format IN ('plain', 'markdown')),
    CONSTRAINT messages_content_not_empty_check CHECK (content <> ''),
    CONSTRAINT messages_citations_object_check
        CHECK (jsonb_typeof(citations) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS messages_conversation_client_message_live_uidx
    ON messages (conversation_id, client_message_id)
    WHERE deleted_at IS NULL AND client_message_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS messages_conversation_id_created_at_live_idx
    ON messages (conversation_id, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS messages_user_id_created_at_live_idx
    ON messages (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE messages IS 'Conversation turns; citations remain jsonb for v1 (R-05).';
COMMENT ON COLUMN messages.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN messages.public_id IS 'Stable API identifier with msg_ prefix.';
COMMENT ON COLUMN messages.user_id IS 'Owning Donna user (tenancy).';
COMMENT ON COLUMN messages.conversation_id IS 'Parent conversation.';
COMMENT ON COLUMN messages.role IS 'user | assistant | system.';
COMMENT ON COLUMN messages.content IS 'Message body.';
COMMENT ON COLUMN messages.content_format IS 'plain | markdown.';
COMMENT ON COLUMN messages.client_message_id IS 'Optional client idempotency key.';
COMMENT ON COLUMN messages.citations IS 'Soft refs object: task_ids / event_ids / memory_ids (prefer internal UUIDs).';
COMMENT ON COLUMN messages.ai_session_id IS 'Optional AI session link; FK deferred until ai_sessions table exists.';
COMMENT ON COLUMN messages.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN messages.updated_at IS 'Last mutation time; content preferably immutable.';
COMMENT ON COLUMN messages.deleted_at IS 'Soft-delete marker; NULL means live.';
