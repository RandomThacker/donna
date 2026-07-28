-- Domain completion: audit_logs.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § audit_logs
-- Append-only: no updated_at, no deleted_at. Subject is polymorphic (no FK).

CREATE TABLE IF NOT EXISTS audit_logs (
    id                  uuid        NOT NULL,
    public_id           text        NOT NULL,
    actor_user_id       uuid        NULL,
    action              text        NOT NULL,
    subject_type        text        NULL,
    subject_id          uuid        NULL,
    subject_public_id   text        NULL,
    request_id          text        NULL,
    metadata            jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT audit_logs_pkey PRIMARY KEY (id),
    CONSTRAINT audit_logs_public_id_key UNIQUE (public_id),
    CONSTRAINT audit_logs_actor_user_id_fkey
        FOREIGN KEY (actor_user_id) REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT audit_logs_public_id_prefix_check CHECK (public_id LIKE 'audit_%'),
    CONSTRAINT audit_logs_action_not_empty_check CHECK (action <> ''),
    CONSTRAINT audit_logs_metadata_object_check
        CHECK (jsonb_typeof(metadata) = 'object'),
    CONSTRAINT audit_logs_subject_type_check
        CHECK (subject_type IS NULL OR subject_type IN (
            'user',
            'auth_identity',
            'connected_account',
            'calendar_source',
            'calendar_event',
            'task',
            'reminder',
            'conversation',
            'message',
            'memory',
            'notification',
            'scheduler_job',
            'user_settings',
            'billing'
        ))
);

CREATE INDEX IF NOT EXISTS audit_logs_actor_user_id_created_at_idx
    ON audit_logs (actor_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_subject_created_at_idx
    ON audit_logs (subject_type, subject_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_action_created_at_idx
    ON audit_logs (action, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_request_id_idx
    ON audit_logs (request_id)
    WHERE request_id IS NOT NULL;

COMMENT ON TABLE audit_logs IS 'Append-only security/compliance log; retention via ops hard-delete.';
COMMENT ON COLUMN audit_logs.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN audit_logs.public_id IS 'Stable API identifier with audit_ prefix.';
COMMENT ON COLUMN audit_logs.actor_user_id IS 'Acting user; NULL means system.';
COMMENT ON COLUMN audit_logs.action IS 'Stable action key.';
COMMENT ON COLUMN audit_logs.subject_type IS 'Polymorphic subject kind; no FK by design.';
COMMENT ON COLUMN audit_logs.subject_id IS 'Internal subject UUID snapshot.';
COMMENT ON COLUMN audit_logs.subject_public_id IS 'Public id snapshot for forensics.';
COMMENT ON COLUMN audit_logs.request_id IS 'Optional request correlation id.';
COMMENT ON COLUMN audit_logs.metadata IS 'Redacted non-secret context (jsonb object); never tokens.';
COMMENT ON COLUMN audit_logs.created_at IS 'Immutable insert time (UTC).';
