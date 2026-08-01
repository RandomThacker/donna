# Occurrence Domain

**Status:** Model + providers + OccurrenceService — **Notification Scheduler cut over to OccurrenceService**  
**Packages:**
- `services/api/internal/occurrence` — scheduling model + OccurrenceService
- `services/api/internal/occurrence/provider` — OccurrenceProvider implementations
- `services/api/internal/recurrence` — shared RRULE helpers (Timeline + Occurrence)

See also: [OCCURRENCE_SERVICE.md](./OCCURRENCE_SERVICE.md).

---

## Why Occurrence exists

Donna has two different jobs that look similar but are not the same:

| Concern | Model | Audience |
| --- | --- | --- |
| **Show the day** | `TimelineItem` (`entity.TimelineItem`) | UI — dashboard, calendar, phone, agenda |
| **Decide what to notify** | `Occurrence` | Schedulers, notification enqueue, future delivery pipelines |

`TimelineItem` is a **UI / presentation aggregate**. It carries display fields (color, read-only, all-day presentation cues, rich metadata for cards) and is shaped for rendering a unified feed.

`Occurrence` is a **scheduling domain unit**. It is the concrete time-bound thing a background worker can evaluate: *who*, *what*, *when*, *from where*, *status* — without caring how a card looks.

Keeping them separate avoids:

1. **Coupling notification logic to UI shape** — schedulers should not depend on colors, icons, or frontend-only flags.
2. **Accidental API leakage** — changing a timeline response field should not rewrite notification enqueue.
3. **Wrong abstractions** — recurrence expansion and reminder policy need a stable scheduling vocabulary, not a view-model.

```text
Providers / Donna tables
        │
        ├──────────────► TimelineProvider ──► TimelineItem ──► HTTP / UI
        │
        └──────────────► OccurrenceProvider ──► OccurrenceService ──► Notification Scheduler
```

Today the Notification Scheduler still calls `Timeline.List()`. Occurrence providers exist so that pipeline can switch to a scheduling model without redesigning Timeline.

---

## TimelineProvider vs OccurrenceProvider

| | TimelineProvider | OccurrenceProvider |
| --- | --- | --- |
| **Package** | `internal/business` | `internal/occurrence/provider` |
| **Method** | `Fetch(...) ([]TimelineItem, error)` | `ListOccurrences(...) ([]Occurrence, error)` |
| **Output** | Presentation items (color, readOnly, all-day, UI metadata) | Scheduling units only |
| **Consumer** | `TimelineService` → HTTP timeline / calendar UI | `OccurrenceService` → Notification Scheduler |
| **Merge / sort** | Done by `TimelineService` | **Not** done by providers |
| **Notification policy** | N/A at provider layer | **Not** applied by providers |

### Why both exist

- **TimelineProvider** optimizes for *what the human sees* — unified agenda with presentation cues.
- **OccurrenceProvider** optimizes for *what the system must schedule* — stable identity, times, source, type, status.

Same underlying repositories and shared `internal/recurrence` expansion; different mapping and product contracts.

### Implemented Occurrence providers

| Provider | Source filter |
| --- | --- |
| `GoogleOccurrenceProvider` | Google `calendar_events` |
| `MicrosoftICSOccurrenceProvider` | Microsoft + ICS `calendar_events` |
| `DonnaEventOccurrenceProvider` | `donna_events` (+ RRULE expand) |
| `DonnaReminderOccurrenceProvider` | `donna_reminders` (+ RRULE expand) |

Each provider only: queries its reader, expands recurrence when needed, returns `[]Occurrence`. No merge, sort, or policy.

---

## What Occurrence contains

Only scheduling-relevant fields:

| Field | Role |
| --- | --- |
| `ID` | Stable row / series identity in Donna’s sense |
| `ParentID` | Series parent when this is a virtual occurrence |
| `OccurrenceID` | Concrete occurrence key (idempotency for notifications) |
| `UserID` | Owner |
| `Source` | `GOOGLE` \| `MICROSOFT_ICS` \| `DONNA` |
| `Type` | `EVENT` \| `REMINDER` |
| `Title` / `Description` | Human content for notification copy |
| `StartAt` / `EndAt` | Instant range |
| `Timezone` | Civil timezone context |
| `RecurrenceRule` | Optional RRULE for series |
| `Status` | `ACTIVE` \| `COMPLETED` \| `CANCELLED` \| `MISSED` |
| `Metadata` | Minimal non-display extras only |

## What Occurrence must not contain

- Colors, icons, badges  
- `ReadOnly` / editability flags  
- All-day *display* conventions as UI concerns  
- Frontend layout or card fields  

---

## Types

Defined in-package (not Timeline constants) so the scheduling domain owns its vocabulary:

- `OccurrenceType`
- `OccurrenceSource`
- `OccurrenceStatus`

Values currently align with timeline string constants for an easy future map, but they are **owned here**.

---

## Shared recurrence

RRULE normalize / expand / virtual IDs live in `internal/recurrence`.

- Timeline providers continue to use thin wrappers in `business` (`ExpandRecurrence`, …).
- Occurrence providers call `recurrence` directly.
- One implementation; no duplicated expansion logic.

---

## Validation

`Occurrence.Validate()` enforces required identity, owner, type/source/status enums, non-empty title, non-zero start/end, `end_at >= start_at`, and timezone.

`Occurrence.Normalize()` trims strings and clears empty optional pointers — call before validate when ingesting external data.

---

## Non-goals (post scheduler cutover)

- No TimelineService / Timeline API changes for UI  
- No Notification Dispatcher / queue / REST notification API changes  
- No Chat / Actions / frontend changes  

OccurrenceService is the scheduling feed for enqueue; Timeline remains presentation-only.
