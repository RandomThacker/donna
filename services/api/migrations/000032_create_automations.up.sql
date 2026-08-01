-- Automations: scheduled execution of Donna chat commands.
-- Migrates legacy intent_rules (Scheduled Intent Check-ins MVP).

CREATE TABLE IF NOT EXISTS automations (
    id                 uuid        NOT NULL,
    public_id          text        NOT NULL,
    user_id            uuid        NOT NULL,
    name               text        NOT NULL,
    description        text        NULL,
    enabled            boolean     NOT NULL DEFAULT true,
    trigger_type       text        NOT NULL DEFAULT 'daily',
    trigger_time       time        NOT NULL,
    timezone           text        NOT NULL DEFAULT 'UTC',
    commands           jsonb       NOT NULL,
    delivery_channels  jsonb       NOT NULL DEFAULT '["chat"]'::jsonb,
    template_id        text        NULL,
    last_run_at        timestamptz NULL,
    next_run_at        timestamptz NULL,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL,
    deleted_at         timestamptz NULL,

    CONSTRAINT automations_pkey PRIMARY KEY (id),
    CONSTRAINT automations_public_id_key UNIQUE (public_id),
    CONSTRAINT automations_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT automations_public_id_prefix_check CHECK (public_id LIKE 'atm_%'),
    CONSTRAINT automations_name_not_empty_check CHECK (name <> ''),
    CONSTRAINT automations_trigger_type_check CHECK (trigger_type IN ('daily')),
    CONSTRAINT automations_timezone_not_empty_check CHECK (timezone <> ''),
    CONSTRAINT automations_commands_array_check CHECK (jsonb_typeof(commands) = 'array'),
    CONSTRAINT automations_delivery_array_check CHECK (jsonb_typeof(delivery_channels) = 'array')
);

CREATE INDEX IF NOT EXISTS automations_user_id_live_idx
    ON automations (user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS automations_enabled_live_idx
    ON automations (enabled)
    WHERE deleted_at IS NULL AND enabled = true;

CREATE INDEX IF NOT EXISTS automations_next_run_at_live_idx
    ON automations (next_run_at)
    WHERE deleted_at IS NULL AND enabled = true;

COMMENT ON TABLE automations IS 'User automations: scheduled chat-command executions posted into delivery channels.';
COMMENT ON COLUMN automations.public_id IS 'Stable API identifier with atm_ prefix.';
COMMENT ON COLUMN automations.trigger_type IS 'Phase 1: daily only. Future: weekly, monthly, cron.';
COMMENT ON COLUMN automations.trigger_time IS 'Civil clock time (HH:MM) in the automation timezone.';
COMMENT ON COLUMN automations.commands IS 'Ordered natural-language commands executed via chat Executor.';
COMMENT ON COLUMN automations.delivery_channels IS 'Delivery targets. Phase 1: chat. Future: telegram, push, whatsapp, email.';
COMMENT ON COLUMN automations.template_id IS 'Optional Intent Catalog template id used at creation.';
COMMENT ON COLUMN automations.last_run_at IS 'Idempotency marker for the last successful run.';
COMMENT ON COLUMN automations.next_run_at IS 'Best-effort next scheduled run (UTC).';
COMMENT ON COLUMN automations.deleted_at IS 'Soft-delete marker; NULL means live.';

-- Migrate Scheduled Intent Check-ins → Automations (no data loss).
INSERT INTO automations (
    id, public_id, user_id, name, description, enabled,
    trigger_type, trigger_time, timezone, commands, delivery_channels,
    template_id, last_run_at, next_run_at, created_at, updated_at, deleted_at
)
SELECT
    r.id,
    'atm_' || substr(r.public_id, 5),
    r.user_id,
    COALESCE(
        NULLIF(btrim(r.label), ''),
        CASE r.intent_kind
            WHEN 'QUERY_DUE_TODAY' THEN 'Due today'
            WHEN 'QUERY_TODAY' THEN 'What''s today'
            WHEN 'QUERY_TOMORROW' THEN 'What''s tomorrow'
            WHEN 'GREETING' THEN 'Say hello'
            ELSE r.intent_kind
        END
    ),
    CASE r.intent_kind
        WHEN 'QUERY_DUE_TODAY' THEN 'Posts open tasks due today into chat.'
        WHEN 'QUERY_TODAY' THEN 'Posts today''s agenda into chat.'
        WHEN 'QUERY_TOMORROW' THEN 'Posts tomorrow''s agenda into chat.'
        WHEN 'GREETING' THEN 'A warm hello from Donna in chat.'
        ELSE NULL
    END,
    r.enabled,
    'daily',
    r.local_time,
    r.timezone,
    CASE r.intent_kind
        WHEN 'QUERY_DUE_TODAY' THEN '["What''s due today?"]'::jsonb
        WHEN 'QUERY_TODAY' THEN '["What do I have today?"]'::jsonb
        WHEN 'QUERY_TOMORROW' THEN '["What do I have tomorrow?"]'::jsonb
        WHEN 'GREETING' THEN '["Hi"]'::jsonb
        ELSE '["Hi"]'::jsonb
    END,
    '["chat"]'::jsonb,
    CASE r.intent_kind
        WHEN 'QUERY_DUE_TODAY' THEN 'task_due'
        WHEN 'QUERY_TODAY' THEN 'todays_agenda'
        WHEN 'QUERY_TOMORROW' THEN 'tomorrow_prep'
        WHEN 'GREETING' THEN 'say_hello'
        ELSE NULL
    END,
    r.last_fired_at,
    NULL,
    r.created_at,
    r.updated_at,
    r.deleted_at
FROM intent_rules r
WHERE NOT EXISTS (
    SELECT 1 FROM automations a WHERE a.id = r.id
);

DROP TABLE IF EXISTS intent_rules;
