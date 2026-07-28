-- Event sync fields beyond the base calendar_events shape.
-- sync_token for events remains on calendar_sources.sync_cursor (per source).

ALTER TABLE calendar_events
    ADD COLUMN IF NOT EXISTS timezone text NULL,
    ADD COLUMN IF NOT EXISTS organizer_summary jsonb NULL,
    ADD COLUMN IF NOT EXISTS provider_updated_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS provider_recurring_event_id text NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'calendar_events_organizer_summary_object_check'
    ) THEN
        ALTER TABLE calendar_events
            ADD CONSTRAINT calendar_events_organizer_summary_object_check
            CHECK (organizer_summary IS NULL OR jsonb_typeof(organizer_summary) = 'object');
    END IF;
END $$;

COMMENT ON COLUMN calendar_events.timezone IS 'Event timezone (IANA), typically from Google start.timeZone.';
COMMENT ON COLUMN calendar_events.organizer_summary IS 'Thin organizer object (email, displayName, self); never secrets.';
COMMENT ON COLUMN calendar_events.provider_updated_at IS 'Provider last-modified instant (Google updated).';
COMMENT ON COLUMN calendar_events.provider_recurring_event_id IS 'Provider series id (Google recurringEventId); Donna FK is recurring_event_id.';
