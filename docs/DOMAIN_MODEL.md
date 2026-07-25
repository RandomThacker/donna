# Donna Domain Model

**Status:** Source of truth for business entities and relationships  
**Milestone:** M2 — Domain Modeling (documentation only)  
**Non-goals:** SQL, table schemas, migrations, repositories, DTOs, handlers  

When implementation diverges from this document, update the document or the code deliberately. Do not invent parallel domain vocabularies.

Related: [PRD.md](./PRD.md), [SYSTEM_DESIGN.md](./SYSTEM_DESIGN.md), [ARCHITECTURE.md](./ARCHITECTURE.md), [DATABASE.md](./DATABASE.md), [OBSERVABILITY.md](./OBSERVABILITY.md).

---

## Introduction

### Purpose of domain modeling

Domain modeling defines the **business language** Donna uses internally: who owns what, how concepts relate, and where boundaries sit.

It answers:

- What is a first-class Donna concept?
- What is only an integration concern?
- Which objects form consistency boundaries (aggregates)?
- How can the product grow for 5–10 years without rewriting the core?

This document is the blueprint for a future database schema. It intentionally contains **no SQL**. Field-level logical design: [DATA_MODEL.md](./DATA_MODEL.md). Persistence conventions: [DATABASE.md](./DATABASE.md).

### Why Donna models its own domain (not providers)

Donna is a **Personal AI Operating System**, not a thin wrapper around Google Calendar or OpenAI.

| Donna owns | Providers own |
| --- | --- |
| Events, tasks, reminders, conversations, memory, plans | Their proprietary event/message/API shapes |
| User identity and preferences | OAuth tokens and provider-side calendars |
| Unified semantics for AI and UI | Sync transport and remote IDs |

**Principle:** External systems synchronize with Donna’s objects. They do not define them.

Example:

```text
Donna Event  (canonical)
 ├── Google Calendar Event  (projection / link)
 ├── Outlook Event          (future)
 └── Apple Calendar Event   (future)
```

If Google’s API changes, adapters change. Donna’s Event model stays stable for the dashboard, phone chat, scheduler, and AI tools.

---

## Core Domains

### Identity

#### Purpose

Establish who the human is inside Donna: the account that owns all personal data and sessions.

#### Responsibilities

**Owns**

- Donna user account
- Profile (display name, avatar metadata, locale/timezone preferences at identity level)
- Authentication sessions / credential links used for *login* (distinct from integration connections)

**Never owns**

- Calendar sync state or provider calendar lists
- Chat message bodies
- Integration OAuth refresh tokens as “the user” (tokens belong under Connected Accounts)
- AI prompt content or embeddings

#### Relationships

- **Root owner** of nearly all other domains: Connected Accounts, Calendar, Tasks, Conversations, Memory, Notifications, Settings, Audit
- Login via Google OAuth creates Identity only; it does **not** automatically create a calendar connection ([ARCHITECTURE.md](./ARCHITECTURE.md))

#### Future integrations

- Additional login providers (Apple, Microsoft) attach as identity providers without changing User as the aggregate root
- Teams / shared workspaces (optional later): Identity becomes a member of a Workspace; personal data remains user-scoped until explicitly shared

---

### Connected Accounts

#### Purpose

Represent third-party accounts the user has **authorized Donna to use as integrations** (calendar, future Slack, Notion, etc.).

#### Responsibilities

**Owns**

- Connection records: provider type, account label, status, scopes, encrypted credentials reference
- Lifecycle: connect, disconnect, reauth, revoke

**Never owns**

- Canonical calendar events or tasks (those are Donna domain objects that *may* link to a connection)
- Login session cookies/JWTs (Identity)
- Provider-specific UI logic

#### Relationships

- Belongs to **User** (Identity)
- Parent of **Calendar Sources** (and future mailbox/drive sources if introduced)
- Referenced by sync jobs (Scheduler) and Audit events

#### Future integrations

```text
User
 └── Connected Account (provider = google | microsoft | apple | …)
        ├── Calendar Sources
        ├── (future) Mail Sources — out of Phase 1 scope
        └── (future) Other capability sources
```

