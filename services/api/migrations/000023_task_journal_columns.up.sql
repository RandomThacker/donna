-- Daily Task Journal: extend tasks with project + labels.
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS project text NULL,
    ADD COLUMN IF NOT EXISTS labels text[] NOT NULL DEFAULT '{}';

COMMENT ON COLUMN tasks.project IS 'Optional project or area name.';
COMMENT ON COLUMN tasks.labels IS 'Optional label tags for the permanent task.';
