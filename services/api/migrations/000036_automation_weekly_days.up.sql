-- Automations: support weekly schedules with selected weekdays.

ALTER TABLE automations
    ADD COLUMN IF NOT EXISTS trigger_days text[] NOT NULL DEFAULT '{}'::text[];

ALTER TABLE automations
    DROP CONSTRAINT IF EXISTS automations_trigger_type_check;

ALTER TABLE automations
    ADD CONSTRAINT automations_trigger_type_check
    CHECK (trigger_type IN ('daily', 'weekly'));

ALTER TABLE automations
    DROP CONSTRAINT IF EXISTS automations_trigger_days_check;

ALTER TABLE automations
    ADD CONSTRAINT automations_trigger_days_check CHECK (
        (
            trigger_type = 'daily'
            AND cardinality(trigger_days) = 0
        )
        OR (
            trigger_type = 'weekly'
            AND cardinality(trigger_days) BETWEEN 1 AND 7
            AND trigger_days <@ ARRAY['MO','TU','WE','TH','FR','SA','SU']::text[]
        )
    );

COMMENT ON COLUMN automations.trigger_type IS 'daily (every day) or weekly (selected weekdays).';
COMMENT ON COLUMN automations.trigger_days IS 'RRULE weekday codes for weekly triggers (MO..SU). Empty for daily.';
