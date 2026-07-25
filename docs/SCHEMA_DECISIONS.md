# Donna Schema Decisions & Review

**Status:** Formal architecture / schema review (pre-implementation)  
**Milestone:** M2.3 — Schema Review (documentation only)  
**Inputs:** [DOMAIN_MODEL.md](./DOMAIN_MODEL.md), [DATA_MODEL.md](./DATA_MODEL.md), [DATABASE.md](./DATABASE.md), [OBSERVABILITY.md](./OBSERVABILITY.md), [PRD.md](./PRD.md), [SYSTEM_DESIGN.md](./SYSTEM_DESIGN.md)  
**Non-goals:** SQL, migrations, Go code  

This document **does not rubber-stamp** the logical model. It records decisions, challenges weak spots, and states what must be resolved (or consciously accepted) before migrations.

**Review verdict (summary):** The design is **strong enough to proceed to SQL** after addressing a short list of **must-fix / must-decide** items in the Migration Readiness Assessment. Core philosophy (Donna-owned objects, provider links) holds.

---

# Architecture Decision Records

## SD-001

### Decision

Use **UUIDv7** as the internal primary key for all durable entities.

### Context

Keys must be globally unique, safe to generate in the API, and friendly to time-ordered inserts/indexes without exposing sequential integers.

### Alternatives Considered

- UUIDv4 (random)
- ULID / KSUID
- Bigserial / identity integers
- Database-only `gen_random_uuid()` (v4 in pgcrypto)

### Trade-offs

| Pros | Cons |
| --- | --- |
| Time-sortable; better B-tree locality than v4 | Slightly more complex generation than v4 |
| No central sequence bottleneck | Must standardize generation (app vs DB) once |
| Opaque; no enumeration attacks like serials | Larger than int8 |

### Consequences

All FKs reference UUID. Public APIs prefer public ids; internal joins use UUID.

### Future Impact

Merges, multi-region writes, and offline-friendly id allocation remain possible without key redesign.

---

## SD-002

### Decision

Expose **prefixed public ids** (`usr_`, `evt_`, …) alongside internal UUIDs.

### Context

Support, URLs, AI tool payloads, and logs need human-discriminable stable identifiers without leaking internal structure.

### Alternatives Considered

- UUID-only in APIs
- NanoID without prefixes
- Slugs from email/title

### Trade-offs

| Pros | Cons |
| --- | --- |
| Type visible in ids; fewer support mistakes | Extra unique column + generation logic |
| Stable if display names change | Must never reuse; collision handling required |
| Safer than email in URLs | Two identifiers to keep in sync |

### Consequences

Every major entity carries `id` + `public_id`. Repositories resolve public → internal at the edge.

### Future Impact

Workspace-scoped resources can keep the same prefix scheme; billing/support tools stay readable for a decade.

---

## SD-003

### Decision

**Donna owns canonical business entities**; providers only attach via link fields and adapters.

### Context

PRD requires unified calendar and multi-provider future without rewriting the product around Google.

### Alternatives Considered

- Store Google event JSON as source of truth
- Separate `google_events` / `outlook_events` tables as primary
- Soft-delete Donna row when provider disconnects immediately

### Trade-offs

| Pros | Cons |
| --- | --- |
| Product works offline from a provider | Sync conflict resolution required |
| Schema stable when APIs churn | Dual-write / upsert complexity |
| AI/UI share one vocabulary | Occasional drift until sync catches up |

### Consequences

`Calendar Event`, `Task`, `Message`, `Memory` remain meaningful if Google vanishes. Provider ids are optional links.

### Future Impact

Outlook/Apple/Google Tasks plug in as adapters + `provider_*` columns, not parallel cores.

---

## SD-004

### Decision

Model integrations as **Connected Account → capability children** (Calendar Source now; Integration Binding later), never as login User.

### Context

Login Google ≠ calendar Google ([ARCHITECTURE.md](./ARCHITECTURE.md)).

### Alternatives Considered

- Auto-create Connected Account on OAuth login
- Embed refresh tokens on User
- One “Google” blob for login + calendar

### Trade-offs

