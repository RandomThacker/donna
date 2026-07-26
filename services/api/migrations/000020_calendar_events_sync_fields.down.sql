ALTER TABLE calendar_events
    DROP CONSTRAINT IF EXISTS calendar_events_organizer_summary_object_check;

ALTER TABLE calendar_events
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS organizer_summary,
    DROP COLUMN IF EXISTS provider_updated_at,
    DROP COLUMN IF EXISTS provider_recurring_event_id;
