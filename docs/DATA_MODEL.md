# Donna Logical Data Model

**Status:** Source of truth for the logical database design (pre-migration)  
**Milestone:** M2.2 — Logical Data Model (documentation only)  
**Non-goals:** SQL, `CREATE TABLE`, migrations, repositories, Go models, DTOs, handlers  

This document translates [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) into a field-level logical design. Persistence conventions remain in [DATABASE.md](./DATABASE.md). Observability correlation fields follow [OBSERVABILITY.md](./OBSERVABILITY.md).

When schema implementation begins, every migration MUST map to entities here. Formal review and ADRs: [SCHEMA_DECISIONS.md](./SCHEMA_DECISIONS.md). If reality diverges, update this document deliberately.

---

## Identifier conventions (applies to all entities)

| Kind | Role |
| --- | --- |
| **Internal primary key** | UUIDv7 — time-sortable, opaque, never reused |
| **Public identifier** | Prefixed stable string for APIs, URLs, logs, support (e.g. `usr_…`) |
| **Provider remote id** | Opaque string from Google/Outlook/etc. — never Donna’s PK |

Public IDs are unique globally (or unique per entity type). Internal UUIDs remain the only foreign-key targets in the logical model.

**Prefix registry**

| Prefix | Entity |
| --- | --- |
| `usr_` | User |
| `acct_` | Connected Account |
| `cal_` | Calendar Source |
| `evt_` | Calendar Event |
| `tsk_` | Task |
| `rem_` | Reminder |
| `conv_` | Conversation |
| `msg_` | Message |
| `mem_` | Memory |
| `ntf_` | Notification |
| `job_` | Scheduler Job |
| `set_` | Settings |
| `audit_` | Audit Log |
| `ais_` | AI Session (future placeholder) |
| `int_` | Integration Capability binding (future placeholder) |

---

# User

## Purpose

The Donna account — tenancy root for Phase 1. Represents the human who owns all personal data. Login identity (e.g. Google OAuth for sign-in) creates or resolves a User; it does not imply a calendar connection.

## Ownership

- **Owned by:** nothing (root). Future: may belong to a Workspace as a member without losing personal ownership of private rows.
- **May be referenced by:** every user-scoped entity (`user_id`).

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal primary key | FK target for all children |
| `public_id` | String | Required | Public id `usr_…` | Unique; API-facing |
| `email` | String | Required | Primary email | Unique among non-deleted users |
| `email_verified` | Boolean | Required | Whether email verified via IdP | Default false until verified |
| `display_name` | String | Optional | Human-readable name | From IdP or user edit |
| `avatar_url` | String | Optional | Profile image URL | May be IdP-hosted; not binary blob |
| `timezone` | String | Required | IANA timezone | e.g. `Asia/Kolkata` |
| `locale` | String | Optional | BCP 47 locale | UI/copy preference |
| `status` | String (enum) | Required | `active` \| `disabled` \| `pending_deletion` | Soft lifecycle |
| `last_login_at` | Timestamp | Optional | Last successful auth | Observability / security |
| `created_at` | Timestamp | Required | Created | |
| `updated_at` | Timestamp | Required | Last mutation | |
| `deleted_at` | Timestamp | Optional | Soft delete | Null = live |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| One-to-Many | Connected Accounts, Calendar Sources (via accounts), Tasks, Conversations, Memories, Notifications, Scheduler Jobs, Audit Logs | User owns |
| One-to-One | Settings | User owns |
| One-to-Many | Calendar Events (denormalized `user_id`) | User owns via Sources |
| Many-to-One | — | Root |

## Constraints

- Unique `email` where `deleted_at` is null
- Unique `public_id`
- Unique `id`
- `status` in allowed set

## Index Recommendations

- `email` (unique partial: live rows)
- `public_id` (unique)
- `status`
- `created_at`

## Soft Delete Policy

**Soft delete.** Account closure and recovery/GDPR workflows need historical integrity and cascading soft-delete or anonymization of children. Hard delete only via explicit retention job.

## Public Identifier Strategy

`usr_` + opaque suffix. Used in `/api/v1/me`, admin tools, and log correlation (`user_id` may be UUID internally; public id preferred in URLs).

## Future Expansion

- Multiple login identities (Apple, Microsoft) as child **Auth Identities** without splitting User
- `workspace_memberships` for Teams
- Billing customer link by `user_id` / later `workspace_id`