| Pros | Cons |
| --- | --- |
| Correct security/product boundary | Extra connect step for users |
| Multiple Google accounts (work/personal) | More entities to teach the team |
| Revoke integration without logging user out | |

### Consequences

Auth Identity (login) should be a **separate future/child entity** from Connected Account — see review finding **R-01**.

### Future Impact

Slack/Notion attach as accounts/bindings without touching User auth rows.

---

## SD-005

### Decision

Enforce **aggregate boundaries** for writes; cross-aggregate links are references only (nullable FKs / citations), not ownership.

### Context

Clean Architecture + long-lived SaaS: prevent “god transactions” and tangled cascades.

### Alternatives Considered

- Single User mega-aggregate for everything
- Event-sourcing all domains day one
- Shared mutable rows across services

### Trade-offs

| Pros | Cons |
| --- | --- |
| Clear module ownership in Go | Cross-aggregate workflows need orchestration |
| Safer deletes/cascades | Occasional denormalized `user_id` |
| Testable units | Reminder dual-association needs rules (R-02) |

### Consequences

Calendar sync must not lock Identity. Chat must not own Task rows.

### Future Impact

Outbox/sagas can be added later without undoing tables.

---

## SD-006

### Decision

Allow **JSON only** for opaque provider payloads, sparse prefs, citation bags, and job/notification channel payloads — never for core meaning (title, times, status, ownership).

### Context

Avoid schemaless entropy while still absorbing provider quirks.

### Alternatives Considered

- All-column schemas for every provider field
- EAV tables
- JSON for entire event bodies

### Trade-offs

| Pros | Cons |
| --- | --- |
| Extensible without constant migrations | Weaker DB constraints on JSON |
| Keeps core queryable | Risk of “just put it in JSON” creep |

### Consequences

JSON Review (below) rejects or constrains each proposed JSON field. Attendee CRM must not grow inside `attendees_summary` unchecked.

### Future Impact

Promote hot JSON keys to columns when query patterns appear.

---

## SD-007

### Decision

Prefer **soft delete** for user-facing durable data; **hard delete / retain-then-purge** for ephemeral jobs; **never delete** audit in product flows.

### Context

Recovery, sync tombstones, and compliance conflict with naive DELETE.

### Alternatives Considered

- Hard delete everywhere
- Status-only without `deleted_at`
- Archive tables from day one

### Trade-offs

| Pros | Cons |
| --- | --- |
| Recoverable UX; sync-friendly | Partial unique indexes required |
| Audit continuity | Larger tables; must filter `deleted_at` |
| | GDPR hard-delete still needed as a job |

### Consequences

Repositories default to live rows. User closure is soft + anonymization pipeline (to be specified at schema impl).

### Future Impact

Retention policies can purge soft-deleted rows without redesigning PKs.

---

## SD-008

### Decision

**Audit Log** is append-only, polymorphic by `subject_type` / `subject_id` (exception to anti-polymorphism rule).

### Context

One table must reference many aggregates for compliance ([OBSERVABILITY.md](./OBSERVABILITY.md)).

### Alternatives Considered

- Per-aggregate audit tables
- Application logs only
- Event store as sole audit

### Trade-offs

| Pros | Cons |
| --- | --- |
| Single query surface for security | Weaker referential integrity |
| Survives subject soft-delete via public_id snapshot | App must validate subject_type |

### Consequences

Core product FKs stay explicit; only Audit (and Message citations) use flexible subject addressing.

### Future Impact

SIEM export and workspace admin views reuse one shape.

---

## SD-009

### Decision

**Scheduler Job** owns *when/whether work runs*, not domain payloads’ business meaning; jobs reference Reminder/Account via FKs + small payload.

### Context

Proactive Donna (briefings, reminders, sync) needs durable scheduling without embedding Calendar/Task logic in the job row.

### Alternatives Considered

- Cron only in process memory
- Domain tables self-schedule (`next_remind_at` only)
- Full workflow engine day one

### Trade-offs

| Pros | Cons |
| --- | --- |
| Observable, retryable, multi-type | Dual sources of truth if Reminder.remind_at and Job.run_at drift (R-03) |
| Queue migration later | Polling index must be solid |

### Consequences

Business services execute; Job records status. Reminder remains policy; Notification remains delivery.

### Future Impact

