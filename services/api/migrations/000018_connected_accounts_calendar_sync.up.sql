-- Calendar sync strategy: first-class sync token + observability on connected_accounts.
-- last_synced_at remains last successful sync. Source of truth for queries stays Donna DB.

ALTER TABLE connected_accounts
    ADD COLUMN IF NOT EXISTS calendar_list_sync_token text NULL,
    ADD COLUMN IF NOT EXISTS calendar_sync_status text NOT NULL DEFAULT 'idle',
    ADD COLUMN IF NOT EXISTS last_failed_sync_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS last_sync_duration_ms integer NULL,
    ADD COLUMN IF NOT EXISTS last_sync_created_count integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_sync_updated_count integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_sync_deleted_count integer NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'connected_accounts_calendar_sync_status_check'
    ) THEN
        ALTER TABLE connected_accounts
            ADD CONSTRAINT connected_accounts_calendar_sync_status_check
            CHECK (calendar_sync_status IN ('idle', 'running', 'succeeded', 'failed'));
    END IF;
END $$;

-- Promote any early sync-token storage out of provider_metadata.
UPDATE connected_accounts
SET calendar_list_sync_token = NULLIF(provider_metadata ->> 'calendar_list_sync_token', '')
WHERE calendar_list_sync_token IS NULL
  AND provider_metadata ? 'calendar_list_sync_token';

UPDATE connected_accounts
SET provider_metadata = provider_metadata - 'calendar_list_sync_token' - 'calendar_list_synced_at'
WHERE provider_metadata ? 'calendar_list_sync_token'
   OR provider_metadata ? 'calendar_list_synced_at';

COMMENT ON COLUMN connected_accounts.calendar_list_sync_token IS 'Google calendarList syncToken; cleared on HTTP 410 then rebuilt by full sync.';
COMMENT ON COLUMN connected_accounts.calendar_sync_status IS 'idle | running | succeeded | failed.';
COMMENT ON COLUMN connected_accounts.last_synced_at IS 'Last successful calendar sources sync (UTC).';
COMMENT ON COLUMN connected_accounts.last_failed_sync_at IS 'Last failed calendar sources sync attempt (UTC).';
COMMENT ON COLUMN connected_accounts.last_sync_duration_ms IS 'Duration of the last sync attempt in milliseconds.';
COMMENT ON COLUMN connected_accounts.last_sync_created_count IS 'Records created in the last successful sync.';
COMMENT ON COLUMN connected_accounts.last_sync_updated_count IS 'Records updated in the last successful sync.';
COMMENT ON COLUMN connected_accounts.last_sync_deleted_count IS 'Records soft-deleted in the last successful sync.';