---

# Connected Account

## Purpose

A third-party account the user authorized for **integrations** (calendar sync today; Slack/Notion/etc. later). Distinct from login IdP used only for authentication.

## Ownership

- **Owned by:** User
- **Referenced by:** Calendar Sources, Scheduler Jobs (sync), Audit Logs, future mail/task sources

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `acct_…` | Unique |
| `user_id` | UUID | Required | Owning user | FK → User |
| `provider` | String (enum) | Required | `google` \| `microsoft` \| `apple` \| … | Extensible enum/text |
| `provider_account_id` | String | Required | Remote account id | Stable per provider |
| `display_name` | String | Optional | “Work Google”, “Personal” | User-facing label |
| `status` | String (enum) | Required | `active` \| `needs_reauth` \| `revoked` \| `disconnected` | |
| `scopes` | String / String[] | Optional | Granted OAuth scopes | Store as list or normalized later |
| `credentials_ref` | String | Required | Reference to encrypted secret blob / vault key | Never store raw refresh token in plain fields |
| `token_expires_at` | Timestamp | Optional | Access token expiry if tracked | |
| `last_synced_at` | Timestamp | Optional | Last successful sync tick | |
| `provider_metadata` | JSON | Optional | Non-secret provider quirks | Dynamic only |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | Soft delete / disconnect | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent owns |
| One-to-Many | Calendar Sources | Account owns sources |
| One-to-Many | (future) Task lists / other sources | Same pattern |

## Constraints

- Unique `(provider, provider_account_id)` among live rows (or unique per `user_id` + provider + remote id)
- Unique `public_id`
- User must exist
- Login Google ≠ auto-create Connected Account (product rule)

## Index Recommendations

- `(user_id, status)`
- `(provider, provider_account_id)` unique (partial live)
- `public_id` unique
- `user_id`

## Soft Delete Policy

**Soft delete** on disconnect (retain sync history links). Credentials wiped or rotated on revoke even if row soft-deleted.

## Public Identifier Strategy

`acct_…` for connection management APIs and audit subjects.

## Future Expansion

- Same entity supports Outlook/Apple/Slack by `provider` value
- Capability flags in metadata or child **Integration Binding** rows (`int_`) without new account types

---

# Calendar Source

## Purpose

A single calendar feed/container under a Connected Account (e.g. “Primary”, “Work”) that Donna syncs into unified events. Also the target of default routing (personal / work / reminder).

## Ownership

- **Owned by:** Connected Account (and thus User); denormalized `user_id` recommended
- **Referenced by:** Calendar Events, Settings (default source pointers), Scheduler sync jobs

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `cal_…` | Unique |
| `user_id` | UUID | Required | Owning user | Denormalized for tenancy queries |
| `connected_account_id` | UUID | Required | Parent connection | FK |
| `provider_calendar_id` | String | Required | Remote calendar id | |
| `name` | String | Required | Display name | |
| `color` | String | Optional | UI color token/hex | |
| `is_primary_on_provider` | Boolean | Required | Provider’s primary flag | Informational |
| `sync_enabled` | Boolean | Required | Whether Donna syncs this source | |
| `sync_token` / `sync_cursor` | String | Optional | Incremental sync cursor | Opaque |
| `last_synced_at` | Timestamp | Optional | | |
| `timezone` | String | Optional | Source default TZ | Fallback to user |
| `provider_metadata` | JSON | Optional | Provider-only extras | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | Connected Account, User | Parents |
| One-to-Many | Calendar Events | Source owns events |

## Constraints

- Unique `(connected_account_id, provider_calendar_id)` live rows
- Unique `public_id`
- `sync_enabled` false ⇒ sync jobs skip

## Index Recommendations

- `(user_id, sync_enabled)`
- `(connected_account_id)`
- `(connected_account_id, provider_calendar_id)` unique partial
- `public_id` unique

## Soft Delete Policy

**Soft delete** when user unsubscribes a calendar or account disconnects; events soft-deleted or marked read-only per product policy at schema time.

## Public Identifier Strategy

`cal_…` for default calendar settings and “create in Office calendar” targeting.

## Future Expansion

- Read-only subscribed calendars (holidays) as Sources with `sync_enabled` create=false flags in metadata
- Apple/Outlook calendars identical shape

