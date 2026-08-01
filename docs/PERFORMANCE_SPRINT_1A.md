# Performance Sprint 1A — Minimize Database Transfer

**Status:** Done (Occurrence / Notification Scheduler path only)  
**Goal:** Cut Neon network transfer on scheduler reads without changing Occurrence or notification behaviour.  
**Out of scope:** Timeline APIs, frontend, Action Layer, recurrence algorithms, Microsoft ICS provider (auth disabled).

---

## Summary

| Path | Before | After |
| --- | --- | --- |
| Google Occurrence | Wide `ListByUserInRangeWithProvider` (all providers + `provider_payload`) | Narrow `ListForSchedulerByUserInRange` filtered to `google` |
| Donna Event Occurrence | Full `donna_events` row | Narrow scheduler projection |
| Donna Reminder Occurrence | Full `donna_reminders` row | Narrow scheduler projection |
| Microsoft ICS Occurrence | Wide query (unchanged) | Unchanged |
| Timeline | Wide queries | Unchanged |

---

## Per-provider analysis

### GoogleOccurrenceProvider

**Old query:** `sqlSelectCalendarEventsByUserRangeWithProvider`

- Selected: full `calendarEventColumnsAliased` (26 event fields) + `ca.provider` + `s.color`
- Included: `provider_payload`, `attendees_summary`, `organizer_summary`, etags, recurrence sync ids, timestamps
- Then filtered `provider == google` in Go (after transferring Microsoft/ICS rows too)

**New query:** `sqlSelectCalendarEventsForScheduler`

```sql
SELECT e.id, e.public_id, e.user_id, e.calendar_source_id, e.title, e.description, e.location,
       e.starts_at, e.ends_at, e.status, e.timezone, e.provider_event_id, e.origin, ca.provider
FROM calendar_events e
JOIN calendar_sources s ON s.id = e.calendar_source_id
JOIN connected_accounts ca ON ca.id = s.connected_account_id
WHERE e.user_id = $1
  AND e.deleted_at IS NULL AND s.deleted_at IS NULL AND ca.deleted_at IS NULL
  AND e.starts_at < $3 AND e.ends_at > $2
  AND ca.provider = ANY($4::text[])   -- ['google']
ORDER BY e.starts_at ASC
```

**Columns removed:** `provider_payload`, `attendees_summary`, `organizer_summary`, `is_all_day`, `visibility`, `recurrence_rule`, `recurring_event_id`, `provider_recurring_event_id`, `provider_etag`, `provider_updated_at`, `created_at`, `updated_at`, `deleted_at`, `s.color`

**Est. bytes/row:** ~3–15 KB → **~500 B** (~85–97% reduction per Google row)  
**Extra win:** no longer ships non-Google calendar rows on this provider.

---

### DonnaEventOccurrenceProvider

**Old:** all 17 `donna_events` columns  

**New:** `id, public_id, user_id, title, description, start_at, end_at, timezone, location, reminder_offset_minutes, recurrence_rule, status`

**Removed:** `all_day`, `color`, `created_at`, `updated_at`, `deleted_at`  

**Est. bytes/row:** ~0.3–2 KB → **~400 B** (~10–25% typical; more when color/description heavy)

WHERE clause unchanged (one-shot overlap + open recurring `start_at < $to`).

---

### DonnaReminderOccurrenceProvider

**Old:** all 13 `donna_reminders` columns  

**New:** `id, public_id, user_id, title, description, trigger_at, timezone, recurrence_rule, status`

**Removed:** `color`, `created_at`, `updated_at`, `deleted_at`  

**Est. bytes/row:** ~0.2–1 KB → **~300 B**

WHERE clause unchanged.

---

### MicrosoftICSOccurrenceProvider

**Intentionally untouched** in Sprint 1A (Microsoft auth disabled). Still uses wide Timeline-style `ListByUserInRangeWithProvider`.

---

## Behaviour guarantees

- Occurrence mapping (`mapCalendarEvent` / Donna expand) only reads the retained columns → identical IDs, times, source, status, RRULE expansion, metadata used for notifications.
- Timeline continues to use wide SELECTs (`ListByUserInRange*`).
- Notification policy offsets and enqueue semantics unchanged.

## Instrumentation

Optimized providers log (module `notification`):

- `provider`
- `columns_selected` / `column_count`
- `est_bytes_per_row` / `rows_returned` / `est_bytes_total`
- `duration_ms`

Event: `occurrence provider query`

## Tests

- `repository.TestSchedulerSQLOmitsHeavyColumns`
- `provider.TestNarrow*ProducesIdenticalOccurrence*`
- `provider.TestGoogleSchedulerPathMatchesWideFilterBehavior`
- Existing provider / notification / occurrence suites

## Follow-ups (not this sprint)

1. Narrow Microsoft ICS when enabled  
2. Bound Donna recurring `start_at < $to` open-ended series fetch  
3. Optional single shared calendar query for all providers (further 2× avoidance once MS is optimized)
