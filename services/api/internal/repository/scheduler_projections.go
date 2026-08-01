package repository

// Scheduler-narrow projections for OccurrenceService (Performance Sprint 1A).
// Timeline paths keep the wide calendarEventColumns / donna*Columns lists.

const calendarEventSchedulerColumnsAliased = `
	e.id, e.public_id, e.user_id, e.calendar_source_id, e.title, e.description, e.location,
	e.starts_at, e.ends_at, e.status, e.timezone, e.provider_event_id, e.origin`

const donnaEventSchedulerColumns = `
	id, public_id, user_id, title, description, start_at, end_at, timezone,
	location, reminder_offset_minutes, recurrence_rule, status`

const donnaReminderSchedulerColumns = `
	id, public_id, user_id, title, description, trigger_at, timezone,
	recurrence_rule, status`

const (
	// sqlSelectCalendarEventsForScheduler is the Occurrence/scheduler path.
	// Omits provider_payload, attendees, organizer, etag, recurrence sync fields, color.
	sqlSelectCalendarEventsForScheduler = `
SELECT` + calendarEventSchedulerColumnsAliased + `, ca.provider
FROM calendar_events e
JOIN calendar_sources s ON s.id = e.calendar_source_id
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE e.user_id = $1
  AND e.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND ca.deleted_at IS NULL
  AND e.starts_at < $3
  AND e.ends_at > $2
  AND ca.provider = ANY($4::text[])
ORDER BY e.starts_at ASC`

	sqlListDonnaEventsForScheduler = `
SELECT` + donnaEventSchedulerColumns + `
FROM donna_events
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status <> 'cancelled'
  AND (
    (
      (recurrence_rule IS NULL OR btrim(recurrence_rule) = '')
      AND start_at < $3
      AND end_at > $2
    )
    OR (
      recurrence_rule IS NOT NULL AND btrim(recurrence_rule) <> ''
      AND start_at < $3
    )
  )
ORDER BY start_at ASC`

	sqlListDonnaRemindersForScheduler = `
SELECT` + donnaReminderSchedulerColumns + `
FROM donna_reminders
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status <> 'cancelled'
  AND (
    (
      (recurrence_rule IS NULL OR btrim(recurrence_rule) = '')
      AND trigger_at >= $2
      AND trigger_at < $3
    )
    OR (
      recurrence_rule IS NOT NULL AND btrim(recurrence_rule) <> ''
      AND trigger_at < $3
    )
  )
ORDER BY trigger_at ASC`
)

// Scheduler projection column names (for provider metrics / docs).
var (
	CalendarEventSchedulerColumnNames = []string{
		"id", "public_id", "user_id", "calendar_source_id", "title", "description", "location",
		"starts_at", "ends_at", "status", "timezone", "provider_event_id", "origin", "provider",
	}
	DonnaEventSchedulerColumnNames = []string{
		"id", "public_id", "user_id", "title", "description", "start_at", "end_at", "timezone",
		"location", "reminder_offset_minutes", "recurrence_rule", "status",
	}
	DonnaReminderSchedulerColumnNames = []string{
		"id", "public_id", "user_id", "title", "description", "trigger_at", "timezone",
		"recurrence_rule", "status",
	}
)