---

# Calendar Event

## Purpose

Donna’s **canonical** calendar event — what the dashboard, AI tools, and phone reason about. Provider events are links, not the primary meaning.

## Ownership

- **Owned by:** Calendar Source → Connected Account → User (`user_id` denormalized)
- **Referenced by:** Reminders (optional), Messages (citations), Notifications, Scheduler Jobs

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `evt_…` | Unique |
| `user_id` | UUID | Required | Owner | Tenancy |
| `calendar_source_id` | UUID | Required | Parent source | FK |
| `title` | String | Required | Event title | Explicit field |
| `description` | String | Optional | Body/notes | |
| `location` | String | Optional | | |
| `starts_at` | Timestamp | Required | Start instant | timestamptz conceptually |
| `ends_at` | Timestamp | Required | End instant | |
| `is_all_day` | Boolean | Required | All-day flag | |
| `status` | String (enum) | Required | `confirmed` \| `tentative` \| `cancelled` | |
| `visibility` | String (enum) | Optional | `default` \| `private` \| `public` | |
| `attendees_summary` | JSON | Optional | Lightweight attendee list | Not full CRM |
| `recurrence_rule` | String | Optional | RRULE or Donna summary | Explicit when possible |
| `recurring_event_id` | UUID | Optional | Series parent | Self-FK optional |
| `provider` | String | Optional | Redundant convenience | Usually via source |
| `provider_event_id` | String | Optional | Remote event id | Required if synced |
| `provider_etag` | String | Optional | Version / etag | Sync concurrency |
| `provider_payload` | JSON | Optional | Opaque provider snapshot | Non-canonical |
| `origin` | String (enum) | Required | `donna` \| `provider_sync` | Who last authored meaning |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | Soft delete / tombstone | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | Calendar Source, User | Parents |
| One-to-Many | Reminders (optional FK) | Event may be referenced |
| Many-to-One (optional) | Task | Soft association if scheduled block links a task — prefer nullable `task_id` only if product needs it; else citation via Message |

Avoid polymorphic “subject_id” for core links; use nullable FKs.

## Constraints

- Unique `(calendar_source_id, provider_event_id)` when `provider_event_id` present (live rows)
- `ends_at` ≥ `starts_at` (or equal for point events — product rule)
- Unique `public_id`

## Index Recommendations

- `(user_id, starts_at)`
- `(calendar_source_id, starts_at, ends_at)`
- `(calendar_source_id, provider_event_id)` unique partial
- `(user_id, status)` where useful
- `public_id` unique

## Soft Delete Policy

**Soft delete** for sync tombstones and user deletes (recovery + provider sync). Cancelled may use `status=cancelled` without delete.

## Public Identifier Strategy

`evt_…` in APIs and AI tool results.

## Future Expansion

- Outlook/Apple = same entity + provider ids
- Attachments as child entity later
- Conference link as explicit field when commonly queried

---

# Task

## Purpose

Actionable work item: quick todo, backlog item, prioritized commitment.

## Ownership

- **Owned by:** User
- **Referenced by:** Reminders, Messages (citations), Daily Plans (future), Notifications

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `tsk_…` | Unique |
| `user_id` | UUID | Required | Owner | FK |
| `title` | String | Required | | |
| `description` | String | Optional | | |
| `status` | String (enum) | Required | `open` \| `completed` \| `cancelled` | |
| `priority` | String (enum) | Optional | `low` \| `medium` \| `high` | |
| `due_at` | Timestamp | Optional | Due instant | |
| `completed_at` | Timestamp | Optional | When completed | |
| `is_backlog` | Boolean | Required | Backlog vs active | |
| `recurrence_rule` | String | Optional | Data-ready recurrence | UI may be basic |
| `provider` | String | Optional | Future Google Tasks / To Do | |
| `provider_task_id` | String | Optional | Remote id | |
| `provider_payload` | JSON | Optional | Opaque sync | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent |
| One-to-Many | Reminders | Task owns reminders when `task_id` set |

## Constraints

- Unique `public_id`
- Unique `(user_id, provider, provider_task_id)` when provider ids present
- Completed ⇒ `completed_at` set (application rule)

## Index Recommendations

- `(user_id, status)`
- `(user_id, due_at)`
- `(user_id, is_backlog, status)`
- `public_id` unique

## Soft Delete Policy

