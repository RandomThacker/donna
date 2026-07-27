-- Daily Task Journal: per-day task snapshots.
CREATE TABLE IF NOT EXISTS task_occurrences (
    id               uuid        NOT NULL,
    public_id        text        NOT NULL,
    task_id          uuid        NOT NULL,
    user_id          uuid        NOT NULL,
    date             date        NOT NULL,
    sort_order       integer     NOT NULL DEFAULT 0,
    completed        boolean     NOT NULL DEFAULT false,
    completed_at     timestamptz NULL,
    carried_forward  boolean     NOT NULL DEFAULT false,
    source           text        NOT NULL,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL,

    CONSTRAINT task_occurrences_pkey PRIMARY KEY (id),
    CONSTRAINT task_occurrences_public_id_key UNIQUE (public_id),
    CONSTRAINT task_occurrences_task_date_key UNIQUE (task_id, date),
    CONSTRAINT task_occurrences_task_id_fkey
        FOREIGN KEY (task_id) REFERENCES tasks (id),
    CONSTRAINT task_occurrences_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT task_occurrences_public_id_prefix_check CHECK (public_id LIKE 'toc_%'),
    CONSTRAINT task_occurrences_source_check
        CHECK (source IN ('manual', 'recurring', 'calendar', 'ai', 'carry_forward')),
    CONSTRAINT task_occurrences_completed_at_check
        CHECK ((completed = false AND completed_at IS NULL) OR (completed = true AND completed_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS task_occurrences_user_id_date_idx
    ON task_occurrences (user_id, date);

CREATE INDEX IF NOT EXISTS task_occurrences_user_id_date_sort_idx
    ON task_occurrences (user_id, date, sort_order);

COMMENT ON TABLE task_occurrences IS 'Immutable daily journal row for a task on a civil day.';
COMMENT ON COLUMN task_occurrences.date IS 'Civil day (user journal page).';
COMMENT ON COLUMN task_occurrences.sort_order IS 'Ordering within the day only.';
COMMENT ON COLUMN task_occurrences.carried_forward IS 'True when cloned from a prior incomplete day.';
COMMENT ON COLUMN task_occurrences.source IS 'manual | recurring | calendar | ai | carry_forward';
