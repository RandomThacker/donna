# Donna Architecture

## System overview

```mermaid
flowchart LR
  Web["apps/web<br/>Next.js"]
  API["services/api<br/>Go Gin"]
  AI["services/ai<br/>FastAPI"]
  DB[(PostgreSQL)]
  Google[Google OAuth and Calendar]

  Web -->|"REST /api/v1"| API
  API --> DB
  API -->|"tool requests"| AI
  AI -->|"persist via API only"| API
  API --> Google
  Web --> Google
```

## Ownership boundaries

| Layer | Owns | Does not own |
| --- | --- | --- |
| `apps/web` | UI, client state, presentation | Business rules, DB, LLM calls |
| `services/api` | Auth, CRUD, providers, scheduling, orchestration | Prompt engineering, embeddings |
| `services/ai` | Intent, planning, memory retrieval, tool decisions, replies | Direct DB writes |
| PostgreSQL | Source of truth | — |

**Rule:** AI reasons. Backend executes. Database stores.

## Request paths

### User chat

```mermaid
sequenceDiagram
  participant User
  participant Web
  participant API
  participant AI
  participant DB

  User->>Web: Send message
  Web->>API: POST /api/v1/conversations/:id/messages
  API->>DB: Persist user message
  API->>AI: Reason with context and tools
  AI-->>API: Reply plus tool calls
  API->>DB: Execute tools and persist assistant message
  API-->>Web: Typed response / stream
  Web-->>User: Phone UI update
```

### Proactive check-in

```mermaid
sequenceDiagram
  participant Scheduler
  participant API
  participant AI
  participant DB
  participant Web

  Scheduler->>API: Trigger morning briefing
  API->>DB: Load plan, tasks, events
  API->>AI: Build briefing
  AI-->>API: Message content
  API->>DB: Persist Donna message
  API-->>Web: Push / websocket / poll
```

## API design

- Versioned REST under `/api/v1`
- Handlers validate and map HTTP only
- Services own business logic
- Repositories own SQL / migrations consumers
- Typed JSON request and response contracts live in `packages/types`

## Calendar unification

```mermaid
flowchart TB
  ProviderIface[CalendarProvider interface]
  GoogleAdapter[GoogleCalendarAdapter]
  FutureOutlook[OutlookAdapter future]
  Unified[Unified Calendar Events]
  Defaults[User calendar defaults]

  ProviderIface --> GoogleAdapter
  ProviderIface --> FutureOutlook
  GoogleAdapter --> Unified
  FutureOutlook --> Unified
  Defaults --> Unified
```

Phase 1 ships Google only. All calendar code goes through a provider interface so Outlook/Apple can plug in later.

## Auth model

- Google OAuth creates a **Donna account** (identity).
- Integration connections are separate rows (`connections`).
- Login Google account is not automatically a calendar connection.
- Session tokens issued by API; web never stores provider refresh tokens in localStorage.

## Frontend feature layout

Each feature follows:

```text
Feature/
  Feature.tsx        # UI only
  Feature.logic.ts   # hooks, handlers
  Feature.styles.ts  # Tailwind class constants
  Feature.types.ts   # local types
  index.ts
```

Dashboard left column = widgets. Right column = persistent Phone chat shell.

## AI tool calling

AI may request tools such as:

- `create_task`, `complete_task`
- `create_event`, `update_event`, `delete_event`
- `save_daily_plan`, `save_memory`
- `search_memory`

API validates permissions, executes via services, returns structured results to AI for the final user-facing reply.

## Packages

| Package | Purpose |
| --- | --- |
| `packages/types` | Shared API DTOs / OpenAPI-derived types |
| `packages/ui` | Shared presentational primitives beyond shadcn |
| `packages/shared` | Pure helpers shared by web (no server secrets) |

## Infra

`infra/docker` runs Postgres (and later Redis for jobs/push fan-out). Local compose brings up DB + API + AI + web.
