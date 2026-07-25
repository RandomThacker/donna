-- Promote writable/access_role out of provider_metadata onto first-class columns.
-- writable is queried frequently (create-event routing, UI filters).

ALTER TABLE calendar_sources
    ADD COLUMN IF NOT EXISTS is_writable boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS access_role text NULL;

-- Backfill from any JSON that already stored these keys during early calendar sync.
UPDATE calendar_sources
SET
    is_writable = CASE
        WHEN COALESCE((provider_metadata ->> 'writable')::boolean, false) THEN true
        WHEN COALESCE(provider_metadata ->> 'access_role', '') IN ('owner', 'writer') THEN true
        ELSE is_writable
    END,
    access_role = COALESCE(
        NULLIF(provider_metadata ->> 'access_role', ''),
        access_role
    )
WHERE provider_metadata ? 'writable'
   OR provider_metadata ? 'access_role';

UPDATE calendar_sources
SET provider_metadata = (provider_metadata - 'writable' - 'access_role')
WHERE provider_metadata ? 'writable'
   OR provider_metadata ? 'access_role';

CREATE INDEX IF NOT EXISTS calendar_sources_user_id_writable_live_idx
    ON calendar_sources (user_id, is_writable)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN calendar_sources.is_writable IS 'Whether Donna can create/update events on this calendar (owner/writer).';
COMMENT ON COLUMN calendar_sources.access_role IS 'Provider access role (owner, writer, reader, freeBusyReader, …).';