Redis/SQS can execute while Job remains the ledger.

---

## SD-010

### Decision

**Notification** is the user-visible delivery aggregate; channels are adapters (`browser_push` now, others later).

### Context

PRD Phase 1 browser push; future Telegram must not fork the model.

### Alternatives Considered

- Reminder = notification
- Provider push only (FCM rows as source of truth)
- Per-channel tables as primary

### Trade-offs

| Pros | Cons |
| --- | --- |
| One inbox model | Status enum currently mixes delivery + read (R-04) |
| Channel-agnostic product language | |

### Consequences

Reminder create ≠ notification send. Scheduler/reminder_fire creates Notification attempts.

### Future Impact

Email/Telegram add enum values + adapters, not new cores.

---

## SD-011

### Decision

**Conversation / Message** owned by User; AI never persists them. Citations are soft references, not ownership.

### Context

Phone UI is the product heart; AI is a reasoner ([ARCHITECTURE.md](./ARCHITECTURE.md)).

### Alternatives Considered

- Store threads only in the AI vendor
- Message owns created Task rows
- Single blob transcript per day

### Trade-offs

| Pros | Cons |
| --- | --- |
| Donna works if OpenAI dies | Citation JSON weaker than junction (R-05) |
| Tool side-effects stay in Task/Calendar | |

### Consequences

API orchestrates persist → AI → tools → persist. Channel field prepares Telegram without changing Message meaning.

### Future Impact

Multi-surface chat reuses Conversation; AI Session hardens usage later.

---

## SD-012

### Decision

**Memory** is user-owned durable knowledge; embeddings are optional side-data; AI proposes, API writes.

### Context

Semantic memory is a PRD pillar; pgvector comes later.

### Alternatives Considered

- Vector DB as only store
- Prompt-stuffing without persistence
- Provider memory APIs as source of truth

### Trade-offs

| Pros | Cons |
| --- | --- |
| Portable facts | Embedding column/provider coupling must stay thin |
| Provenance to messages | Extraction quality is product risk, not schema |

### Consequences

`content` is canonical; `embedding` nullable. Delete is soft.

### Future Impact

Swap embedding models via `embedding_model`; add graph edges later.

---

## SD-013

### Decision

Defer physical **Auth Identity**, **Planning** (goals/plans/reviews), and **Notes** tables from the first migration wave if needed — but they are **not optional in the product**; track as known gaps (R-06).

### Context

DATA_MODEL lists Phase 1 chat/calendar/tasks thoroughly; PRD also requires daily planning and notes. Shipping auth without Auth Identity is a leak risk.

### Alternatives Considered

- Stuff IdP subject into User.email only
- Encode plans as Memories
- Block all SQL until Planning entities exist

### Trade-offs

| Pros of deferring Planning/Notes | Cons |
| --- | --- |
| Smaller first schema | Accountability features blocked |
| | Risk of misusing Memory/Task for plans |

### Consequences

**Must-add before or with Auth implementation:** Auth Identity (or equivalent). **Must-add before M7 accountability:** Planning entities. Documented so SQL generation does not “forget” PRD.

### Future Impact

Avoids emergency redesign when morning planning ships.

---

# Schema Review Report

Evaluation of each entity in [DATA_MODEL.md](./DATA_MODEL.md).

### User

| Criterion | Assessment |
| --- | --- |
| Ownership clear? | **Yes** — tenancy root |
| Aggregate boundary correct? | **Mostly** — Identity aggregate root |
| Lifecycle defined? | **Partial** — `active/disabled/pending_deletion` + soft delete; transitions underspecified |
| Delete clear? | **Yes** — soft; hard via retention |
| Indexing sufficient? | **Yes** for auth lookup; add `(status, pending_deletion)` if cleanup jobs need it |
| Future expansion? | **Yes** — workspaces/billing via links |
| Responsibilities clear? | **Yes** |
| Provider leak? | **Low** — but login IdP subject not modeled (R-01) |

**Concern:** Missing Auth Identity entity for `provider + subject` login links. Email-only uniqueness is insufficient for multi-IdP.

