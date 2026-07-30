-- Web Push subscriptions (Phase 2.3). Multiple devices per user.

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id           uuid        NOT NULL,
    public_id    text        NOT NULL,
    user_id      uuid        NOT NULL,
    endpoint     text        NOT NULL,
    p256dh       text        NOT NULL,
    auth         text        NOT NULL,
    user_agent   text        NULL,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    deleted_at   timestamptz NULL,

    CONSTRAINT push_subscriptions_pkey PRIMARY KEY (id),
    CONSTRAINT push_subscriptions_public_id_key UNIQUE (public_id),
    CONSTRAINT push_subscriptions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT push_subscriptions_public_id_prefix_check CHECK (public_id LIKE 'psub_%'),
    CONSTRAINT push_subscriptions_endpoint_not_empty_check CHECK (endpoint <> ''),
    CONSTRAINT push_subscriptions_p256dh_not_empty_check CHECK (p256dh <> ''),
    CONSTRAINT push_subscriptions_auth_not_empty_check CHECK (auth <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS push_subscriptions_user_id_endpoint_live_uidx
    ON push_subscriptions (user_id, endpoint)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS push_subscriptions_user_id_live_idx
    ON push_subscriptions (user_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE push_subscriptions IS 'Browser Web Push subscriptions; multiple devices per user.';
COMMENT ON COLUMN push_subscriptions.public_id IS 'Stable API identifier with psub_ prefix.';
COMMENT ON COLUMN push_subscriptions.endpoint IS 'Push service endpoint URL.';
COMMENT ON COLUMN push_subscriptions.p256dh IS 'Client public key (base64url).';
COMMENT ON COLUMN push_subscriptions.auth IS 'Auth secret (base64url).';
COMMENT ON COLUMN push_subscriptions.user_agent IS 'Optional client user-agent at subscribe time.';
COMMENT ON COLUMN push_subscriptions.deleted_at IS 'Soft-delete on unsubscribe.';
