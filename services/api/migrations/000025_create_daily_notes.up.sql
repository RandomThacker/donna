-- Daily Task Journal: per-day markdown notes.
CREATE TABLE IF NOT EXISTS daily_notes (
    id          uuid        NOT NULL,
    public_id   text        NOT NULL,
    user_id     uuid        NOT NULL,
    date        date        NOT NULL,
    content     text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL,

    CONSTRAINT daily_notes_pkey PRIMARY KEY (id),
    CONSTRAINT daily_notes_public_id_key UNIQUE (public_id),
    CONSTRAINT daily_notes_user_date_key UNIQUE (user_id, date),
    CONSTRAINT daily_notes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT daily_notes_public_id_prefix_check CHECK (public_id LIKE 'dnt_%')
);

CREATE INDEX IF NOT EXISTS daily_notes_user_id_date_idx
    ON daily_notes (user_id, date);

COMMENT ON TABLE daily_notes IS 'Markdown daily note for a civil journal day.';