### Connected Account

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** — User |
| Aggregate | **Acceptable** inside Identity for credential lifecycle; sync should not lock User |
| Lifecycle | **Good** — active/needs_reauth/revoked/disconnected |
| Delete | **Soft** — good |
| Indexes | **Good** — provider upsert key present |
| Expansion | **Excellent** |
| Provider leak? | **No** — `provider` is a discriminator, not Google-specific columns |

**Concern:** `scopes` type ambiguous (string vs list) — normalize before SQL.

### Calendar Source

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** — Account + denormalized User |
| Aggregate | **Correct** — Calendar aggregate with Events |
| Lifecycle | **Implicit** via sync_enabled + soft delete — OK |
| Delete | **Soft** — OK |
| Indexes | **Good** |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** None material. Ensure denormalized `user_id` always matches Account.user_id (app invariant).

### Calendar Event

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** |
| Lifecycle | **status** + soft delete — map cancelled vs deleted carefully |
| Delete | **Soft** — OK for sync |
| Indexes | **Good** for dashboard ranges |
| Expansion | **Yes** |
| Provider leak? | **Mild** — redundant `provider` if always derivable from Source (R-07) |

**Concern:** `attendees_summary` JSON can become a dumping ground — constrain size/shape (JSON Review).

### Task

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** (with Reminders) |
| Lifecycle | **Thin** — no `in_progress` / archived; see Lifecycle Review |
| Delete | **Soft** — OK |
| Indexes | **Good** |
| Expansion | **Yes** — provider task ids ready |
| Provider leak? | **No** |

**Concern:** Align status enum with product lifecycle vocabulary before UI/AI tools freeze it (R-08).

### Reminder

| Criterion | Assessment |
| --- | --- |
| Ownership | **User clear**; Task vs Event association **ambiguous for aggregate writes** (R-02) |
| Aggregate | **Split brain** — Task aggregate vs Calendar reference |
| Lifecycle | **Good** scheduled→sent/failed/cancelled |
| Delete | **Soft** — OK |
| Indexes | **Good** for due scanning |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** Define write rules: which service may mutate Reminder; require XOR or allow both FKs with documented semantics.

### Conversation

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** |
| Lifecycle | **active/archived** — good |
| Delete | **Soft** — OK |
| Indexes | **Good** |
| Expansion | **channel** field — good |
| Provider leak? | **No** |

**Concern:** Optional rule for single “primary” phone thread — decide product-wise, not blocking.

### Message

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** |
| Lifecycle | **Weak** — no delivered/read for user messages beyond conversation unread_count |
| Delete | **Soft** — OK |
| Indexes | **Good** |
| Expansion | **parts later** — OK |
| Provider leak? | **No** |

**Concern:** `citations` JSON (R-05). Prefer promoting to junction if citation queries become first-class.

### Memory

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** |
| Lifecycle | **Implicit** via soft delete — OK |
| Delete | **Soft** — OK |
| Indexes | **OK**; vector index later |
| Expansion | **Yes** |
| Provider leak? | **No** — embedding_model is metadata |

**Concern:** None blocking.

### Notification

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** |
| Aggregate | **Correct** |
| Lifecycle | **Mixed** delivery + read in one enum (R-04) |
| Delete | **Soft** with optional hard purge — OK |
| Indexes | **Good** |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** Split delivery_status vs read/dismissed fields or document ordinal state machine strictly.

### Scheduler Job

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** (user-scoped) |
| Aggregate | **Correct** |
| Lifecycle | **Good** |
| Delete | **Hard/purge** — OK for ephemeral |
| Indexes | **Good** poller pattern |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** Keep `run_at` authoritative vs Reminder.remind_at (R-03). Prefer Job created from Reminder with single source of schedule truth.

### Settings

| Criterion | Assessment |
| --- | --- |
| Ownership | **Yes** 1:1 |
| Aggregate | **Identity** — OK |
| Lifecycle | **N/A** |
| Delete | **With user** — OK |
| Indexes | **Sufficient** |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** `public_id` on Settings is low value (R-09) — acceptable but optional to drop to reduce noise.

### Audit Log

| Criterion | Assessment |
| --- | --- |
| Ownership | **Platform / actor reference** — OK |
| Aggregate | **Correct** append-only |
| Lifecycle | **Immutable** |
| Delete | **Never** (product) — OK |
| Indexes | **Good** |
| Expansion | **Yes** |
| Provider leak? | **No** |

