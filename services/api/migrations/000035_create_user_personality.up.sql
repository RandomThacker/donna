-- Phase: Personality Engine — one personality preference row per user.

CREATE TABLE IF NOT EXISTS user_personality (
    id                    uuid        NOT NULL,
    user_id               uuid        NOT NULL,
    personality_id        text        NOT NULL,
    display_name          text        NULL,
    nickname              text        NULL,
    emoji_level           text        NOT NULL DEFAULT 'none',
    humor_level           text        NOT NULL DEFAULT 'none',
    greeting_style        text        NOT NULL DEFAULT 'formal',
    encouragement_level   text        NOT NULL DEFAULT 'low',
    response_style        text        NOT NULL DEFAULT 'concise',
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT user_personality_pkey PRIMARY KEY (id),
    CONSTRAINT user_personality_user_id_key UNIQUE (user_id),
    CONSTRAINT user_personality_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION,
    CONSTRAINT user_personality_personality_id_check
        CHECK (personality_id IN ('professional', 'casual', 'flirty')),
    CONSTRAINT user_personality_emoji_level_check
        CHECK (emoji_level IN ('none', 'low', 'medium', 'high')),
    CONSTRAINT user_personality_humor_level_check
        CHECK (humor_level IN ('none', 'low', 'medium', 'high')),
    CONSTRAINT user_personality_encouragement_level_check
        CHECK (encouragement_level IN ('none', 'low', 'medium', 'high'))
);

CREATE INDEX IF NOT EXISTS user_personality_personality_id_idx
    ON user_personality (personality_id);

COMMENT ON TABLE user_personality IS 'Per-user Donna Personality Engine preferences (1:1 with users).';
COMMENT ON COLUMN user_personality.personality_id IS 'Built-in personality key: professional, casual, flirty.';
COMMENT ON COLUMN user_personality.display_name IS 'Preferred name for salutations; falls back to users.display_name.';
COMMENT ON COLUMN user_personality.nickname IS 'Optional nickname substituted as {nickname}.';

-- Existing users default to Professional.
INSERT INTO user_personality (
    id, user_id, personality_id, display_name, nickname,
    emoji_level, humor_level, greeting_style, encouragement_level, response_style,
    created_at, updated_at
)
SELECT
    u.id, -- reuse user id as row id for stable 1:1 seeding (app may replace with UUIDv7 on upsert)
    u.id,
    'professional',
    u.display_name,
    NULL,
    'none',
    'none',
    'formal',
    'low',
    'concise',
    now(),
    now()
FROM users u
WHERE u.deleted_at IS NULL
ON CONFLICT (user_id) DO NOTHING;
