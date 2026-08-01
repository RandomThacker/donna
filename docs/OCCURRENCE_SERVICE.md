# Occurrence Service

**Status:** Live for Notification Scheduler enqueue  
**Package:** `services/api/internal/occurrence`  
**Entry point:** `Service.ListUpcoming`

---

## Why it exists (separate from TimelineService)

| | TimelineService | OccurrenceService |
| --- | --- | --- |
| **Job** | Show the day | Decide what can be scheduled / notified |
| **Output** | `TimelineItem` (presentation) | `Occurrence` (scheduling) |
| **Consumers** | HTTP timeline / calendar UI | Notification Scheduler, future Telegram / AI planner / jobs |
| **Knows about** | Colors, read-only, UI metadata | Identity, times, source, type, status |

Same repositories and shared `internal/recurrence` helpers underneath. Different product contracts on purpose: changing a timeline card must not rewrite notification enqueue.

---

## Notification Scheduler Migration

### Before

```text
NotificationScheduler
        │
        ▼
NotificationService.EnqueueForUser
        │
        ▼
TimelineService.List  →  TimelineItem
        │
        ▼
NotificationPolicy (TimelineItem)
        │
        ▼
Notification Queue (unchanged)
```

### After (Task 4)

```text
NotificationScheduler
        │
        ▼
NotificationService.EnqueueForUser
        │
        ▼
OccurrenceService.ListUpcoming  →  Occurrence
        │
        ▼
NotificationPolicy (Occurrence)
        │
        ▼
Notification Queue (unchanged)
```

### Dependency graph

```text
app.go
  ├─ TimelineService          → TimelineHandler / Actions (UI only)
  └─ OccurrenceService        → NotificationService → NotificationScheduler
         │
         ├─ GoogleOccurrenceProvider
         ├─ MicrosoftICSOccurrenceProvider
         ├─ DonnaEventOccurrenceProvider
         └─ DonnaReminderOccurrenceProvider
```

**Guarantees kept:** offsets (Google/MS 10m, Donna event 15m, reminder exact), occurrence ids, payload keys (`timelineItemId`, `occurrenceId`, `source`, `type`, …), idempotency on `(occurrence_id, notification_type)`, dispatcher / queue / notification APIs untouched.

**Explicit non-dependency:** `NotificationService` and `NotificationScheduler` no longer reference `TimelineService`.

---

## Responsibilities

OccurrenceService **does**:

1. Call every `OccurrenceProvider` for the requested window only
2. Expand remaining series templates (RRULE + no `ParentID`)
3. Normalize and validate scheduling fields
4. Sort chronologically with a stable source priority
5. Deduplicate by `OccurrenceID`
6. Return `[]Occurrence`

OccurrenceService **does not**:

- Apply `NotificationPolicy` (NotificationService does)
- Generate or persist notifications
- Convert to `TimelineItem`
- Know about UI, REST, or Actions

---

## Dependencies

```text
OccurrenceService
        │
        ▼
OccurrenceProvider(s)
        │
        ▼
Repositories (calendar_events, donna_events, donna_reminders)
        │
        ▼
internal/recurrence  (shared with Timeline providers)
```

No imports from Timeline, Notification, handler, or actions packages.

---

## Data flow (pipeline)

```text
Providers → Collector → Recurrence Expansion → Normalizer → Sorter → Deduplicator → []Occurrence
```

### Source priority (stable)

1. `GOOGLE`  2. `MICROSOFT_ICS`  3. `DONNA`  4. Unknown last

---

## Construction

```go
occurrenceSvc := occurrence.NewService(occurrence.ServiceDeps{
    Providers: []occurrence.Provider{
        provider.NewGoogleOccurrenceProvider(calendarEventsRepo),
        provider.NewMicrosoftICSOccurrenceProvider(calendarEventsRepo),
        provider.NewDonnaEventOccurrenceProvider(donnaEventRepo),
        provider.NewDonnaReminderOccurrenceProvider(donnaReminderRepo),
    },
})
notificationSvc := business.NewNotificationService(
    notificationRepo,
    occurrenceSvc,
    business.NewNotificationPolicyResolver(),
)
```

The `Provider` interface lives on the `occurrence` package (aliased as `provider.OccurrenceProvider`) to avoid an import cycle with implementations.

---

## Scheduler tick metrics

Every `NotificationScheduler.Tick` logs counters with `feed_source=occurrence`:

| Field | Meaning |
| --- | --- |
| `occurrences` | After provider collect |
| `after_expansion` | After service expansion stage |
| `after_dedup` | After service dedup stage |
| `notifications` | Newly created PENDING rows |
| `db_queries` / `alloc_bytes` / `duration_ms` | Cost signals |

Compare against pre-cutover `feed_source=timeline` samples for before/after measurement.