New providers add a `provider` value and an adapter. No new “GoogleUser” table in the core domain.

---

### Calendar

#### Purpose

Maintain Donna’s **unified calendar**: sources the user cares about and events Donna (and the AI) reason about.

#### Responsibilities

**Owns**

- **Calendar Source** — a calendar feed/container under a Connected Account (e.g. “Personal”, “Work”) plus Donna-side defaults (personal / work / reminder routing)
- **Calendar Event** — canonical event: title, time range, location, attendees summary, status, visibility, recurrence summary as Donna understands it
- Sync metadata *attached to Donna objects* (last synced at, provider external id, etag/version) without making the provider the source of truth for meaning

**Never owns**

- Raw Google/Outlook API payloads as the primary model (store opaque provider payload only as integration side-data if needed)
- Reminder delivery channels (Notifications)
- Task backlog items (Tasks) — an event may *link* to a task, but tasks are separate

#### Relationships

- Calendar Source → Connected Account → User
- Calendar Event → Calendar Source (and thus User)
- May reference Tasks or Conversations optionally (soft links / associations)
- Scheduler jobs may create or update events; AI requests tools that the API executes against this domain

#### Future integrations

```text
Donna Event
 ├── link → Google event id (via Google adapter)
 ├── link → Outlook event id
 └── link → Apple event id
```

Adapters implement get/create/update/delete against providers. Unified queries always read Donna Events. Multi-account Google (personal + work) = multiple Connected Accounts / Sources, one Event model.

---

### Tasks

#### Purpose

Capture actionable work the user (or Donna) tracks: quick todos, backlog, priorities, due dates, completion.

#### Responsibilities

**Owns**

- Task entities: title, status, priority, due date, recurrence intent (data ready even if UI is basic), backlog vs active
- Optional linkage to goals / daily plans (rhythm concepts may sit adjacent; see Aggregates)

**Never owns**

- Calendar event time-blocks as a substitute for tasks (use Calendar Event + optional link)
- Push delivery (Notifications)
- Long-form notes as the task itself (Notes may link to tasks; notes are related work content)

#### Relationships

- Belongs to User
- May own or associate **Reminders**
- May be created/updated via AI tools (`create_task`, `complete_task`) executed by the API
- May appear in Daily Plans / accountability flows without merging domains

#### Future integrations

- Provider task lists (Google Tasks, Microsoft To Do) sync **into** Donna Tasks via adapters — Donna Task remains canonical
- No Phase 1 requirement to sync outbound to Google Tasks

---

### Reminders

#### Purpose

Represent “nudge me about X at time T” as a first-class commitment, whether tied to a task, event, or free-standing note from chat.

#### Responsibilities

**Owns**

- Reminder definition: when, what, status (scheduled / sent / cancelled / failed)
- Association to Task and/or Calendar Event and/or Conversation Message when relevant

**Never owns**

- Actual push/email/Telegram transport (Notifications + channel adapters)
- Cron infrastructure details (Scheduler owns job execution records)

#### Relationships

- Usually under Task Aggregate; may also reference Calendar Event
- Scheduler materializes delivery attempts; Notifications record user-visible alerts
- Audit records creation/deletion of reminders

#### Future integrations

- Channel is pluggable (browser push now; Telegram/WhatsApp later) without changing Reminder meaning
- Provider-native reminders (e.g. Google Popups) are sync projections, not Donna’s reminder authority

---

### Conversations

#### Purpose

Hold the phone-chat thread(s) between the user and Donna — the primary interpersonal surface.

#### Responsibilities

**Owns**

- Conversation: participants (user + Donna), state (active/archived), unread indicators, metadata for UI
- Ordering and ownership of Messages within the conversation

**Never owns**

- LLM token accounting as the conversation itself (AI usage is observability/AI domain)
- Memory extraction results (Memory domain stores durable facts; conversation may *trigger* extraction)

#### Relationships

- Belongs to User
- Contains Messages
- May reference Daily Plan / check-in type as conversation purpose (morning, midday, evening)
- AI service is called with conversation context; persistence stays in API

#### Future integrations

