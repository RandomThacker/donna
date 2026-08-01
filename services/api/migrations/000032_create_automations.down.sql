-- Reverse: restore intent_rules from automations is lossy for multi-command rows.
-- Prefer restoring from backup. This down migration only drops automations.
DROP TABLE IF EXISTS automations;