**Concern:** Polymorphism is justified; validate `subject_type` allow-list in app.

### AI Session (placeholder)

| Criterion | Assessment |
| --- | --- |
| Ownership | **Planned clear** |
| Aggregate | **Logical AI** — OK as placeholder |
| Lifecycle | **Defined in placeholder** |
| Delete | **Retention purge** — OK |
| Indexes | **Adequate for later** |
| Expansion | **Yes** |
| Provider leak? | **No** — provider is discriminator |

**Concern:** Do not invent tables until usage volume demands; logs may suffice initially — **accepted**.

### Integration Binding (placeholder)

| Criterion | Assessment |
| --- | --- |
| Ownership | **Clear** |
| Aggregate | **Under Connected Account** — OK |
| Provider leak? | **No** — prevents Slack-as-Calendar-Source mistake |

**Concern:** None — good forward slot.

---

# Relationship Validation

| Relationship | Cardinality | Ambiguity risk | Recommendation |
| --- | --- | --- | --- |
| User → Settings | 1:1 | Low | Keep required Settings row on user create |
| User → Connected Accounts | 1:N | Low | OK |
| Connected Account → Calendar Sources | 1:N | Low | OK |
| Calendar Source → Events | 1:N | Low | OK |
| User → Tasks / Conversations / Memories / Notifications / Jobs | 1:N | Low | OK |
| Conversation → Messages | 1:N | Low | OK |
| Task → Reminders | 1:N | Medium | OK if Reminder.task_id set |
| Event → Reminders | 1:N | Medium | OK if calendar_event_id set |
| Reminder → Task **and** Event | N:1 + N:1 | **High** | Document: allowed for “task due at meeting”; otherwise prefer one. No M:N junction needed yet |
| Notification → Reminder/Event/Job/Conversation | N:1 optional | Low–Medium | Prefer keeping explicit FKs; do **not** collapse to polymorphic subject on Notification |
| Message → Task/Event/Memory | Soft M:N via citations JSON | Medium | If querying “messages that cite task X” matters, replace with **message_citations** junction (R-05) |
| Audit → subjects | Polymorphic | Accepted | Keep |
| Settings → default Calendar Sources | N:1 optional | Low | Validate same user_id |
| Job → Reminder/Account | N:1 optional | Low | Prefer FKs over payload-only ids |

### Circular dependencies

No ownership cycles. **Logical cycles to watch in application code:** Message cites Task that cites Message provenance — citations must be write-once soft links, not bidirectional ownership.

### Hidden M:N

None required for Phase 1 cores. Future: User↔Workspace membership junction; Message↔Entity citations junction if needed.

---

# Aggregate Review

### Identity

| Question | Answer |
| --- | --- |
| Root | **User** |
| Inside | User, Settings, Connected Account; (**missing**) Auth Identity |
| Can others modify? | Calendar must **not** mutate credentials; only Identity/Auth services |
| Independence | **Mostly** — add Auth Identity; treat Account credential updates as Identity-only |

### Calendar

| Question | Answer |
| --- | --- |
| Root | **Calendar Source** (events consistency per source); User is tenancy only |
| Inside | Calendar Source, Calendar Event |
| Can others modify? | Sync adapters + Calendar business services only; Chat/AI call API tools which invoke Calendar service — **not** direct repo access from Chat |
| Independence | **Yes** |

### Task

| Question | Answer |
| --- | --- |
| Root | **Task** |
| Inside | Task, Reminder (when task-scoped) |
| Can others modify? | Only Task/Reminder services; AI via tools |
| Independence | **Yes**, with Reminder dual-FK caveat |

### Conversation

| Question | Answer |
| --- | --- |
| Root | **Conversation** |
| Inside | Conversation, Message |
| Can others modify? | No — Memory may reference Message ids read-only |
| Independence | **Yes** |

### Memory

| Question | Answer |
| --- | --- |
| Root | **Memory** |
| Inside | Memory (+ embedding side-structure) |
| Can others modify? | Only Memory service (API); AI proposes |
| Independence | **Yes** |

