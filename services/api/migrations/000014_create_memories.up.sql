-- Domain completion: memories.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § memories
-- Embedding column deferred until pgvector search ships.

CREATE TABLE IF NOT EXISTS memories (
    id                        uuid        NOT NULL,
    public_id                 text        NOT NULL,
    user_id                   uuid        NOT NULL,
    content                   text        NOT NULL,
    category                  text        NULL,
    importance                integer     NULL,
    source                    text        NOT NULL,
    source_conversation_id    uuid        NULL,
    source_message_id         uuid        NULL,
    embedding_model           text        NULL,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz NULL,

    CONSTRAINT memories_pkey PRIMARY KEY (id),
    CONSTRAINT memories_public_id_key UNIQUE (public_id),
    CONSTRAINT memories_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT memories_source_conversation_id_fkey
        FOREIGN KEY (source_conversation_id) REFERENCES conversations (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT memories_source_message_id_fkey
        FOREIGN KEY (source_message_id) REFERENCES messages (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT memories_public_id_prefix_check CHECK (public_id LIKE 'mem_%'),
    CONSTRAINT memories_content_not_empty_check CHECK (content <> ''),
    CONSTRAINT memories_source_check
        CHECK (source IN ('explicit', 'chat_extract', 'review', 'system')),
    CONSTRAINT memories_category_check
        CHECK (category IS NULL OR category IN (
            'preference', 'person', 'project', 'commitment', 'idea', 'other'
        )),
    CONSTRAINT memories_importance_check
        CHECK (importance IS NULL OR (importance >= 1 AND importance <= 100))
);

CREATE INDEX IF NOT EXISTS memories_user_id_updated_at_live_idx
    ON memories (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS memories_user_id_category_live_idx
    ON memories (user_id, category)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE memories IS 'Durable recalled knowledge; vector embedding column deferred to search milestone.';
COMMENT ON COLUMN memories.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN memories.public_id IS 'Stable API identifier with mem_ prefix.';
COMMENT ON COLUMN memories.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN memories.content IS 'Fact / memory text.';
COMMENT ON COLUMN memories.category IS 'Optional category: preference | person | project | commitment | idea | other.';
COMMENT ON COLUMN memories.importance IS 'Optional rank 1–100.';
COMMENT ON COLUMN memories.source IS 'Provenance kind: explicit | chat_extract | review | system.';
COMMENT ON COLUMN memories.source_conversation_id IS 'Optional provenance conversation.';
COMMENT ON COLUMN memories.source_message_id IS 'Optional provenance message.';
COMMENT ON COLUMN memories.embedding_model IS 'Model used when embedding is generated later.';
COMMENT ON COLUMN memories.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN memories.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN memories.deleted_at IS 'Soft-delete marker; NULL means live.';
