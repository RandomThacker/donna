-- Wave 1: auth_identities (login IdP bindings; distinct from connected_accounts).
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § auth_identities

CREATE TABLE IF NOT EXISTS auth_identities (
    id                uuid        NOT NULL,
    public_id         text        NOT NULL,
    user_id           uuid        NOT NULL,
    provider          text        NOT NULL,
    provider_subject  text        NOT NULL,
    email             text        NULL,
    email_verified    boolean     NOT NULL DEFAULT false,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz NULL,

    CONSTRAINT auth_identities_pkey PRIMARY KEY (id),
    CONSTRAINT auth_identities_public_id_key UNIQUE (public_id),
    CONSTRAINT auth_identities_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT auth_identities_provider_not_empty_check CHECK (provider <> ''),
    CONSTRAINT auth_identities_provider_subject_not_empty_check CHECK (provider_subject <> ''),
    CONSTRAINT auth_identities_public_id_prefix_check CHECK (public_id LIKE 'aid_%')
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identities_provider_subject_live_uidx
    ON auth_identities (provider, provider_subject)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS auth_identities_user_id_live_idx
    ON auth_identities (user_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE auth_identities IS 'Login identity-provider bindings. Separate from integration connected_accounts.';
COMMENT ON COLUMN auth_identities.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN auth_identities.public_id IS 'Stable API identifier with aid_ prefix.';
COMMENT ON COLUMN auth_identities.user_id IS 'Owning Donna user.';
COMMENT ON COLUMN auth_identities.provider IS 'IdP name: google, apple, microsoft, etc.';
COMMENT ON COLUMN auth_identities.provider_subject IS 'Stable subject (sub) from the identity provider.';
COMMENT ON COLUMN auth_identities.email IS 'Email captured at link time (optional).';
COMMENT ON COLUMN auth_identities.email_verified IS 'Whether IdP asserted email verification.';
COMMENT ON COLUMN auth_identities.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN auth_identities.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN auth_identities.deleted_at IS 'Soft-unlink marker; NULL means live.';
