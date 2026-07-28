-- Domain completion: calendar_events.
-- Source: docs/PHYSICAL_DATABASE_DESIGN.md § calendar_events
-- No provider column — derive via calendar_source → connected_account (R-07).

CREATE TABLE IF NOT EXISTS calendar_events (
    id                   uuid        NOT NULL,
    public_id            text        NOT NULL,
    user_id              uuid        NOT NULL,
    calendar_source_id   uuid        NOT NULL,
    title                text        NOT NULL,
    description          text        NULL,
    location             text        NULL,
    starts_at            timestamptz NOT NULL,
    ends_at              timestamptz NOT NULL,
    is_all_day           boolean     NOT NULL DEFAULT false,
    status               text        NOT NULL DEFAULT 'confirmed',
    visibility           text        NULL,
    attendees_summary    jsonb       NOT NULL DEFAULT '[]'::jsonb,
    recurrence_rule      text        NULL,
    recurring_event_id   uuid        NULL,
    provider_event_id    text        NULL,
    provider_etag        text        NULL,
    provider_payload     jsonb       NULL,
    origin               text        NOT NULL DEFAULT 'donna',
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz NULL,

    CONSTRAINT calendar_events_pkey PRIMARY KEY (id),
    CONSTRAINT calendar_events_public_id_key UNIQUE (public_id),
    CONSTRAINT calendar_events_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_events_calendar_source_id_fkey
        FOREIGN KEY (calendar_source_id) REFERENCES calendar_sources (id)
        ON DELETE RESTRICT
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_events_recurring_event_id_fkey
        FOREIGN KEY (recurring_event_id) REFERENCES calendar_events (id)
        ON DELETE SET NULL
        ON UPDATE NO ACTION,
    CONSTRAINT calendar_events_public_id_prefix_check CHECK (public_id LIKE 'evt_%'),
    CONSTRAINT calendar_events_title_not_empty_check CHECK (title <> ''),
    CONSTRAINT calendar_events_ends_after_starts_check CHECK (ends_at >= starts_at),
    CONSTRAINT calendar_events_status_check
        CHECK (status IN ('confirmed', 'tentative', 'cancelled')),
    CONSTRAINT calendar_events_visibility_check
        CHECK (visibility IS NULL OR visibility IN ('default', 'private', 'public')),
    CONSTRAINT calendar_events_origin_check
        CHECK (origin IN ('donna', 'provider_sync')),
    CONSTRAINT calendar_events_attendees_summary_array_check
        CHECK (jsonb_typeof(attendees_summary) = 'array'),
    CONSTRAINT calendar_events_provider_payload_object_check
        CHECK (provider_payload IS NULL OR jsonb_typeof(provider_payload) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS calendar_events_source_provider_event_live_uidx
    ON calendar_events (calendar_source_id, provider_event_id)
    WHERE deleted_at IS NULL AND provider_event_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS calendar_events_user_id_starts_at_live_idx
    ON calendar_events (user_id, starts_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS calendar_events_source_range_live_idx
    ON calendar_events (calendar_source_id, starts_at, ends_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS calendar_events_user_id_status_live_idx
    ON calendar_events (user_id, status)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE calendar_events IS 'Donna canonical calendar event; provider derived via source → account.';
COMMENT ON COLUMN calendar_events.id IS 'Internal primary key (UUIDv7, application-generated).';
COMMENT ON COLUMN calendar_events.public_id IS 'Stable API identifier with evt_ prefix.';
COMMENT ON COLUMN calendar_events.user_id IS 'Owning Donna user (tenancy).';
COMMENT ON COLUMN calendar_events.calendar_source_id IS 'Parent calendar source.';
COMMENT ON COLUMN calendar_events.title IS 'Event title.';
COMMENT ON COLUMN calendar_events.description IS 'Optional event body.';
COMMENT ON COLUMN calendar_events.location IS 'Optional location.';
COMMENT ON COLUMN calendar_events.starts_at IS 'Event start instant (UTC).';
COMMENT ON COLUMN calendar_events.ends_at IS 'Event end instant (UTC); must be >= starts_at.';
COMMENT ON COLUMN calendar_events.is_all_day IS 'All-day event flag.';
COMMENT ON COLUMN calendar_events.status IS 'confirmed | tentative | cancelled.';
COMMENT ON COLUMN calendar_events.visibility IS 'Optional visibility: default | private | public.';
COMMENT ON COLUMN calendar_events.attendees_summary IS 'Thin attendee list (jsonb array); size capped in application.';
COMMENT ON COLUMN calendar_events.recurrence_rule IS 'RRULE or recurrence summary text.';
COMMENT ON COLUMN calendar_events.recurring_event_id IS 'Optional series parent event.';
COMMENT ON COLUMN calendar_events.provider_event_id IS 'Remote event id when synced.';
COMMENT ON COLUMN calendar_events.provider_etag IS 'Provider sync version / etag.';
COMMENT ON COLUMN calendar_events.provider_payload IS 'Opaque provider snapshot (jsonb object); never secrets; never canonical.';
COMMENT ON COLUMN calendar_events.origin IS 'donna | provider_sync.';
COMMENT ON COLUMN calendar_events.created_at IS 'Row creation time (UTC).';
COMMENT ON COLUMN calendar_events.updated_at IS 'Last mutation time; maintained by application.';
COMMENT ON COLUMN calendar_events.deleted_at IS 'Soft-delete / sync tombstone; NULL means live.';
