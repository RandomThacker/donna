-- Standalone Keep-style notes (user-owned quick notes).
CREATE TABLE IF NOT EXISTS notes (
    id          uuid        NOT NULL,
    public_id   text        NOT NULL,
    user_id     uuid        NOT NULL,
    title       text        NOT NULL DEFAULT '',
    content     text        NOT NULL DEFAULT '',
    color       text        NOT NULL DEFAULT 'default',
    pinned      boolean     NOT NULL DEFAULT false,
    archived    boolean     NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL,
    deleted_at  timestamptz NULL,

    CONSTRAINT notes_pkey PRIMARY KEY (id),
    CONSTRAINT notes_public_id_key UNIQUE (public_id),
    CONSTRAINT notes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT notes_public_id_prefix_check CHECK (public_id LIKE 'nte_%'),
    CONSTRAINT notes_color_check CHECK (color IN (
        'default', 'coral', 'sage', 'sky', 'blush', 'sand', 'lilac'
    ))
);

CREATE INDEX IF NOT EXISTS notes_user_id_updated_at_live_idx
    ON notes (user_id, pinned DESC, updated_at DESC)
    WHERE deleted_at IS NULL AND archived = false;

COMMENT ON TABLE notes IS 'User-owned quick notes (Keep-style).';
COMMENT ON COLUMN notes.public_id IS 'Stable API identifier with nte_ prefix.';
COMMENT ON COLUMN notes.color IS 'default | coral | sage | sky | blush | sand | lilac';
COMMENT ON COLUMN notes.pinned IS 'Pinned notes sort above unpinned.';
COMMENT ON COLUMN notes.archived IS 'Archived notes are hidden from the main grid.';
COMMENT ON COLUMN notes.deleted_at IS 'Soft-delete marker; NULL means live.';
