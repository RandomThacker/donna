# Action Layer (Phase 2.4)

Donna’s **Action Layer** is the single entry point for domain workflows.

Every interface — REST today, Chat / Telegram / AI later — calls the same Actions.

```text
REST / Chat / Telegram / AI
            ↓
         Actions
            ↓
         Services
            ↓
       Repositories
```

## Rules

| Layer | May | Must not |
| --- | --- | --- |
| **Handler** | Bind HTTP, map to Action request, map Action result to HTTP | Own business rules, call repositories |
| **Action** | Validate input, call one or more services, coordinate workflows, publish domain events (placeholder), return domain DTOs | Import Gin/HTTP/Telegram/AI |
| **Service** | Entity rules, persistence orchestration | Know about transport |
| **Repository** | SQL | Business workflows |

## Package

`services/api/internal/actions/`

Each Action is an independent type with:

- Explicit constructor DI (`NewCreateEventAction(EventService, DomainEventPublisher)`)
- `Execute(ctx, Request) (Result, error)`
- Action DTO → HTTP `model` mapping lives in handlers (avoids `model` ↔ `actions` import cycles via `business`/`validation`).


### Phase 2.4 Actions

| Action | Purpose |
| --- | --- |
| `CreateEventAction` / `UpdateEventAction` / `DeleteEventAction` | Donna events |
| `CreateReminderAction` / `UpdateReminderAction` / `DeleteReminderAction` | Donna reminders |
| `CreateTaskAction` / `UpdateTaskAction` / `CompleteTaskAction` / `DeleteTaskAction` | Tasks |
| `QueryTimelineAction` | Unified timeline window |
| `GetNotificationsAction` / `MarkNotificationReadAction` / `DismissNotificationAction` | Notification inbox |

`actions.NewRegistry(Deps)` wires constructors for `app.go`.

## Domain events

Actions accept a `DomainEventPublisher`. Phase 2.4 uses `NoopPublisher`. Future bus implementations plug in without changing Action call sites.

## Reuse example

```go
// REST
result, err := createEvent.Execute(ctx, actions.CreateEventRequest{UserID: uid, Title: "...", ...})

// Future Chat / AI — same Action, no handler changes
result, err := createEvent.Execute(ctx, actions.CreateEventRequest{UserID: uid, Title: parsedTitle, ...})
```

## Testing

Unit tests live in `internal/actions` with stub service ports. They verify validation and orchestration (which services were called, what DTOs returned). No HTTP required.

## Out of scope

Chat, Telegram, AI adapters, real domain-event bus.