- Additional surfaces (Telegram bot, WhatsApp) become **channels** attached to the same Conversation/Message model, or linked conversations — core chat semantics unchanged
- Phase 1: web phone UI only

---

### Messages

#### Purpose

Individual turns in a conversation: user text, Donna replies, system notices, typing-related state as needed for UX.

#### Responsibilities

**Owns**

- Message body (user-visible), role (user / assistant / system), timestamps, delivery/read state for UI
- Optional structured attachments metadata (e.g. “created task X”) without embedding full domain graphs

**Never owns**

- Tool execution side effects (API services mutate Tasks/Events; message may cite the result)
- Raw full LLM traces with secrets (observability / AI usage logs)

#### Relationships

- Belongs to Conversation → User
- May link to created Task / Event / Memory ids as citations
- Feeds Memory extractor (AI proposes → API stores Memory)

#### Future integrations

- Streaming is a transport concern; stored Message is the durable unit
- Multimodal content later (voice transcript) extends message parts without replacing Message

---

### Memory

#### Purpose

Durable knowledge Donna should remember across days: projects, preferences, people, commitments, ideas — with semantic retrieval later.

#### Responsibilities

**Owns**

- Memory records: content, category/type, importance, source (chat, explicit save, review)
- Lifecycle: create, update, soft-delete / archive
- Hooks for embeddings (vector side-car or column later) without defining SQL here

**Never owns**

- Ephemeral chat context windows
- Calendar event instances as “memory” (events stay in Calendar; memory may summarize preferences about them)

#### Relationships

- Belongs to User
- May cite Conversation / Message as provenance
- Retrieved by AI via API tools (`search_memory`); AI never writes Memory directly

#### Future integrations

- Embedding providers (OpenAI, local models) are swappable behind an embedding port
- Memory remains Donna-owned text/structured facts + vectors

---

### Scheduler

#### Purpose

Run time-based and queued work that makes Donna proactive: morning briefing, midday check-in, evening reflection, reminder firing, calendar sync ticks.

#### Responsibilities

**Owns**

- Scheduler Job definitions and run records: type, schedule/trigger, payload reference, status, attempts
- Coordination of *when* work runs — not the full business mutation logic (that stays in domain services)

**Never owns**

- Notification channel SDKs
- Canonical Task/Event content (jobs reference ids and call domain services)

#### Relationships

- Jobs are scoped to User (and sometimes Connection / Reminder / Conversation)
- Invokes Business layer use cases; emits Notifications; writes Audit where required
- Observable via job_id / scheduler_id ([OBSERVABILITY.md](./OBSERVABILITY.md))

#### Future integrations

- Backing store may move from DB-poll → Redis/queue without changing Job as a domain concept
- New job types (billing renewals, workspace digests) add job kinds, not new product cores

---

### Notifications

#### Purpose

User-visible alerts and delivery attempts across channels.

#### Responsibilities

**Owns**

- Notification records: title/body or template key, priority, read/dismissed state, channel, status
- Correlation to Reminder / Event / Scheduler Job / Conversation

**Never owns**

- Reminder *policy* (when to remind) — that is Reminders
- Browser push subscription crypto details beyond what’s needed (infra concern with a thin domain reference)

#### Relationships

- Belongs to User
- Produced by Scheduler, Calendar (meeting soon), Chat (Donna ping), system
- Channel adapters: browser push (Phase 1), later email/Telegram/WhatsApp

#### Future integrations

```text
Notification (canonical)
 ├── Browser Push delivery
 ├── Email delivery (future)
 └── Telegram delivery (future)
```

---

### Settings

#### Purpose

User-configurable behavior: defaults, preferences, feature toggles that are not secrets.

#### Responsibilities

**Owns**

- Preferences: default personal/work/reminder calendars, quiet hours, briefing times, locale/timezone overrides if not on profile
- Notification preferences

**Never owns**

- OAuth tokens
- Feature-flag infrastructure for the whole fleet (may use settings for user opts only)

#### Relationships

- 1:1 (or 1:few) with User under Identity Aggregate
- Read by Calendar routing, Scheduler, Notifications, AI personalization (via API-provided context)

