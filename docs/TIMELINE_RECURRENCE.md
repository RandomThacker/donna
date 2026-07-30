# Timeline Recurrence (Phase 2.1)

Donna stores **one base row** per recurring event or reminder. Occurrences are **never persisted**.

## Storage

- `donna_events.recurrence_rule` / `donna_reminders.recurrence_rule`
- iCalendar RRULE only (e.g. `FREQ=WEEKLY;BYDAY=WE,FR`)
- Optional `RRULE:` prefix accepted on write; stored without the prefix

## Expansion

On `GET /timeline?from=&to=`:

1. Load base Donna rows that can produce occurrences in the window  
   (recurring series with `start_at` / `trigger_at` before `to`)
2. Expand RRULE inside `[from, to)` with [`teambition/rrule-go`](https://github.com/teambition/rrule-go)
3. Respect the item’s timezone (`DTSTART` in that location)
4. Cap at 2000 occurrences per series per query

## Timeline item fields

| Field | Meaning |
| --- | --- |
| `is_recurring` | Expanded from RRULE |
| `parent_id` | Base Donna event/reminder id |
| `occurrence_id` | Stable virtual id: `{parent}_{YYYYMMDDTHHMMSSZ}` |
| `occurrence_start` / `occurrence_end` | This instance’s window |
| `recurrence_rule` | Normalized RRULE |

Non-recurring items keep `occurrence_id = id` and omit parent/occurrence timestamps.

## Validation

Invalid RRULE on create/update → `400 VALIDATION_ERROR`.

## Out of scope

Notifications, EXDATE/RDATE editing UI, Google/ICS recurrence expansion (provider events stay as synced instances).
