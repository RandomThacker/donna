-- Rollback: calendar_sources is_writable / access_role columns

DROP INDEX IF EXISTS calendar_sources_user_id_writable_live_idx;

ALTER TABLE calendar_sources
    DROP COLUMN IF EXISTS access_role,
    DROP COLUMN IF EXISTS is_writable;