### Notification

| Question | Answer |
| --- | --- |
| Root | **Notification** |
| Inside | Notification |
| Can others modify? | Scheduler/Reminder services create; client marks read/dismissed |
| Independence | **Yes** |

### Scheduler

| Question | Answer |
| --- | --- |
| Root | **Scheduler Job** |
| Inside | Job records |
| Can others modify? | Scheduler worker owns status transitions; domain services invoked as effects |
| Independence | **Yes** — must not embed Event/Task rows |

### Audit

| Question | Answer |
| --- | --- |
| Root | **Audit Log** (append stream) |
| Inside | Audit Log |
| Can others modify? | **No updates**; writers append only |
| Independence | **Yes** |

### AI

| Question | Answer |
| --- | --- |
| Root | **AI Session** (when materialized) |
| Inside | AI Session (+ future steps) |
| Can others modify? | API records usage; AI service does not write DB |
| Independence | **Yes** — must never own Message/Task |

**Aggregate independence verdict:** Valid, contingent on enforcing module boundaries in code (repos not shared across aggregates) and fixing Reminder write ownership.

---

# Provider Independence Review

| Scenario | Outcome |
| --- | --- |
| Google Calendar disappears | **Donna continues** for Donna-originated events, tasks, chat, memory, local reminders. Sync jobs fail; Connected Account → `needs_reauth`/`disconnected`. Events with `origin=donna` remain. Synced-only copies remain as last-known Donna Events until policy archive. |
| OpenAI changes / removed | **Conversations/Messages remain.** Swap provider via AI adapters; AI Session/provider fields change. No Conversation schema change. |
| Telegram removed | **Notifications remain** (`channel` values). Web push unaffected. Conversation `channel=web` unaffected. |
| Outlook added | **No core schema redesign** — new `provider` value, Connected Account, Sources, Event provider ids, Calendar adapter. |
| Google Tasks / To Do added | Task `provider_*` fields absorb links; optional Integration Binding `capability=tasks`. |

### Leak hunt

| Location | Verdict |
| --- | --- |
| Entity names | **Clean** — no `google_*` tables |
| `provider` enums | **OK** — discriminator |
| `provider_payload` | **OK** if opaque and non-canonical |
| Login conflated with Google Calendar | **Risk** until Auth Identity exists (R-01) |
| Notification FCM-specific columns | **None** — good |
| Message OpenAI ids as PK | **None** — good |

---

# Lifecycle Review

### Task

```text
Created (open, optionally backlog)
    ↓
Open / active work          ← optional future: in_progress
    ↓
Completed  ↔  reopen → Open
    ↓
Cancelled
    ↓
Soft-deleted (deleted_at)
    ↓
Hard purge (retention)
```

**Gap:** DATA_MODEL lacks `in_progress` / `archived`. Either add statuses or document that `open` covers in-progress and backlog flag covers archive-like behavior (R-08).

### Conversation

```text
Created (active)
    ↓
Active (messages append)
    ↓
Archived
    ↓
Soft-deleted
```

**OK.**

### Notification

```text
Created (pending)
    ↓
Queued (optional — may equal pending)
    ↓
Sent / Failed
    ↓
Read (user)
    ↓
Dismissed
    ↓
Soft-deleted / retention purge
```

**Issue:** Single `status` enum cannot be both `sent` and `read` without overwrite (R-04). Prefer:

- `delivery_status`: pending → sent → failed  
- `read_at` / `dismissed_at` timestamps (already partially present)

### Scheduler Job

```text
Created (pending)
    ↓
Scheduled (run_at in future)  ← pending covers this
    ↓
Running
    ↓
Succeeded → retain → Hard purge
    or Failed → Retry (attempt_count++) → Running …
    or Cancelled → purge
```

**OK.** Align naming: use `pending` not separate `scheduled` unless you add it.

### Reminder

```text
Created (scheduled)
    ↓
Sent  |  Failed  |  Cancelled
    ↓
Soft-deleted
```

Job retries are Scheduler concerns; Reminder `failed` may flip back to `scheduled` on reschedule — document transition.

### Calendar Event

```text
Created (confirmed|tentative)
    ↓
Updated (sync or Donna)
    ↓
Cancelled (status) and/or Soft-deleted (tombstone)
```

