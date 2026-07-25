-- Wave 1: user_settings (1:1 preferences).
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § user_settings
--
-- Columns for default_*_calendar_source_id are present as designed.
-- Foreign keys to calendar_sources are deferred until that table exists
-- (dependency order); they will be added in the calendar_sources migration wave.

CREATE TABLE IF NOT EXISTS user_settings (
    id                                    uuid        NOT NULL,
    public_id                             text        NOT NULL,
    user_id                               uuid        NOT NULL,
    default_personal_calendar_source_id   uuid        NULL,
    default_work_calendar_source_id       uuid        NULL,
    default_reminder_calendar_source_id   uuid        NULL,
    morning_briefing_time                 time        NULL,
    midday_checkin_time                   time        NULL,
    evening_reflection_time               time        NULL,
    quiet_hours_start                     time        NULL,
    quiet_hours_end                       time        NULL,
    notifications_enabled                 boolean     NOT NULL DEFAULT true,
    preferences                           jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at                            timestamptz NOT NULL DEFAULT now(),
    updated_at                            timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT user_settings_pkey PRIMARY KEY (id),
    CONSTRAINT user_settings_public_id_key UNIQUE (public_id),
    CONSTRAINT user_settings_user_id_key UNIQUE (user_id),
    CONSTRAINT user_settings_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION,
    CONSTRAINT user_settings_public_id_prefix_check CHECK (public_id LIKE 'set_%'),
    CONSTRAINT user_settings_preferences_object_check
        CHECK (jsonb_typeof(preferences) = 'object')
);

COMMENT ON TABLE user_settings IS 'Per-user preferences and default calendar routing (1:1 with users).';
COMMENT ON COLUMN user_settings.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN user_settings.public_id IS 'Stable API identifier with set_ prefix.';
COMMENT ON COLUMN user_settings.user_id IS 'Owning user; exactly one settings row per user.';
COMMENT ON COLUMN user_settings.default_personal_calendar_source_id IS 'Optional default personal calendar source. FK to calendar_sources deferred until that table exists.';
COMMENT ON COLUMN user_settings.default_work_calendar_source_id IS 'Optional default work calendar source. FK deferred until calendar_sources exists.';
COMMENT ON COLUMN user_settings.default_reminder_calendar_source_id IS 'Optional default reminder calendar source. FK deferred until calendar_sources exists.';
COMMENT ON COLUMN user_settings.morning_briefing_time IS 'Local time-of-day for morning briefing; interpret with users.timezone.';
COMMENT ON COLUMN user_settings.midday_checkin_time IS 'Local time-of-day for midday check-in.';
COMMENT ON COLUMN user_settings.evening_reflection_time IS 'Local time-of-day for evening reflection.';
COMMENT ON COLUMN user_settings.quiet_hours_start IS 'Quiet hours start (local time-of-day).';
COMMENT ON COLUMN user_settings.quiet_hours_end IS 'Quiet hours end (local time-of-day).';
COMMENT ON COLUMN user_settings.notifications_enabled IS 'Master notification toggle.';
COMMENT ON COLUMN user_settings.preferences IS 'Sparse non-core preferences (jsonb object); allow-listed keys in application.';
COMMENT ON COLUMN user_settings.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN user_settings.updated_at IS 'Last mutation time; maintained by application.';
