-- Timeline-derived notification queue (Phase 2.2).
-- Evolves existing notifications for occurrence-based scheduling.
-- Overall status tracks lifecycle; channel_delivery_status tracks per-channel delivery (future).

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_delivery_status_check;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_sent_at_check;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_channel_check;

ALTER TABLE notifications RENAME COLUMN delivery_status TO delivery_status_legacy;

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS timeline_item_parent_id text NULL,
    ADD COLUMN IF NOT EXISTS occurrence_id text NULL,
    ADD COLUMN IF NOT EXISTS notification_type text NULL,
    ADD COLUMN IF NOT EXISTS scheduled_for timestamptz NULL,
    ADD COLUMN IF NOT EXISTS status text NULL,
    ADD COLUMN IF NOT EXISTS delivery_channels text[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS channel_delivery_status jsonb NOT NULL DEFAULT '{}'::jsonb;

UPDATE notifications SET
    status = CASE
        WHEN dismissed_at IS NOT NULL THEN 'DISMISSED'
        WHEN read_at IS NOT NULL THEN 'READ'
        WHEN delivery_status_legacy = 'sent' THEN 'SENT'
        WHEN delivery_status_legacy = 'failed' THEN 'FAILED'
        ELSE 'PENDING'
    END
WHERE status IS NULL;

UPDATE notifications SET
    delivery_channels = CASE channel
        WHEN 'browser_push' THEN ARRAY['WEB_PUSH']::text[]
        WHEN 'telegram' THEN ARRAY['TELEGRAM']::text[]
        WHEN 'email' THEN ARRAY['WEB_PUSH']::text[]
        ELSE ARRAY['WEB_PUSH']::text[]
    END
WHERE cardinality(delivery_channels) = 0;

ALTER TABLE notifications
    ALTER COLUMN status SET DEFAULT 'PENDING';

UPDATE notifications SET status = 'PENDING' WHERE status IS NULL;

ALTER TABLE notifications
    ALTER COLUMN status SET NOT NULL;

ALTER TABLE notifications DROP COLUMN delivery_status_legacy;

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_status_check,
    ADD CONSTRAINT notifications_status_check
        CHECK (status IN ('PENDING', 'SENT', 'READ', 'DISMISSED', 'FAILED'));

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_type_check,
    ADD CONSTRAINT notifications_type_check
        CHECK (notification_type IS NULL OR notification_type IN ('EVENT', 'REMINDER'));

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_channel_delivery_status_object_check,
    ADD CONSTRAINT notifications_channel_delivery_status_object_check
        CHECK (jsonb_typeof(channel_delivery_status) = 'object');

-- Idempotency: one notification per occurrence + type (live rows).
CREATE UNIQUE INDEX IF NOT EXISTS notifications_occurrence_type_live_uidx
    ON notifications (occurrence_id, notification_type)
    WHERE deleted_at IS NULL
      AND occurrence_id IS NOT NULL
      AND notification_type IS NOT NULL;

CREATE INDEX IF NOT EXISTS notifications_status_scheduled_for_live_idx
    ON notifications (status, scheduled_for)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS notifications_user_id_status_live_idx
    ON notifications (user_id, status, scheduled_for DESC)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN notifications.timeline_item_parent_id IS 'Base timeline parent id (event/reminder); occurrence may differ.';
COMMENT ON COLUMN notifications.occurrence_id IS 'Virtual occurrence id; unique with notification_type for idempotent enqueue.';
COMMENT ON COLUMN notifications.notification_type IS 'EVENT | REMINDER.';
COMMENT ON COLUMN notifications.scheduled_for IS 'When the notification should fire (from NotificationPolicy).';
COMMENT ON COLUMN notifications.status IS 'PENDING | SENT | READ | DISMISSED | FAILED — overall lifecycle.';
COMMENT ON COLUMN notifications.delivery_channels IS 'Intended channels: WEB_PUSH | CHAT | TELEGRAM | WHATSAPP.';
COMMENT ON COLUMN notifications.channel_delivery_status IS 'Per-channel delivery map e.g. {"WEB_PUSH":"PENDING"}; delivery phase fills this.';
