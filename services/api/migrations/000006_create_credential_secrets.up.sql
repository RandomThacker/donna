-- Wave 1 addendum: sealed OAuth credential blobs referenced by connected_accounts.credentials_ref.
-- Raw tokens are never stored on connected_accounts.

CREATE TABLE IF NOT EXISTS credential_secrets (
    id           uuid        NOT NULL,
    ref          text        NOT NULL,
    ciphertext   bytea       NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz NULL,

    CONSTRAINT credential_secrets_pkey PRIMARY KEY (id),
    CONSTRAINT credential_secrets_ref_key UNIQUE (ref),
    CONSTRAINT credential_secrets_ref_prefix_check CHECK (ref LIKE 'cred_%'),
    CONSTRAINT credential_secrets_ciphertext_not_empty_check CHECK (octet_length(ciphertext) > 0)
);

CREATE INDEX IF NOT EXISTS credential_secrets_live_idx
    ON credential_secrets (ref)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE credential_secrets IS 'AES-GCM sealed provider tokens; referenced by connected_accounts.credentials_ref.';
COMMENT ON COLUMN credential_secrets.ref IS 'Stable credentials_ref value (cred_…).';
COMMENT ON COLUMN credential_secrets.ciphertext IS 'Nonce-prefixed AES-GCM ciphertext; never log or return to clients.';
