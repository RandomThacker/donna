-- Phase 1.6: migrate automations.commands from string[] to structured
-- [{ "command": "...", "variables": { ... } }].

UPDATE automations
SET commands = (
    SELECT COALESCE(jsonb_agg(converted ORDER BY ord), '[]'::jsonb)
    FROM (
        SELECT
            ord,
            CASE
                WHEN jsonb_typeof(elem) = 'object' THEN elem
                WHEN lower(elem #>> '{}') IN (
                    'what do i have today?',
                    'what''s on today',
                    'whats on today'
                ) THEN jsonb_build_object(
                    'command', 'todays_agenda',
                    'variables', jsonb_build_object('range', 'today')
                )
                WHEN lower(elem #>> '{}') IN (
                    'what do i have tomorrow?',
                    'what''s on tomorrow',
                    'whats on tomorrow'
                ) THEN jsonb_build_object(
                    'command', 'todays_agenda',
                    'variables', jsonb_build_object('range', 'tomorrow')
                )
                WHEN lower(elem #>> '{}') IN (
                    'what''s due today?',
                    'whats due today?',
                    'due today'
                ) THEN jsonb_build_object(
                    'command', 'tasks_due',
                    'variables', jsonb_build_object('priority', 'all')
                )
                WHEN lower(elem #>> '{}') IN ('hi', 'hello', 'hey') THEN
                    jsonb_build_object('command', 'greeting')
                ELSE jsonb_build_object(
                    'command', 'chat_message',
                    'variables', jsonb_build_object('message', elem #>> '{}')
                )
            END AS converted
        FROM jsonb_array_elements(commands) WITH ORDINALITY AS t(elem, ord)
    ) mapped
)
WHERE commands IS NOT NULL;

COMMENT ON COLUMN automations.commands IS
    'Ordered structured commands: [{command, variables}]. Legacy strings migrated to chat_message.';