**OK** if product distinguishes cancel vs delete.

### User

```text
Created (active)
    ↓
Disabled ↔ Active
    ↓
pending_deletion
    ↓
Soft-deleted → anonymize children → Hard purge
```

**OK** — expand in Auth milestone.

### Memory

```text
Created → Updated → Soft-deleted → Purge
```

**OK.**

---

# Index Review (query perspective)

| Access pattern | Conceptual indexes |
| --- | --- |
| Login by email | Unique partial `User.email` where live |
| Resolve public id | Unique `public_id` per entity |
| Dashboard today’s tasks | `(user_id, status, due_at)` incl. backlog filter; consider `(user_id, is_backlog, status, due_at)` |
| Upcoming events | `(user_id, starts_at)` and/or `(calendar_source_id, starts_at, ends_at)` |
| Recent messages / history | `(conversation_id, created_at)` |
| Conversation list | `(user_id, last_message_at DESC)` |
| Unread notifications | `(user_id, status, created_at DESC)` — **revisit after R-04** (index `read_at IS NULL`) |
| Calendar sync upsert | Unique `(calendar_source_id, provider_event_id)` live |
| Account sync | Unique `(provider, provider_account_id)`; `(user_id, status)` |
| Reminder processing | `(status, remind_at)` / `(remind_at, status)` |
| Scheduler poll | `(status, run_at)` |
| Memory retrieval (keyword) | `(user_id, updated_at)`; later vector index on embedding |
| Audit by actor/subject | `(actor_user_id, created_at)`; `(subject_type, subject_id, created_at)` |
| Job by reminder | `reminder_id` |

**Missing / strengthen:** partial indexes for soft-delete uniqueness (called out in DATABASE.md) — mandatory at SQL time.

---

# JSON Review

| Field | Entity | Why JSON? | Explicit columns instead? | Validation? | Index later? | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `provider_metadata` | Connected Account / Source | Provider quirks | Only if a key becomes queryable | Schema-less allow-list size cap | Unlikely | **Keep** |
| `provider_payload` | Event / Task | Opaque sync repair | Never as canonical | Size cap; no secrets | No | **Keep** |
| `attendees_summary` | Event | Variable attendee list | Child `event_attendees` if CRM/search needed | Max N attendees; email/name shape | GIN only if search | **Keep thin**; promote if product searches attendees |
| `citations` | Message | Avoid M:N early | **Junction recommended if queried** | UUID allow-list | Junction indexes | **Conditional reject as long-term** (R-05) |
| `preferences` | Settings | Sparse flags | Promote when product-stable | Known keys only | No | **Keep** with key allow-list |
| `payload` | Notification / Job | Channel/job params | Prefer FKs already added; payload for extras only | No secrets; size cap | No | **Keep** but forbid ids-only-in-JSON when FK exists (R-03) |
| `tools_used` | AI Session | List | Array column OK | Enum tool names | Rarely | **Keep** as array |
| `scopes` | Connected Account | Listed as String/String[] | **Prefer text array / separate rows** | Known scope strings | Filter rarely | **Normalize** — not freeform JSON string |
| `config` | Integration Binding | Capability config | Promote hot keys | Per-capability schema | Maybe | **Keep** (future) |
| `metadata` | Audit | Sparse context | No | Redaction required | GIN optional | **Keep** |

**Rejected pattern:** Entire Event/Task body as JSON. **Not present** — good.

---

# Deletion Strategy Review

| Entity | Strategy | Why | Conceptual FK behavior from parent |
| --- | --- | --- | --- |
| User | Soft → anonymize → Hard purge | GDPR + recovery | Children: soft-delete or anonymize (Restrict hard delete while children live) |
| Connected Account | Soft | Reconnect / audit | Sources: Cascade soft-delete (app-level) |
| Calendar Source | Soft | Resubscribe | Events: Cascade soft-delete / tombstone |
| Calendar Event | Soft (+ cancel status) | Sync tombstones | Reminders: Set Null or Cascade soft — **decide (R-02)** |
| Task | Soft | UX recovery | Reminders: Cascade soft |
| Reminder | Soft | History | Notifications: Set Null on reminder_id or keep historical FK Restrict |
| Conversation | Soft | History | Messages: Cascade soft |
| Message | Soft | Moderation | Citations: ignore |
| Memory | Soft | User control | — |
| Notification | Soft; failed may Hard purge | Inbox vs ops noise | — |
| Scheduler Job | Hard purge after retention | Ephemeral | — |
| Settings | Never alone; with User | 1:1 | Cascade with User soft |
| Audit Log | Never (product) | Compliance | No FK cascade from subjects |
| AI Session | Retention Hard purge | Cost logs | — |
| Integration Binding | Soft | Disconnect capability | — |

