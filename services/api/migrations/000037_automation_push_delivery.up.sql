-- Automations: enable Web Push delivery alongside chat.
ALTER TABLE automations
    ALTER COLUMN delivery_channels SET DEFAULT '["chat","push"]'::jsonb;

UPDATE automations
SET
    delivery_channels = (
        SELECT COALESCE(jsonb_agg(to_jsonb(ch)), '["chat","push"]'::jsonb)
        FROM (
            SELECT DISTINCT ch
            FROM (
                SELECT jsonb_array_elements_text(delivery_channels) AS ch
                UNION ALL
                SELECT 'push'
            ) expanded
            WHERE ch IN ('chat', 'push', 'telegram', 'whatsapp', 'email')
        ) uniq
    ),
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM jsonb_array_elements_text(delivery_channels) AS t(ch)
      WHERE t.ch = 'push'
  );

COMMENT ON COLUMN automations.delivery_channels IS
    'Delivery targets. Supported: chat, push. Future: telegram, whatsapp, email.';