**Soft delete.** Users recover mistaken deletes; AI-created tasks need audit trail.

## Public Identifier Strategy

`tsk_…` for tools `create_task` / `complete_task` and dashboard URLs.

## Future Expansion

- Google Tasks / Microsoft To Do sync via provider fields
- Subtasks as child rows later (`parent_task_id`) without redesigning Task
- Goal linkage via nullable `goal_id` when Planning schema lands

---

# Reminder

## Purpose

“Nudge at time T about X” — first-class commitment. Delivery is Notifications; timing policy is Reminder.

## Ownership

- **Owned by:** User (always); optionally associated with Task and/or Calendar Event
- **Referenced by:** Scheduler Jobs, Notifications, Audit Logs

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `rem_…` | Unique |
| `user_id` | UUID | Required | Owner | FK |
| `task_id` | UUID | Optional | Associated task | FK; prefer for task nudges |
| `calendar_event_id` | UUID | Optional | Associated event | FK |
| `title` | String | Required | Reminder text | May copy from task/event |
| `remind_at` | Timestamp | Required | When to fire | |
| `status` | String (enum) | Required | `scheduled` \| `sent` \| `cancelled` \| `failed` | |
| `channel_preference` | String | Optional | Override channel | Else user settings |
| `last_error` | String | Optional | Last failure reason | Non-secret |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Required parent |
| Many-to-One | Task | Optional |
| Many-to-One | Calendar Event | Optional |
| One-to-Many | Notifications / Jobs | Reminder referenced |

**Rule:** At least one of (`task_id`, `calendar_event_id`, standalone title context) — standalone reminders allowed with title only.

## Constraints

- Unique `public_id`
- Not both required; avoid polymorphic subject — use explicit nullable FKs
- `remind_at` required when `status=scheduled`

## Index Recommendations

- `(user_id, status, remind_at)`
- `(remind_at, status)` for due scanning
- `task_id`, `calendar_event_id`
- `public_id` unique

## Soft Delete Policy

**Soft delete** on user cancel/history. Scheduler skips deleted.

## Public Identifier Strategy

`rem_…` for APIs and audit subjects.

## Future Expansion

- Recurring reminders via `recurrence_rule`
- Telegram channel preference without new entity
- Provider-native popup sync remains non-authoritative

---

# Conversation

## Purpose

Phone-chat thread between the user and Donna (morning/midday/evening or freeform).

## Ownership

- **Owned by:** User
- **Referenced by:** Messages, AI Sessions (future), Memories (provenance)

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `conv_…` | Unique |
| `user_id` | UUID | Required | Owner | FK |
| `title` | String | Optional | Optional thread title | |
| `purpose` | String (enum) | Optional | `general` \| `morning` \| `midday` \| `evening` \| `system` | |
| `status` | String (enum) | Required | `active` \| `archived` | |
| `unread_count` | Integer | Required | Unread Donna messages | Denormalized counter |
| `last_message_at` | Timestamp | Optional | For list sorting | |
| `channel` | String (enum) | Required | `web` (Phase 1) | Future: `telegram` |
| `channel_thread_id` | String | Optional | External thread id | Future channels |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent |
| One-to-Many | Messages | Conversation owns |

## Constraints

- Unique `public_id`
- One active “primary phone” conversation may be enforced in app (optional unique partial later)

## Index Recommendations

- `(user_id, last_message_at DESC)`
- `(user_id, status)`
- `public_id` unique

## Soft Delete Policy

**Soft delete** (archive + delete). Messages inherit visibility rules.

## Public Identifier Strategy

`conv_…` in chat APIs `/conversations/:public_id`.

## Future Expansion

- Telegram/WhatsApp as `channel` + `channel_thread_id`
- Multiple parallel conversations without schema break

---

# Message

## Purpose

One turn in a conversation: user, Donna (assistant), or system.

## Ownership

- **Owned by:** Conversation → User (`user_id` denormalized)
- **Referenced by:** Memories (source), AI Sessions (future)

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `msg_…` | Unique |
| `user_id` | UUID | Required | Tenancy | Denormalized |
| `conversation_id` | UUID | Required | Parent | FK |
| `role` | String (enum) | Required | `user` \| `assistant` \| `system` | |
| `content` | String | Required | User-visible text | Explicit field |
| `content_format` | String | Optional | `plain` \| `markdown` | Default plain |
| `client_message_id` | String | Optional | Idempotency from client | Unique per conversation |
| `citations` | JSON | Optional | `{task_ids, event_ids, memory_ids}` | Avoid polymorphic FKs |
| `token_usage_ref` | String / UUID | Optional | Link to AI usage/session | Not full trace |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | Hide without destroy | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | Conversation, User | Parents |
| Soft refs | Task / Event / Memory via `citations` JSON ids | Not ownership |