#### Future integrations

- Workspace-level settings later nest under Workspace without deleting user settings
- Billing plan limits referenced here as read-only prefs, not payment ledgers

---

### Audit

#### Purpose

Permanent, queryable record of security- and compliance-sensitive actions.

#### Responsibilities

**Owns**

- Audit Log entries: actor, action, subject type/id, metadata (non-secret), timestamp
- Coverage per [OBSERVABILITY.md](./OBSERVABILITY.md): auth, account linking, calendar connections, reminder create/delete, settings changes, future billing

**Never owns**

- High-volume debug logs (those are observability logs)
- Full request payloads with tokens

#### Relationships

- References User (actor) and optional subject entities
- Written by API Business layer on sensitive mutations — not by AI

#### Future integrations

- Export to SIEM later; schema of Audit Log stays Donna-owned
- Workspace admins query audit without changing entry shape

---

### AI

#### Purpose

Represent AI as a **bounded integration and usage concern**, not as the owner of business data.

Donna’s AI domain in the product model covers:

- Usage/accounting records needed for product analytics and cost control (aligned with observability AI usage)
- Prompt/version references used when generating replies (identifiers, not prompt secret stores in user tables)
- Tool invocation intents as ephemeral orchestration — durable effects land in Tasks/Calendar/Memory/Messages

#### Responsibilities

**Owns**

- Conceptual **AI Usage** records (model, provider, tokens, latency, cost estimate, conversation/user refs) — may start as structured logs and later harden into storeable entities
- Provider port: which model backend fulfilled a request (OpenAI, Anthropic, local)

**Never owns**

- Authoritative Tasks, Events, Memories, Messages (API persists those)
- Direct database writes from the AI service ([ARCHITECTURE.md](./ARCHITECTURE.md))

#### Relationships

- Called by API during Conversations and Scheduler briefings
- Reads Memory / Calendar / Tasks **through API-assembled context**, not by owning those tables
- Tool requests return to API for execution

#### Future integrations

```text
AI Service (reasoning)
 ├── OpenAI adapter
 ├── Anthropic adapter
 └── Local model adapter
```

Swapping LLM providers does not change Conversation, Message, Task, or Memory shapes.

---

## Aggregates

Aggregates define **consistency and ownership boundaries**. Transactions should not span unrelated aggregates without an explicit application workflow (saga/process manager later if needed).

### Identity Aggregate

**Members:** User, Profile (identity facets), Connected Account, Settings  

**Responsibility:** Authenticate the human, hold integration credentials at the connection boundary, and store preferences. Connected Accounts are included here because their lifecycle is account-security sensitive and user-rooted; Calendar Sources hang off connections but event consistency lives in Calendar.

**Note:** Large sync operations should not lock the whole Identity Aggregate — treat Connection as a boundary that Calendar sync addresses by id.

### Calendar Aggregate

**Members:** Calendar Source, Calendar Event (and provider link side-data)  

**Responsibility:** Keep unified calendar state consistent per source/event. Sync adapters update this aggregate; UI and AI read Events as canonical.

### Task Aggregate

**Members:** Task, Reminder (when reminder is primarily task-scoped)  

**Responsibility:** Actionable work and its nudges. Reminders that are only event-scoped may be modeled as part of Calendar or as shared associations — prefer a single Reminder entity with optional foreign references rather than duplicate types.

### Conversation Aggregate

**Members:** Conversation, Message  

**Responsibility:** Chat history integrity and ordering. Streaming is outside the aggregate; commits append Messages.

### Memory Aggregate

**Members:** Memory  

**Responsibility:** Durable recalled knowledge per user. Embedding updates are part of the same conceptual aggregate even if stored in a side structure later.

### Notification Aggregate

**Members:** Notification  

**Responsibility:** User-visible alert state and delivery attempts. Does not own Reminder policy.

### Scheduler Aggregate

**Members:** Scheduler Job (definition + run/execution records as needed)  

**Responsibility:** Time-based orchestration records. Invokes other aggregates’ application services; does not embed their data.

### Audit Aggregate

**Members:** Audit Log  

