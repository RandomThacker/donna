UPDATE automations
SET
    delivery_channels = (
        SELECT COALESCE(jsonb_agg(to_jsonb(ch)), '["chat"]'::jsonb)
        FROM (
            SELECT jsonb_array_elements_text(delivery_channels) AS ch
        ) expanded
        WHERE ch <> 'push'
    ),
    updated_at = NOW()
WHERE deleted_at IS NULL;

ALTER TABLE automations
    ALTER COLUMN delivery_channels SET DEFAULT '["chat"]'::jsonb;

COMMENT ON COLUMN automations.delivery_channels IS
    'Delivery targets. Phase 1: chat. Future: telegram, push, whatsapp, email.';
