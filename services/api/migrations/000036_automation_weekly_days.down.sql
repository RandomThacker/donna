ALTER TABLE automations
    DROP CONSTRAINT IF EXISTS automations_trigger_days_check;

ALTER TABLE automations
    DROP CONSTRAINT IF EXISTS automations_trigger_type_check;

UPDATE automations
SET trigger_type = 'daily',
    trigger_days = '{}'::text[]
WHERE trigger_type = 'weekly';

ALTER TABLE automations
    ADD CONSTRAINT automations_trigger_type_check
    CHECK (trigger_type IN ('daily'));

ALTER TABLE automations
    DROP COLUMN IF EXISTS trigger_days;