**Responsibility:** Append-only sensitive action history. Writes are independent transactions after (or within) the business operation per [DATABASE.md](./DATABASE.md).

### AI Aggregate (logical)

**Members:** AI Usage (and future prompt-template references)  

**Responsibility:** Track reasoning cost/usage. Not a parent of Messages; correlates by ids.

### Rhythm / Planning (adjacent — Phase 1 product, light aggregate)

PRD includes Daily Plans, Goals, Daily Reviews, Check-ins, Notes. Treat as **Planning Aggregate** (Goals, Daily Plan, Daily Review, Check-in) and **Notes** under Work or Planning as links evolve — without bloating Calendar or Chat.

| Concept | Aggregate home |
| --- | --- |
| Goal | Planning |
| Daily Plan / Review / Check-in | Planning |
| Note | Notes (user-owned; link to Conversation/Task) |

These support accountability without changing the provider-unification story.

---

## Entity relationship (ownership)

Textual ER — ownership and containment. Not a table list.

```text
User
│
├── Profile / Settings
│
├── Connected Accounts
│      └── Calendar Sources
│              └── Calendar Events
│                     └── (optional link) Reminders
│
├── Tasks
│      └── Reminders
│
├── Conversations
│      └── Messages
│             └── (optional citations → Task / Event / Memory)
│
├── Memories
│
├── Notifications
│
├── Scheduler Jobs
│
├── AI Usage (correlation)
│
├── Planning (Goals, Daily Plans, Reviews, Check-ins)
│
├── Notes
│
└── Audit Logs
```

### Relationship rules

1. **User is the tenancy root** for Phase 1 (single-user personal OS).
2. **Connected Account** never owns canonical Events; it owns the right to sync Sources.
3. **Calendar Event** is Donna’s event; provider ids are links.
4. **Message** never bypasses API to mutate Tasks/Events; citations are after-the-fact.
5. **AI** never appears as owner of User data in this tree.

### Integration overlay (not ownership)

```text
Calendar Event ──provider_link──► Google / Outlook / Apple remote id
Task            ──provider_link──► (future) Google Tasks / To Do
Notification    ──delivery──────► Push / Email / Telegram adapters
Message         ──channel───────► Web / (future) Telegram thread id
```

---

## Architecture alignment review

| Concern | Alignment |
| --- | --- |
| Clean Architecture | Domains map to Business/Entity packages later; providers are interface adapters; handlers stay thin |
| Module boundaries | Identity, Calendar, Tasks, Chat, Memory, Scheduler, Notifications, AI match logger module names and future Go packages |
| Future AI | AI Aggregate + tool ports; no AI DB writes; Memories/Messages stay API-owned |
| Future calendars | Unified Event + Source + Connection; adapters per provider |
| Future notifications | Notification Aggregate + channel adapters |
| Future scheduler | Job aggregate invokes use cases; Redis/queue is infra |
| Future billing | New Billing Aggregate under User/Workspace; Audit already lists billing events; do not entangle Calendar |
| Future teams / workspaces | Introduce Workspace as optional parent above shared resources; keep User personal data paths stable; avoid premature shared-calendar schema |

### Evolution notes (do not build yet)

- **Workspace:** `Workspace → Members → shared Calendars/Tasks` while personal User tree remains.
- **Billing:** `Customer → Subscription → Invoice`; gate features via Settings/entitlements read models.
- **Outbox / saga:** Cross-aggregate workflows (e.g. “create event + reminder + notification”) may gain an outbox later; not required for domain vocabulary now.
- **CQRS read models:** Dashboard widgets may later use projections; write model stays as above.

### Explicit non-goals for this milestone

- No SQL, migrations, indexes, or repository code
- No Google/OpenAI table shapes as primary keys of the business
- No Telegram/WhatsApp/Gmail/Drive/GitHub domain objects in Phase 1 core (channels may appear later as integration overlays)

---

## Summary

Donna’s database will store **Donna’s world**: people, preferences, unified time, work, conversation, memory, and the jobs/alerts that make her proactive.

Google, Outlook, Apple, OpenAI, and chat transports **sync and serve** that world. They do not define it.
