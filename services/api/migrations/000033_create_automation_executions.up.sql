-- Automation execution history (observability). Automations are templates; executions are runs.

CREATE TABLE IF NOT EXISTS automation_executions (
    id                  uuid        NOT NULL,
    public_id           text        NOT NULL,
    automation_id       uuid        NOT NULL,
    user_id             uuid        NOT NULL,
    started_at          timestamptz NOT NULL,
    completed_at        timestamptz NULL,
    status              text        NOT NULL,
    duration_ms         integer     NULL,
    commands_total      integer     NOT NULL DEFAULT 0,
    commands_success    integer     NOT NULL DEFAULT 0,
    commands_failed     integer     NOT NULL DEFAULT 0,
    trigger_source      text        NOT NULL DEFAULT 'scheduler',
    delivery_channels   jsonb       NOT NULL DEFAULT '["chat"]'::jsonb,
    delivery_status     text        NULL,
    response            text        NULL,
    error               text        NULL,
    created_at          timestamptz NOT NULL,
    updated_at          timestamptz NOT NULL,

    CONSTRAINT automation_executions_pkey PRIMARY KEY (id),
    CONSTRAINT automation_executions_public_id_key UNIQUE (public_id),
    CONSTRAINT automation_executions_automation_id_fkey
        FOREIGN KEY (automation_id) REFERENCES automations (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT automation_executions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT automation_executions_public_id_prefix_check CHECK (public_id LIKE 'aex_%'),
    CONSTRAINT automation_executions_status_check CHECK (status IN (
        'RUNNING', 'SUCCESS', 'PARTIAL_SUCCESS', 'FAILED', 'CANCELLED'
    )),
    CONSTRAINT automation_executions_trigger_source_not_empty_check CHECK (trigger_source <> ''),
    CONSTRAINT automation_executions_delivery_array_check CHECK (jsonb_typeof(delivery_channels) = 'array'),
    CONSTRAINT automation_executions_counts_nonneg_check
        CHECK (commands_total >= 0 AND commands_success >= 0 AND commands_failed >= 0),
    CONSTRAINT automation_executions_duration_nonneg_check
        CHECK (duration_ms IS NULL OR duration_ms >= 0)
);

CREATE INDEX IF NOT EXISTS automation_executions_automation_id_started_idx
    ON automation_executions (automation_id, started_at DESC);

CREATE INDEX IF NOT EXISTS automation_executions_user_id_started_idx
    ON automation_executions (user_id, started_at DESC);

CREATE INDEX IF NOT EXISTS automation_executions_status_idx
    ON automation_executions (status);

COMMENT ON TABLE automation_executions IS 'One row per automation run; history for observability, replay, and future delivery channels.';
COMMENT ON COLUMN automation_executions.public_id IS 'Stable API identifier with aex_ prefix.';
COMMENT ON COLUMN automation_executions.trigger_source IS 'How the run was started: scheduler, manual, retry, replay (future).';
COMMENT ON COLUMN automation_executions.delivery_status IS 'Aggregate delivery outcome: PENDING, SENT, FAILED, SKIPPED.';
COMMENT ON COLUMN automation_executions.response IS 'Combined assistant response delivered (or attempted).';

CREATE TABLE IF NOT EXISTS automation_command_executions (
    id              uuid        NOT NULL,
    public_id       text        NOT NULL,
    execution_id    uuid        NOT NULL,
    order_index     integer     NOT NULL,
    command         text        NOT NULL,
    command_type    text        NULL,
    started_at      timestamptz NOT NULL,
    completed_at    timestamptz NULL,
    status          text        NOT NULL,
    duration_ms     integer     NULL,
    response        text        NULL,
    error           text        NULL,
    created_at      timestamptz NOT NULL,

    CONSTRAINT automation_command_executions_pkey PRIMARY KEY (id),
    CONSTRAINT automation_command_executions_public_id_key UNIQUE (public_id),
    CONSTRAINT automation_command_executions_execution_id_fkey
        FOREIGN KEY (execution_id) REFERENCES automation_executions (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION,
    CONSTRAINT automation_command_executions_public_id_prefix_check CHECK (public_id LIKE 'ace_%'),
    CONSTRAINT automation_command_executions_status_check CHECK (status IN (
        'SUCCESS', 'FAILED', 'SKIPPED'
    )),
    CONSTRAINT automation_command_executions_order_nonneg_check CHECK (order_index >= 0),
    CONSTRAINT automation_command_executions_command_not_empty_check CHECK (command <> ''),
    CONSTRAINT automation_command_executions_duration_nonneg_check
        CHECK (duration_ms IS NULL OR duration_ms >= 0),
    CONSTRAINT automation_command_executions_execution_order_uidx UNIQUE (execution_id, order_index)
);

CREATE INDEX IF NOT EXISTS automation_command_executions_execution_id_idx
    ON automation_command_executions (execution_id, order_index ASC);

COMMENT ON TABLE automation_command_executions IS 'Per-command results within an automation execution.';
COMMENT ON COLUMN automation_command_executions.public_id IS 'Stable API identifier with ace_ prefix.';
COMMENT ON COLUMN automation_command_executions.command_type IS 'Parsed intent kind when available (e.g. QUERY_TODAY).';
