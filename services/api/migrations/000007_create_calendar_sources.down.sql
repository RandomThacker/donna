-- Rollback: calendar_sources (+ deferred user_settings FKs)
ALTER TABLE user_settings
    DROP CONSTRAINT IF EXISTS user_settings_default_reminder_calendar_source_id_fkey,
    DROP CONSTRAINT IF EXISTS user_settings_default_work_calendar_source_id_fkey,
    DROP CONSTRAINT IF EXISTS user_settings_default_personal_calendar_source_id_fkey;

DROP TABLE IF EXISTS calendar_sources;
