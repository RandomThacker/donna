-- User-defined colored tags for tasks.
CREATE TABLE IF NOT EXISTS task_tags (
    id          uuid        NOT NULL,
    public_id   text        NOT NULL,
    user_id     uuid        NOT NULL,
    name        text        NOT NULL,
    color       text        NOT NULL,
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL,

    CONSTRAINT task_tags_pkey PRIMARY KEY (id),
    CONSTRAINT task_tags_public_id_key UNIQUE (public_id),
    CONSTRAINT task_tags_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT task_tags_public_id_prefix_check CHECK (public_id LIKE 'tag_%'),
    CONSTRAINT task_tags_name_not_empty_check CHECK (name <> ''),
    CONSTRAINT task_tags_color_check CHECK (color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS task_tags_user_id_name_lower_uidx
    ON task_tags (user_id, lower(name));

CREATE TABLE IF NOT EXISTS task_tag_assignments (
    task_id     uuid        NOT NULL,
    tag_id      uuid        NOT NULL,
    created_at  timestamptz NOT NULL,

    CONSTRAINT task_tag_assignments_pkey PRIMARY KEY (task_id, tag_id),
    CONSTRAINT task_tag_assignments_task_id_fkey
        FOREIGN KEY (task_id) REFERENCES tasks (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION,
    CONSTRAINT task_tag_assignments_tag_id_fkey
        FOREIGN KEY (tag_id) REFERENCES task_tags (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION
);

CREATE INDEX IF NOT EXISTS task_tag_assignments_tag_id_idx
    ON task_tag_assignments (tag_id);

COMMENT ON TABLE task_tags IS 'User-owned colored labels for tasks.';
COMMENT ON COLUMN task_tags.public_id IS 'Stable API identifier with tag_ prefix.';
COMMENT ON COLUMN task_tags.color IS 'Hex color (#RRGGBB) chosen by the user.';
COMMENT ON TABLE task_tag_assignments IS 'Many-to-many link between tasks and tags.';
