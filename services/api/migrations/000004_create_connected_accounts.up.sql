-- Wave 1: connected_accounts (integration OAuth accounts; not login identity).
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § connected_accounts

CREATE TABLE IF NOT EXISTS connected_accounts (
    id                   uuid        NOT NULL,
    public_id            text        NOT NULL,
    user_id              uuid        NOT NULL,
    provider             text        NOT NULL,
    provider_account_id  text        NOT NULL,
    display_name         text        NULL,
    status               text        NOT NULL DEFAULT 'active',
    scopes               text[]      NOT NULL DEFAULT '{}',
    credentials_ref      text        NOT NULL,
    token_expires_at     timestamptz NULL,
    last_synced_at       timestamptz NULL,
    provider_metadata    jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,

    CONSTRAINT connected_accounts_pkey PRIMARY KEY (id),
    CONSTRAINT connected_accounts_public_id_key UNIQUE (public_id),
    CONSTRAINT connected_accounts_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT connected_accounts_status_check
        CHECK (status IN ('active', 'needs_reauth', 'revoked', 'disconnected')),
    CONSTRAINT connected_accounts_public_id_prefix_check CHECK (public_id LIKE 'acct_%'),
    CONSTRAINT connected_accounts_provider_not_empty_check CHECK (provider <> ''),
    CONSTRAINT connected_accounts_credentials_ref_not_empty_check CHECK (credentials_ref <> ''),
    CONSTRAINT connected_accounts_provider_metadata_object_check
        CHECK (jsonb_typeof(provider_metadata) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS connected_accounts_provider_account_live_uidx
    ON connected_accounts (provider, provider_account_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS connected_accounts_user_id_status_live_idx
    ON connected_accounts (user_id, status)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE connected_accounts IS 'Third-party integration accounts (calendar sync, etc.). Not used for login.';
COMMENT ON COLUMN connected_accounts.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN connected_accounts.public_id IS 'Stable API identifier with acct_ prefix.';
COMMENT ON COLUMN connected_accounts.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN connected_accounts.provider IS 'Integration provider: google, microsoft, apple, etc.';
COMMENT ON COLUMN connected_accounts.provider_account_id IS 'Stable remote account identifier from the provider.';
COMMENT ON COLUMN connected_accounts.display_name IS 'Optional user-facing label (e.g. Work Google).';
COMMENT ON COLUMN connected_accounts.status IS 'active | needs_reauth | revoked | disconnected.';
COMMENT ON COLUMN connected_accounts.scopes IS 'Granted OAuth scopes as text array.';
COMMENT ON COLUMN connected_accounts.credentials_ref IS 'Reference to encrypted secret material / vault key; never store raw tokens here.';
COMMENT ON COLUMN connected_accounts.token_expires_at IS 'Optional access-token expiry if tracked.';
COMMENT ON COLUMN connected_accounts.last_synced_at IS 'Last successful sync tick.';
COMMENT ON COLUMN connected_accounts.provider_metadata IS 'Non-secret provider quirks (jsonb object).';
COMMENT ON COLUMN connected_accounts.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN connected_accounts.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN connected_accounts.deleted_at IS 'Soft-disconnect marker; NULL means live.';