## Constraints

- Unique `public_id`
- Unique `(conversation_id, client_message_id)` when client id present
- `role` in allowed set

## Index Recommendations

- `(conversation_id, created_at)`
- `(user_id, created_at)`
- `public_id` unique

## Soft Delete Policy

**Soft delete** for moderation/user delete; retain for audit/memory provenance as needed.

## Public Identifier Strategy

`msg_…` for deep links and debugging.

## Future Expansion

- `parts` JSON or child **Message Part** rows for multimodal (voice transcript, images)
- Streaming: persist final Message only

---

# Memory

## Purpose

Durable knowledge Donna remembers across days (projects, people, preferences, commitments).

## Ownership

- **Owned by:** User
- **Referenced by:** AI retrieval tools; optional provenance from Message/Conversation

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `mem_…` | Unique |
| `user_id` | UUID | Required | Owner | FK |
| `content` | String | Required | Memory text / fact | Explicit |
| `category` | String (enum) | Optional | `preference` \| `person` \| `project` \| `commitment` \| `idea` \| `other` | |
| `importance` | Integer / enum | Optional | Ranking for retrieval | |
| `source` | String (enum) | Required | `explicit` \| `chat_extract` \| `review` \| `system` | |
| `source_conversation_id` | UUID | Optional | Provenance | FK soft |
| `source_message_id` | UUID | Optional | Provenance | FK soft |
| `embedding` | Vector / Binary | Optional | Semantic search | Side structure allowed later |
| `embedding_model` | String | Optional | Model id used | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent |
| Many-to-One (optional) | Conversation, Message | Provenance only |

## Constraints

- Unique `public_id`
- AI service never inserts directly — API only

## Index Recommendations

- `(user_id, category)`
- `(user_id, updated_at DESC)`
- Vector index later on `embedding`
- `public_id` unique

## Soft Delete Policy

**Soft delete.** Memories are sensitive long-lived data; user may resurrect.

## Public Identifier Strategy

`mem_…` for `save_memory` / `search_memory` tools.

## Future Expansion

- Swap embedding providers via `embedding_model`
- Memory graph / relations as association table later

---

# Notification

## Purpose

User-visible alert and delivery attempt record across channels.

## Ownership

- **Owned by:** User
- **Referenced by:** (references Reminder / Event / Job / Conversation — does not own them)

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `ntf_…` | Unique |
| `user_id` | UUID | Required | Owner | FK |
| `channel` | String (enum) | Required | `browser_push` \| `email` \| `telegram` \| … | Phase 1: browser_push |
| `title` | String | Required | | |
| `body` | String | Required | | |
| `priority` | String (enum) | Optional | `low` \| `normal` \| `high` | |
| `status` | String (enum) | Required | `pending` \| `sent` \| `failed` \| `read` \| `dismissed` | |
| `reminder_id` | UUID | Optional | Correlation | Explicit FK |
| `calendar_event_id` | UUID | Optional | Correlation | |
| `scheduler_job_id` | UUID | Optional | Correlation | |
| `conversation_id` | UUID | Optional | Correlation | |
| `payload` | JSON | Optional | Channel-specific non-secret data | |
| `sent_at` | Timestamp | Optional | | |
| `read_at` | Timestamp | Optional | | |
| `dismissed_at` | Timestamp | Optional | | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent |
| Many-to-One (optional) | Reminder, Calendar Event, Scheduler Job, Conversation | Correlation only |

## Constraints

- Unique `public_id`
- Prefer explicit nullable FKs over polymorphic `subject_type/subject_id` for core product links

## Index Recommendations

- `(user_id, status, created_at DESC)`
- `(user_id, channel)`
- `reminder_id`, `scheduler_job_id`
- `public_id` unique

## Soft Delete Policy

**Soft delete** for user clear-history; delivery logs may hard-delete after retention for `failed` noise.

