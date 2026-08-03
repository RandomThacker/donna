# Performance Sprint 1B — Reduce Query Count

**Status:** Done (Occurrence / Notification Scheduler path only)  
**Depends on:** [PERFORMANCE_SPRINT_1A.md](./PERFORMANCE_SPRINT_1A.md)  
**Goal:** Avoid duplicate `calendar_events` reads without changing Occurrence or notification behaviour.

---

## Problem

After Sprint 1A, OccurrenceService still registered:

| Provider | Query |
| --- | --- |
| `GoogleOccurrenceProvider` | narrow `calendar_events` (`provider = google`) |
| `MicrosoftICSOccurrenceProvider` | **wide** `calendar_events` (all providers, filter in Go) |

Same table, same range, two round trips per user tick. Microsoft auth is disabled in product, so the second query was pure waste (and still transferred `provider_payload`).

---

## Solution

### Shared repository method

```go
ListCalendarOccurrences(ctx, userID, from, to, providers []string)
```

Alias of the Sprint 1A narrow projection with SQL:

```sql
AND ca.provider = ANY($4::text[])
```

Provider filtering stays in SQL — never fetch providers that were not requested.

### Shared Occurrence provider

`SharedCalendarOccurrenceProvider` issues **one** `ListCalendarOccurrences` call for `ActiveCalendarOccurrenceProviders` (currently `["google","ics","microsoft"]`), then maps rows to Occurrences in memory.

`OccurrenceService` wiring:

```text
Before: Google + MicrosoftICS + DonnaEvent + DonnaReminder  →  4 DB queries
After:  SharedCalendar + DonnaEvent + DonnaReminder         →  3 DB queries
```

Calendar portion: **2 → 1**.

When Microsoft/ICS are re-enabled, add them to `ActiveCalendarOccurrenceProviders` — still one query.

---

## Before / After metrics (calendar path)

Illustrative busy user with ~8 Google rows in the 35m window (MS unused):

| Metric | Before (1A dual providers) | After (1B shared) |
| --- | --- | --- |
| `calendar_events` queries | 2 | **1** |
| Rows transferred (Google path) | ~8 (narrow) | ~8 (narrow) |
| Rows transferred (Microsoft path) | ~8–N all-provider wide rows | **0** |
| Est. bytes (Google) | ~8 × 500 B ≈ 4 KB | ~4 KB |
| Est. bytes (Microsoft wide) | ~8 × 3–15 KB ≈ 24–120 KB | **0** |
| Est. calendar transfer | ~28–124 KB | **~4 KB** |
| Duration | 2 RTT | **1 RTT** |

Dominant win: eliminating the unused wide Microsoft ICS query, not compressing Google further.

Full OccurrenceService tick (1 user):

| | Before | After |
| --- | --- | --- |
| Total provider queries | 4 | **3** |
| Calendar queries | 2 | **1** |

Instrumentation (`occurrence calendar query consolidated`):

- `queries_executed` (= 1)
- `provider_filters`
- `rows_returned` / `est_bytes_total`
- `duration_ms`

---

## Behaviour guarantees

- Occurrences for Google rows match prior `GoogleOccurrenceProvider` output (same narrow map).
- Donna Event / Reminder providers unchanged.
- Timeline still uses separate wide providers (untouched).
- Notification policy / enqueue unchanged.

## Tests

- `TestSharedCalendarSingleQuery` — before 2 calls, after 1
- `TestSharedCalendarMatchesGoogleProviderOutput` — Occurrence equality
- `TestSharedCalendarMultiProviderFilterInSQL` — IN-list ready for MS/ICS
- `TestSharedCalendarDefaultsToActiveProviders`

## Follow-ups

1. ~~Enable `microsoft` / `ics` in `ActiveCalendarOccurrenceProviders`~~ — done (ICS work calendars were silently skipped for notifications)
2. Optional further consolidation of Donna tables is out of scope (different schemas)
