-- Domain completion: tasks.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § tasks
-- Status vocabulary locked (R-08): open | completed | cancelled.

CREATE TABLE IF NOT EXISTS tasks (
    id                uuid        NOT NULL,
    public_id         text        NOT NULL,
    user_id           uuid        NOT NULL,
    title             text        NOT NULL,
    description       text        NULL,
    status            text        NOT NULL DEFAULT 'open',
    priority          text        NULL,
    due_at            timestamptz NULL,
    completed_at      timestamptz NULL,
    is_backlog        boolean     NOT NULL DEFAULT false,
    recurrence_rule   text        NULL,
    provider          text        NULL,
    provider_task_id  text        NULL,
    provider_payload  jsonb       NULL,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz NULL,

    CONSTRAINT tasks_pkey PRIMARY KEY (id),
    CONSTRAINT tasks_public_id_key UNIQUE (public_id),
    CONSTRAINT tasks_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT tasks_public_id_prefix_check CHECK (public_id LIKE 'tsk_%'),
    CONSTRAINT tasks_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT tasks_status_check
        CHECK (status IN ('open', 'completed', 'cancelled')),
    CONSTRAINT tasks_priority_check
        CHECK (priority IS NULL OR priority IN ('low', 'medium', 'high')),
    CONSTRAINT tasks_completed_at_check
        CHECK ((status = 'completed' AND completed_at IS NOT NULL) OR (status <> 'completed')),
    CONSTRAINT tasks_provider_payload_object_check
        CHECK (provider_payload IS NULL OR jsonb_typeof(provider_payload) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS tasks_user_provider_task_live_uidx
    ON tasks (user_id, provider, provider_task_id)
    WHERE deleted_at IS NULL AND provider_task_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS tasks_user_id_status_due_at_live_idx
    ON tasks (user_id, status, due_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS tasks_user_id_backlog_status_live_idx
    ON tasks (user_id, is_backlog, status)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE tasks IS 'Actionable work items owned by a Donna user.';
COMMENT ON COLUMN tasks.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN tasks.public_id IS 'Stable API identifier with tsk_ prefix.';
COMMENT ON COLUMN tasks.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN tasks.title IS 'Task title.';
COMMENT ON COLUMN tasks.description IS 'Optional task body.';
COMMENT ON COLUMN tasks.status IS 'open | completed | cancelled.';
COMMENT ON COLUMN tasks.priority IS 'Optional priority: low | medium | high.';
COMMENT ON COLUMN tasks.due_at IS 'Optional due instant (UTC).';
COMMENT ON COLUMN tasks.completed_at IS 'Required when status = completed; cleared on reopen.';
COMMENT ON COLUMN tasks.is_backlog IS 'Backlog widget flag.';
COMMENT ON COLUMN tasks.recurrence_rule IS 'Optional recurrence summary.';
COMMENT ON COLUMN tasks.provider IS 'Optional future external tasks provider.';
COMMENT ON COLUMN tasks.provider_task_id IS 'Remote task id when synced.';
COMMENT ON COLUMN tasks.provider_payload IS 'Opaque provider sync payload (jsonb object).';
COMMENT ON COLUMN tasks.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN tasks.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN tasks.deleted_at IS 'Soft-delete marker; NULL means live.';