## Public Identifier Strategy

`ntf_…` for client dismiss/read APIs.

## Future Expansion

- New channels = enum values + adapters
- Digests batching without new notification core

---

# Scheduler Job

## Purpose

Record of scheduled or queued work: briefings, reminder firing, calendar sync ticks.

## Ownership

- **Owned by:** User (tenant scope); system jobs may use null user only if ever introduced (prefer always user-scoped in Phase 1)
- **References:** Reminder, Connected Account, Conversation as payload targets — not ownership of those aggregates

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `job_…` | Unique |
| `user_id` | UUID | Required | Scope | FK |
| `job_type` | String (enum) | Required | `morning_briefing` \| `midday_checkin` \| `evening_reflection` \| `reminder_fire` \| `calendar_sync` \| … | |
| `status` | String (enum) | Required | `pending` \| `running` \| `succeeded` \| `failed` \| `cancelled` | |
| `run_at` | Timestamp | Required | When due | |
| `attempt_count` | Integer | Required | Retry count | Default 0 |
| `max_attempts` | Integer | Required | Cap | |
| `payload` | JSON | Optional | Ids + params (no secrets) | Dynamic |
| `reminder_id` | UUID | Optional | Explicit link | |
| `connected_account_id` | UUID | Optional | Sync jobs | |
| `last_error` | String | Optional | | |
| `started_at` | Timestamp | Optional | | |
| `finished_at` | Timestamp | Optional | | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User | Parent |
| Many-to-One (optional) | Reminder, Connected Account | Reference |

## Constraints

- Unique `public_id`
- Idempotency keys inside `payload` for sync/webhooks recommended

## Index Recommendations

- `(status, run_at)` — poller
- `(user_id, job_type, status)`
- `public_id` unique

## Soft Delete Policy

**Hard delete** (or retain then purge) after retention window for succeeded jobs. **No soft delete** required for ephemeral runs; keep failed rows longer for ops. Optionally archive table later.

## Public Identifier Strategy

`job_…` for observability (`job_id` / `scheduler_id` correlation).

## Future Expansion

- Move execution to Redis/queue while keeping Job as durable record
- Billing renewal job types without new core entity

---

# Settings

## Purpose

Per-user preferences: default calendars, quiet hours, briefing times, notification opts.

## Ownership

- **Owned by:** User (one-to-one)
- **Referenced by:** Calendar routing, Scheduler, Notifications, AI context assembly

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `set_…` | Unique |
| `user_id` | UUID | Required | Owner | Unique FK (1:1) |
| `default_personal_calendar_source_id` | UUID | Optional | Default personal | FK → Calendar Source |
| `default_work_calendar_source_id` | UUID | Optional | Default work | FK |
| `default_reminder_calendar_source_id` | UUID | Optional | Default reminder cal | FK |
| `morning_briefing_time` | String / Time | Optional | Local time-of-day | Interpret with user timezone |
| `midday_checkin_time` | String / Time | Optional | | |
| `evening_reflection_time` | String / Time | Optional | | |
| `quiet_hours_start` | String / Time | Optional | | |
| `quiet_hours_end` | String / Time | Optional | | |
| `notifications_enabled` | Boolean | Required | Master toggle | |
| `preferences` | JSON | Optional | Sparse non-core prefs only | Not core routing fields |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| One-to-One | User | User owns |
| Many-to-One (optional) | Calendar Sources (defaults) | Reference |

## Constraints

- Unique `user_id` (exactly one settings row per user)
- Unique `public_id`
- Default calendar sources must belong to same `user_id` (app/DB check)

## Index Recommendations

- `user_id` unique
- `public_id` unique

## Soft Delete Policy

**Never deletes** independently — deleted with User soft-delete cascade / anonymize. No standalone soft delete.

## Public Identifier Strategy

`set_…` rarely in URLs; included for consistency and admin tools.

## Future Expansion

- Workspace settings sibling entity
- Entitlements/billing flags as read model fields, not payment ledger

---

# Audit Log

## Purpose

Append-oriented record of security- and compliance-sensitive actions.

## Ownership

- **Owned by:** platform; **actor** is User (nullable for system)
- **References:** subject entities by type + id/public_id — does not own them