**Restrict** hard deletes of User while live children exist. **No Action/Restrict** from Audit to subjects (subjects may soft-delete). **Set Null** optional correlation FKs on Notification when source Reminder hard-purged — prefer soft everywhere to avoid Set Null complexity.

---

# Findings Register (must address before / during first schema)

| ID | Severity | Finding | Recommendation |
| --- | --- | --- | --- |
| **R-01** | **High** | No Auth Identity / IdP subject entity | Add logical entity before Auth SQL: `(user_id, provider, subject, email)` unique |
| **R-02** | **Medium** | Reminder dual Task+Event FKs; cascade unclear | Document write owner; pick Cascade soft from Task; Set Null or soft from Event |
| **R-03** | **Medium** | Reminder.remind_at vs Job.run_at dual schedule | Job derives from Reminder; Reminder is source of truth for fire time |
| **R-04** | **Medium** | Notification status mixes delivery + read | Use delivery_status + read_at/dismissed_at |
| **R-05** | **Low–Medium** | Message citations as JSON | Accept v1; plan `message_citations` if queried |
| **R-06** | **Medium** (product) | Planning/Notes absent from DATA_MODEL detail | Add logical entities before accountability milestone; don’t abuse Memory |
| **R-07** | **Low** | Event.provider redundant | Optional denormalization; document or drop |
| **R-08** | **Low–Medium** | Task statuses vs lifecycle vocabulary | Freeze enum with product before tools ship |
| **R-09** | **Low** | Settings.public_id low value | Keep for consistency or drop |
| **R-10** | **Low** | Connected Account.scopes typing | Use string array, not JSON string |

---

# Migration Readiness Assessment

| Criterion | Status | Notes |
| --- | --- | --- |
| Business model complete | ✓ | DOMAIN_MODEL solid |
| Logical model complete | ✓ with gaps | DATA_MODEL strong; R-01/R-06 gaps |
| Ownership validated | ✓ | With Reminder caveat R-02 |
| Relationships validated | ✓ | Citations/Reminder noted |
| Aggregate boundaries validated | ✓ | Enforce in code |
| Future integrations supported | ✓ | Provider independence holds |
| Delete strategy defined | ✓ | Table above; cascade choices R-02 |
| Public IDs defined | ✓ | Prefix registry |
| Index strategy reviewed | ✓ | Query matrix above |
| Provider abstraction validated | ✓ | Leak risks limited to auth modeling |
| JSON discipline validated | ✓ | Conditional items listed |
| Lifecycles documented | ✓ | Fix Notification/Task enums |
| **Ready for SQL generation** | **Conditional Yes** | Proceed to schema SQL **after** resolving **R-01** and **R-04** (or accepting R-04 with documented state machine), and recording decisions for **R-02/R-03**. R-05/R-06/R-07/R-08/R-09/R-10 may be accepted with tickets. |

### Recommendation

**Do not generate migrations in this milestone.**  

**Next implementation milestone** should:

1. Amend [DATA_MODEL.md](./DATA_MODEL.md) for R-01 (Auth Identity) and R-04 (Notification fields) — small doc PR.  
2. Add ADR acceptance notes for R-02/R-03 into this file or DATA_MODEL.  
3. Then generate SQL/migrations mapped 1:1 from [PHYSICAL_DATABASE_DESIGN.md](./PHYSICAL_DATABASE_DESIGN.md) under [DATABASE.md](./DATABASE.md).

The architecture is **production-credible** for a 5–10 year Personal AI OS **if** provider independence and aggregate write discipline are enforced in code—not only in docs.
