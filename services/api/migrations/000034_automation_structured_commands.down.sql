-- Best-effort reverse: structured commands → displayable string messages.

UPDATE automations
SET commands = (
    SELECT COALESCE(jsonb_agg(to_jsonb(label) ORDER BY ord), '[]'::jsonb)
    FROM (
        SELECT
            ord,
            CASE
                WHEN elem->>'command' = 'greeting' THEN 'Hi'
                WHEN elem->>'command' = 'todays_agenda'
                     AND coalesce(elem->'variables'->>'range', 'today') = 'tomorrow'
                    THEN 'What do I have tomorrow?'
                WHEN elem->>'command' = 'todays_agenda'
                    THEN 'What do I have today?'
                WHEN elem->>'command' = 'tasks_due'
                    THEN 'What''s due today?'
                WHEN elem->>'command' = 'chat_message'
                    THEN coalesce(elem->'variables'->>'message', 'Hi')
                ELSE coalesce(elem->>'command', 'Hi')
            END AS label
        FROM jsonb_array_elements(commands) WITH ORDINALITY AS t(elem, ord)
    ) mapped
)
WHERE commands IS NOT NULL;
