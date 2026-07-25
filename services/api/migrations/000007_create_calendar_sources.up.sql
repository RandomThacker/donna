-- Domain completion: calendar_sources.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § calendar_sources
-- Also attaches deferred FKs from user_settings.default_*_calendar_source_id.

CREATE TABLE IF NOT EXISTS calendar_sources (
    id                       uuid        NOT NULL,
    public_id                text        NOT NULL,
    user_id                  uuid        NOT NULL,
    connected_account_id     uuid        NOT NULL,
    provider_calendar_id     text        NOT NULL,
    name                     text        NOT NULL,
    color                    text        NULL,
    is_primary_on_provider   boolean     NOT NULL DEFAULT false,
    sync_enabled             boolean     NOT NULL DEFAULT true,
    sync_cursor              text        NULL,
    last_synced_at           timestamptz NULL,
    timezone                 text        NULL,
    provider_metadata        jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    deleted_at               timestamptz NULL,

    CONSTRAINT calendar_sources_pkey PRIMARY KEY (id),
    CONSTRAINT calendar_sources_public_id_key UNIQUE (public_id),
    CONSTRAINT calendar_sources_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_sources_connected_account_id_fkey
        FOREIGN KEY (connected_account_id) REFERENCES connected_accounts (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_sources_public_id_prefix_check CHECK (public_id LIKE 'cal_%'),
    CONSTRAINT calendar_sources_name_not_empty_check CHECK (name <> ''),
    CONSTRAINT calendar_sources_provider_metadata_object_check
        CHECK (jsonb_typeof(provider_metadata) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS calendar_sources_account_provider_calendar_live_uidx
    ON calendar_sources (connected_account_id, provider_calendar_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS calendar_sources_user_id_sync_enabled_live_idx
    ON calendar_sources (user_id, sync_enabled)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS calendar_sources_connected_account_id_live_idx
    ON calendar_sources (connected_account_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE calendar_sources IS 'Synced calendar feed under a connected account.';
COMMENT ON COLUMN calendar_sources.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN calendar_sources.public_id IS 'Stable API identifier with cal_ prefix.';
COMMENT ON COLUMN calendar_sources.user_id IS 'Denormalized tenancy; must match connected_accounts.user_id (app-enforced).';
COMMENT ON COLUMN calendar_sources.connected_account_id IS 'Parent integration account.';
COMMENT ON COLUMN calendar_sources.provider_calendar_id IS 'Remote calendar id from the provider.';
COMMENT ON COLUMN calendar_sources.name IS 'Display name.';
COMMENT ON COLUMN calendar_sources.color IS 'Optional UI color hint.';
COMMENT ON COLUMN calendar_sources.is_primary_on_provider IS 'Whether this is the provider primary calendar.';
COMMENT ON COLUMN calendar_sources.sync_enabled IS 'Donna sync toggle.';
COMMENT ON COLUMN calendar_sources.sync_cursor IS 'Incremental sync token / cursor.';
COMMENT ON COLUMN calendar_sources.last_synced_at IS 'Last successful sync tick.';
COMMENT ON COLUMN calendar_sources.timezone IS 'Optional source timezone hint.';
COMMENT ON COLUMN calendar_sources.provider_metadata IS 'Non-secret provider quirks (jsonb object).';
COMMENT ON COLUMN calendar_sources.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN calendar_sources.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN calendar_sources.deleted_at IS 'Soft-delete marker; NULL means live.';

-- Deferred FKs from user_settings (columns already exist from 000005).
ALTER TABLE user_settings
    ADD CONSTRAINT user_settings_default_personal_calendar_source_id_fkey
        FOREIGN KEY (default_personal_calendar_source_id) REFERENCES calendar_sources (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    ADD CONSTRAINT user_settings_default_work_calendar_source_id_fkey
        FOREIGN KEY (default_work_calendar_source_id) REFERENCES calendar_sources (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    ADD CONSTRAINT user_settings_default_reminder_calendar_source_id_fkey
        FOREIGN KEY (default_reminder_calendar_source_id) REFERENCES calendar_sources (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION;

COMMENT ON COLUMN user_settings.default_personal_calendar_source_id IS 'Optional default personal calendar source.';
COMMENT ON COLUMN user_settings.default_work_calendar_source_id IS 'Optional default work calendar source.';
COMMENT ON COLUMN user_settings.default_reminder_calendar_source_id IS 'Optional default reminder calendar source.';
