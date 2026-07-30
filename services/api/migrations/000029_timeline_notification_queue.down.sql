-- Rollback Phase 2.2 timeline notification queue columns.

DROP INDEX IF EXISTS notifications_user_id_status_live_idx;
DROP INDEX IF EXISTS notifications_status_scheduled_for_live_idx;
DROP INDEX IF EXISTS notifications_occurrence_type_live_uidx;

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_channel_delivery_status_object_check;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_status_check;

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS delivery_status text NOT NULL DEFAULT 'pending';

UPDATE notifications SET delivery_status = CASE status
    WHEN 'SENT' THEN 'sent'
    WHEN 'FAILED' THEN 'failed'
    ELSE 'pending'
END;

ALTER TABLE notifications
    DROP COLUMN IF EXISTS channel_delivery_status,
    DROP COLUMN IF EXISTS delivery_channels,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS scheduled_for,
    DROP COLUMN IF EXISTS notification_type,
    DROP COLUMN IF EXISTS occurrence_id,
    DROP COLUMN IF EXISTS timeline_item_parent_id;

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_delivery_status_check,
    ADD CONSTRAINT notifications_delivery_status_check
        CHECK (delivery_status IN ('pending', 'sent', 'failed'));

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_sent_at_check,
    ADD CONSTRAINT notifications_sent_at_check
        CHECK ((delivery_status <> 'sent') OR (sent_at IS NOT NULL));

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_channel_check,
    ADD CONSTRAINT notifications_channel_check
        CHECK (channel IN ('browser_push', 'email', 'telegram'));
