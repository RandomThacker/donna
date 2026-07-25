# Donna Observability

**Status:** Source of truth  
**Applies to:** All Donna services and future milestones  

Every observability decision for Donna MUST follow this document. When implementation and this doc diverge, update the implementation or amend this document explicitly — never invent ad-hoc logging.

Related: [SYSTEM_DESIGN.md](./SYSTEM_DESIGN.md) §19, [ARCHITECTURE.md](./ARCHITECTURE.md), [CURSOR_RULES.md](./CURSOR_RULES.md).

---

## Philosophy

Donna is a production-grade SaaS.

Every request, scheduler execution, AI call, background worker, database query, and external integration should be observable.

Observability consists of:

| Pillar | Purpose |
| --- | --- |
| Logging | What happened, correlated by request and module |
| Metrics | How often / how fast / how expensive (future) |
| Tracing | Where time was spent across spans (future) |
| Error monitoring | Unexpected failures with stack and release (future) |
| Audit logs | Permanent record of security- and compliance-sensitive actions |
| AI usage monitoring | Tokens, latency, cost, prompt version per LLM call |

The purpose is fast debugging, production visibility, and future scalability over a 5–10 year product life.

**Invariant:** Observability is a platform concern. Feature modules consume the Logger Factory; they do not invent their own logging stacks.

---

## Logging Standard

Use Go’s standard `log/slog` package (API service). Equivalent structured logging on other services must match field names in this document.

**Never use:**

- `fmt.Println`
- `log.Println`
- `println`
- ad-hoc string concatenation as the primary log form

Only structured logging.

| Environment | Handler |
| --- | --- |
| `development` | Human-readable text (`slog.TextHandler`) |
| `staging` / `production` | JSON (`slog.JSONHandler`) |

Log level is configurable via app config (`log_level` / `LOG_LEVEL`).

---

## Logger Factory

Donna uses **module loggers**.

Packages MUST NOT instantiate `slog.Logger` (or wrappers) directly.

Create loggers only through the centralized Logger Factory:

```go
calendarLogger := logger.Module("calendar")
chatLogger := logger.Module("chat")
dashboardLogger := logger.Module("dashboard")
schedulerLogger := logger.Module("scheduler")
authLogger := logger.Module("auth")
aiLogger := logger.Module("ai")
memoryLogger := logger.Module("memory")
notificationLogger := logger.Module("notification")
httpLogger := logger.Module("http")
databaseLogger := logger.Module("database")
workerLogger := logger.Module("worker")
appLogger := logger.Module("app")
```

Each module logger automatically injects:

| Field | Value |
| --- | --- |
| `module` | Module name passed to `Module(...)` |
| `service` | Process service name (e.g. `donna-api`) |
| `environment` | `development` \| `staging` \| `production` |

Future global fields (region, version, instance id) MUST be added at the Factory root so every module inherits them automatically.

This eliminates duplicated logging fields and keeps schemas consistent.

Every future module must obtain its logger from the Logger Factory.

---

## Standard Context

Every request (and job) should propagate a logging context.

Supported fields (include **only** when present):

| Field | When |
| --- | --- |
| `request_id` | Every HTTP request (required) |
| `user_id` | Authenticated requests |
| `conversation_id` | Chat flows |
| `session_id` | Session-backed auth |
| `calendar_source_id` | Calendar provider operations |
| `connection_id` | Linked external accounts |
| `job_id` | Background / queue jobs |
| `scheduler_id` | Scheduler executions |
| `trace_id` | When OpenTelemetry is enabled (future) |

The logger MUST automatically inherit these fields from `context.Context`.

Future middleware and services MUST preserve and forward context (never drop `request_id`).

HTTP middleware:

1. Accepts inbound `X-Request-ID` or generates one
2. Sets response header `X-Request-ID`
3. Stores fields on Gin context and `request.Context()`

---

## Log Levels

| Level | Use for |
| --- | --- |
| **DEBUG** | Development detail: SQL, cache, prompt building, tool execution internals |
| **INFO** | Business and lifecycle events: auth, calendar, tasks, dashboard, scheduler, memory, successful requests |
| **WARN** | Retries, slow queries/requests, fallbacks, recoverable failures, budget breaches |
| **ERROR** | Unexpected failures, panics, database / Google / AI / Redis failures that need attention |

Do not log routine success paths at ERROR. Do not bury production incidents at DEBUG.

---

## Log Format

Every log line SHOULD include:

| Field | Required |
| --- | --- |
| `timestamp` | Yes (handler) |
| `level` | Yes |
| `module` | Yes (factory) |
| `service` | Yes (factory) |
| `msg` / message | Yes |
| `environment` | Yes (factory) |
| `request_id` | When in a request |
| `duration_ms` | When measuring work |
| Optional metadata | As structured attrs |

### Never log

- Passwords
- JWTs / session tokens
- OAuth access or refresh tokens
- API keys / client secrets
- Authorization headers
- Raw prompt contents that embed secrets
- Personally sensitive data beyond stable opaque IDs needed for ops (prefer `user_id` over email in logs)

Use redaction helpers when logging maps or headers.

