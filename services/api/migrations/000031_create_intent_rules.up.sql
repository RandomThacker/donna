-- User-scheduled daily intent check-ins (run a command-chat intent at a local time).

CREATE TABLE IF NOT EXISTS intent_rules (
    id             uuid        NOT NULL,
    public_id      text        NOT NULL,
    user_id        uuid        NOT NULL,
    intent_kind    text        NOT NULL,
    label          text        NULL,
    timezone       text        NOT NULL DEFAULT 'UTC',
    local_time     time        NOT NULL,
    recurrence     text        NOT NULL DEFAULT 'daily',
    enabled        boolean     NOT NULL DEFAULT true,
    last_fired_at  timestamptz NULL,
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL,
    deleted_at     timestamptz NULL,

    CONSTRAINT intent_rules_pkey PRIMARY KEY (id),
    CONSTRAINT intent_rules_public_id_key UNIQUE (public_id),
    CONSTRAINT intent_rules_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT intent_rules_public_id_prefix_check CHECK (public_id LIKE 'inr_%'),
    CONSTRAINT intent_rules_intent_kind_check CHECK (intent_kind IN (
        'QUERY_DUE_TODAY',
        'QUERY_TODAY',
        'QUERY_TOMORROW',
        'GREETING'
    )),
    CONSTRAINT intent_rules_recurrence_check CHECK (recurrence IN ('daily')),
    CONSTRAINT intent_rules_timezone_not_empty_check CHECK (timezone <> '')
);

CREATE INDEX IF NOT EXISTS intent_rules_user_id_live_idx
    ON intent_rules (user_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS intent_rules_user_kind_time_recurrence_live_uidx
    ON intent_rules (user_id, intent_kind, local_time, recurrence)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS intent_rules_enabled_live_idx
    ON intent_rules (enabled)
    WHERE deleted_at IS NULL AND enabled = true;

COMMENT ON TABLE intent_rules IS 'Scheduled check-ins: run a query/greeting intent at a local time and post into chat.';
COMMENT ON COLUMN intent_rules.public_id IS 'Stable API identifier with inr_ prefix.';
COMMENT ON COLUMN intent_rules.intent_kind IS 'Whitelisted command-chat intent kind.';
COMMENT ON COLUMN intent_rules.local_time IS 'Civil clock time (HH:MM) in the rule timezone.';
COMMENT ON COLUMN intent_rules.recurrence IS 'MVP: daily only.';
COMMENT ON COLUMN intent_rules.last_fired_at IS 'Idempotency marker for the last successful fire.';
COMMENT ON COLUMN intent_rules.deleted_at IS 'Soft-delete marker; NULL means live.';