## Fields

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `audit_…` | Unique |
| `actor_user_id` | UUID | Optional | Who acted | Null = system |
| `action` | String | Required | e.g. `auth.login`, `connection.linked` | Stable vocabulary |
| `subject_type` | String | Optional | `user` \| `connected_account` \| `reminder` \| … | Audit exception to polymorphism |
| `subject_id` | UUID | Optional | Internal id of subject | |
| `subject_public_id` | String | Optional | Public id snapshot | Survives deletes |
| `request_id` | String | Optional | Correlation | From observability |
| `metadata` | JSON | Optional | Non-secret context | Redact tokens |
| `created_at` | Timestamp | Required | Immutable | No `updated_at` |

## Relationships

| Type | Related | Ownership |
| --- | --- | --- |
| Many-to-One | User (actor) | Reference |
| Logical | Any subject | Reference by type/id |

Polymorphism is **allowed here only** because audit must reference many aggregates without exploding schema. Core product FKs remain explicit elsewhere.

## Constraints

- Unique `public_id`
- Append-only: no updates in normal operation
- Never store secrets in `metadata`

## Index Recommendations

- `(actor_user_id, created_at DESC)`
- `(subject_type, subject_id, created_at DESC)`
- `(action, created_at DESC)`
- `public_id` unique
- `request_id`

## Soft Delete Policy

**Never deletes** in product flows. Retention/purge via compliance job only (hard delete or cold storage).

## Public Identifier Strategy

`audit_…` for support export and SIEM correlation.

## Future Expansion

- Workspace-scoped audit (`workspace_id`)
- Billing actions reuse same shape
- Export pipelines without redesign

---

# AI Session (future placeholder)

## Purpose

Placeholder for a durable **reasoning session** record linking a conversation turn batch to model/provider usage. Phase 1 may rely on structured logs ([OBSERVABILITY.md](./OBSERVABILITY.md)); this entity is reserved so schema can harden without redesign.

## Ownership

- **Owned by:** User
- **References:** Conversation, Message(s), usage metrics

## Fields (planned)

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `ais_…` | |
| `user_id` | UUID | Required | Owner | |
| `conversation_id` | UUID | Optional | Context | |
| `trigger_message_id` | UUID | Optional | User message that started turn | |
| `provider` | String | Required | `openai` \| `anthropic` \| … | |
| `model` | String | Required | Model id | |
| `prompt_version` | String | Optional | Template version | |
| `input_tokens` | Integer | Optional | | |
| `output_tokens` | Integer | Optional | | |
| `latency_ms` | Integer | Optional | | |
| `estimated_cost_usd` | Decimal | Optional | | |
| `tools_used` | String[] / JSON | Optional | | |
| `status` | String (enum) | Required | `started` \| `succeeded` \| `failed` | |
| `error_summary` | String | Optional | Non-secret | |
| `created_at` | Timestamp | Required | | |
| `finished_at` | Timestamp | Optional | | |

## Relationships

- Many-to-One: User, Conversation  
- Does **not** own Messages, Tasks, Memories

## Constraints

- Unique `public_id`
- AI service does not write this row directly — API records after/during orchestration

## Index Recommendations

- `(user_id, created_at DESC)`
- `(conversation_id, created_at)`
- `(provider, model, created_at)` for cost rollups

## Soft Delete Policy

**Hard delete / aggregate purge** after retention; not user soft-delete primary. Optional soft delete if product shows history.

## Public Identifier Strategy

`ais_…`

## Future Expansion

- Multi-step agent runs as child **AI Steps**
- Provider swap without changing Conversation/Message

---

# Integration Binding (future placeholder)

## Purpose

Placeholder for attaching **non-calendar capabilities** (Slack, Notion, GitHub later) to a Connected Account without polluting Calendar Source. Phase 1 may omit physical storage; the logical slot prevents “Slack Calendar Source” anti-patterns.

## Ownership

- **Owned by:** Connected Account → User

## Fields (planned)

| Name | Data Type | Required | Description | Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID (v7) | Required | Internal PK | |
| `public_id` | String | Required | `int_…` | |
| `user_id` | UUID | Required | Tenancy | |
| `connected_account_id` | UUID | Required | Parent account | |
| `capability` | String (enum) | Required | `calendar` \| `tasks` \| `messaging` \| `docs` \| … | |
| `status` | String (enum) | Required | `active` \| `disabled` | |
| `provider_resource_id` | String | Optional | Remote resource | |
| `config` | JSON | Optional | Non-secret capability config | |
| `created_at` | Timestamp | Required | | |
| `updated_at` | Timestamp | Required | | |
| `deleted_at` | Timestamp | Optional | Soft | |