---

## Request Logging

Every HTTP request MUST automatically log:

| Field | Source |
| --- | --- |
| Request ID | Middleware |
| Method | Request |
| Path | Request URL path |
| Status | Response status |
| Duration | Middleware timer (`duration_ms`) |
| Remote IP | Client IP |
| User Agent | Request header |
| Authenticated user | `user_id` when available |

Rules:

- Log completed requests at **INFO** by default
- Log at **WARN** when duration exceeds **500ms** (slow request threshold)
- Log at **ERROR** when the handler panics (recovery middleware)

---

## Business Event Logging

Use module loggers and stable event names (helpers preferred).

### Authentication

- Login
- Logout
- Refresh
- Google Account Linked

### Calendar

- Event Created / Updated / Deleted
- Calendar Sync

### Tasks

- Created / Updated / Completed / Deleted

### Scheduler

- Reminder Scheduled / Sent / Failed

### Notifications

- Delivered / Failed / Dismissed

### Chat

- Conversation Started
- Message Sent / Received

### Memory

- Stored / Retrieved / Updated / Deleted

### AI

Record via AI usage helpers (see below): model, latency, prompt version, tokens, estimated cost, tools used, memory retrieval count.

---

## AI Usage Tracking

Every LLM request SHOULD automatically record (structured attrs / helper):

| Field | Description |
| --- | --- |
| `model` | Model id |
| `provider` | e.g. openai |
| `input_tokens` | Prompt tokens |
| `output_tokens` | Completion tokens |
| `latency_ms` | End-to-end LLM latency |
| `estimated_cost_usd` | Best-effort cost estimate |
| `conversation_id` | When applicable |
| `user_id` | When applicable |
| `prompt_version` | Prompt / template version |
| `tools_used` | Tool names invoked |
| `memory_retrieval_count` | Memories fetched for context |

Future admin dashboards will aggregate these metrics. Logging is the Phase 1 capture path; metrics exporters come later.

---

## Metrics (Future)

Documented for roadmap — **not implemented** in the current foundation:

- API request count / latency histograms
- Database query latency
- Scheduler job count / duration
- Reminder job outcomes
- Memory retrieval latency / count
- Calendar sync duration / failures
- AI request count / tokens / cost
- Failure rates by module

Target stack: Prometheus + Grafana.

---

## Tracing (Future)

Donna will adopt **OpenTelemetry**.

| Span | Parent |
| --- | --- |
| HTTP request | Trace root |
| Database calls | Child spans |
| Google API calls | Child spans |
| AI calls | Child spans |
| Scheduler executions | Trace root or child of trigger |

`trace_id` will join logs and traces. Until then, `request_id` / `job_id` are the correlation keys.

---

## Error Monitoring (Future)

| Environment | Sink |
| --- | --- |
| Development | Console (structured logs) |
| Production | Sentry (planned) |

Every captured exception SHOULD include: request ID, user (if any), stack trace, environment, release version.

Do not send secrets to error trackers.

---

## Audit Logs

Permanent audit records (DB or append-only store — future table/API) for:

- Authentication events
- Account linking
- Calendar connections
- Reminder creation / deletion
- Settings changes
- Future billing events

Audit logs are distinct from application INFO logs: they are retained longer and are queryable for security/compliance. Application logs may emit a mirror event at INFO when an audit row is written.

---

## Performance Budgets

| Surface | Target | Action when exceeded |
| --- | --- | --- |
| API handler latency | &lt; 200ms | Prefer WARN (budget); request middleware WARNs at ≥ 500ms |
| Database query | &lt; 50ms | WARN via database logger helper |
| Scheduler job | &lt; 1s | WARN via scheduler helper |
| AI request | &lt; 5s | WARN via AI helper |

Warn whenever budgets are exceeded. Do not fail the request solely because a budget was missed unless a product rule says otherwise.

---

## Future Stack

| Now | Later |
| --- | --- |
| `slog` | OpenTelemetry |
| Text / JSON logs | Loki (log aggregation) |
| Module Logger Factory | Prometheus + Grafana |
| Request ID correlation | Sentry |
| AI / domain helpers | Full audit store |

---

## Implementation map (API)

| Concern | Location |
| --- | --- |
| Factory + module loggers | `services/api/internal/logger` |
| Context field propagation | `services/api/internal/logger` (+ middleware) |
| HTTP request logging | `services/api/internal/middleware` |
| Domain helpers (AI, auth, calendar, …) | `services/api/internal/logger` |
| Constants / thresholds | `services/api/internal/constant` + logger package |

Non-goals for the current foundation: Prometheus, Grafana, Loki, OpenTelemetry SDK wiring, Sentry.

---

## Compliance checklist for new modules

1. Obtain logger via `factory.Module("<name>")` only.
2. Pass `context.Context` into log calls so request/job fields attach.
3. Use helpers for AI / auth / calendar / scheduler / DB / worker events when applicable.
4. Never log secrets; redact headers and token-like fields.
5. Preserve context across goroutines and outbound calls.
6. Prefer INFO for business events, WARN for budgets/retries, ERROR for unexpected failures.
