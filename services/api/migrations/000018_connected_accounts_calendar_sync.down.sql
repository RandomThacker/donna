-- Rollback: connected_accounts calendar sync observability columns

ALTER TABLE connected_accounts
    DROP CONSTRAINT IF EXISTS connected_accounts_calendar_sync_status_check;

ALTER TABLE connected_accounts
    DROP COLUMN IF EXISTS last_sync_deleted_count,
    DROP COLUMN IF EXISTS last_sync_updated_count,
    DROP COLUMN IF EXISTS last_sync_created_count,
    DROP COLUMN IF EXISTS last_sync_duration_ms,
    DROP COLUMN IF EXISTS last_failed_sync_at,
    DROP COLUMN IF EXISTS calendar_sync_status,
    DROP COLUMN IF EXISTS calendar_list_sync_token;