## Relationships

- Many-to-One: Connected Account  
- Calendar Sources remain the calendar-specific child; bindings cover other capabilities

## Constraints

- Unique `(connected_account_id, capability)` live rows (or + resource id)

## Index Recommendations

- `(user_id, capability)`
- `(connected_account_id, capability)` unique partial

## Soft Delete Policy

**Soft delete** on capability disconnect.

## Public Identifier Strategy

`int_…`

## Future Expansion

- Slack/Telegram messaging capabilities
- Google Tasks / To Do as `capability=tasks` alongside Donna Task sync links

---

# Relationship Validation

| Check | Result |
| --- | --- |
| Every entity has a clear owner | **Pass** — User is tenancy root; Settings 1:1 User; Sources → Accounts → User; Events → Sources; Messages → Conversations → User; others → User |
| Every child has exactly one parent unless intentionally shared | **Pass** — Reminders may reference Task *and/or* Event (intentional dual association, still single User owner). Notifications/Jobs reference many subjects via optional FKs, owned only by User |
| Circular dependencies avoided | **Pass** — no A owns B owns A; citations/audit are references only |
| Aggregates remain independent | **Pass** — Calendar, Task, Conversation, Memory, Notification, Scheduler, Audit, Identity boundaries respected; cross-links are nullable FKs or JSON citations |
| Future providers attach via adapters | **Pass** — `provider` + remote ids + `provider_payload` JSON; no `google_events` canonical table; AI Session/Integration Binding reserved without forcing redesign |
| Polymorphism minimized | **Pass** — avoided on core product FKs; allowed on Audit `subject_type/subject_id` and Message `citations` JSON by design |
| Donna owns business objects | **Pass** — Events/Tasks/Messages/Memories are Donna entities; providers sync |

---

# Database Design Principles

1. **Donna owns the business objects.** Canonical meaning lives in Donna entities (Event, Task, Message, Memory, …).
2. **External providers never own Donna’s data.** Providers contribute sync links and opaque payloads only.
3. **Every aggregate has a single root.** Consistency boundaries follow [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) (User/Identity, Calendar Source→Event, Conversation→Message, etc.).
4. **UUIDv7 is the primary identifier.** Internal `id` is UUIDv7 for all durable entities.
5. **Public IDs are stable API identifiers.** Prefixed `usr_`, `evt_`, … for URLs, tools, and support; never reuse.
6. **JSON fields are only for dynamic provider metadata** (and sparse prefs / citations / payloads) — not for title, times, status, or ownership keys.
7. **Core business data always has explicit fields.** Queryable meaning is columnar/logical fields first.
8. **Avoid polymorphic relationships unless absolutely necessary.** Prefer explicit nullable FKs; exceptions: Audit subjects, message citations bag.
9. **Design for extensibility without premature abstraction.** Placeholders (AI Session, Integration Binding) exist as logical slots — implement when needed.
10. **Optimize for readability over cleverness.** Clear names, obvious ownership, boring structures that last a decade.

---

# Migration Readiness Checklist

| Item | Status |
| --- | --- |
| Every required entity is defined | ✓ User, Connected Account, Calendar Source, Calendar Event, Task, Reminder, Conversation, Message, Memory, Notification, Scheduler Job, Settings, Audit Log, AI Session (placeholder), Integration Binding (placeholder) |
| Every relationship is documented | ✓ Per-entity relationship tables + validation section |
| Every ownership rule is documented | ✓ Ownership sections + aggregate alignment |
| Constraints are identified | ✓ Uniqueness, enums, 1:1 settings, provider upsert keys |
| Index recommendations exist | ✓ Query-pattern based, non-SQL |
| Delete policies exist | ✓ Soft / hard / never per entity |
| Public IDs are defined | ✓ Prefix registry + per-entity strategy |
| Future expansion is documented | ✓ Providers, channels, workspaces, billing hooks |
| No SQL / migrations generated in this milestone | ✓ Documentation only |

**Next step (later milestone):** physical schema + golang-migrate files mapped 1:1 from this document under [DATABASE.md](./DATABASE.md) standards — still without leaking provider ownership into Donna’s core tables.
