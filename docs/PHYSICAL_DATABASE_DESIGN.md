# Donna Physical Database Design

**Status:** PostgreSQL-ready physical specification (pre-migration)  
**Milestone:** M2.4 — Physical Design (documentation only)  
**Non-goals:** `CREATE TABLE`, migration files, Go structs, repositories  

**Inputs:** [DATA_MODEL.md](./DATA_MODEL.md), [DATABASE.md](./DATABASE.md), [SCHEMA_DECISIONS.md](./SCHEMA_DECISIONS.md), [DOMAIN_MODEL.md](./DOMAIN_MODEL.md)

This document **locks architectural choices** from the schema review so an engineer can implement migrations without inventing policy. Where [SCHEMA_DECISIONS.md](./SCHEMA_DECISIONS.md) findings required changes, those decisions are **baked in here** (Auth Identity, notification delivery vs read, Reminder/Job schedule ownership, `text[]` scopes, no redundant `calendar_events.provider`).

---

## Global conventions (locked)

| Topic | Decision |
| --- | --- |
| RDBMS | PostgreSQL 16+ |
| PK type | `uuid` — **UUIDv7**, generated in the **API application** before INSERT (not DB random v4) |
| Public id | `text`, unique, format `{prefix}{crockford_base32_of_uuid_bytes}` or `{prefix}{uuid_hex_no_dashes}` — **application generates** at create; never update |
| Timestamps | `timestamptz` for all instants; store UTC |
| Strings | `text` (not `varchar(n)` unless a hard protocol limit exists) |
| Booleans | `boolean`, NEVER null for flags — use `NOT NULL DEFAULT …` |
| Money/cost | `numeric(12,6)` when needed |
| Status / enums | `text` + `CHECK` (not PostgreSQL `ENUM` types) for easier evolution |
| Soft delete | `deleted_at timestamptz NULL`; live rows = `deleted_at IS NULL` |
| Soft-delete uniques | **Partial unique indexes** `WHERE deleted_at IS NULL` |
| JSON | `jsonb` only where listed below |
| Arrays | `text[]` for OAuth scopes / tool name lists |
| Vectors | Defer `vector` column to memory search milestone; nullable placeholder allowed as comment in migration later |
| FK updates | `ON UPDATE CASCADE` unused — PKs immutable; use `ON UPDATE NO ACTION` |
| Hard DELETE | Rare; prefer soft delete. FKs use `RESTRICT` / `NO ACTION` unless noted. App performs cascading soft-deletes. |
| Schema name | `public` (default) |
| Extensions | `pgcrypto` already enabled (M1); UUIDv7 remains app-side |

**Public id prefixes (locked)**

| Prefix | Table |
| --- | --- |
| `usr_` | `users` |
| `aid_` | `auth_identities` |
| `acct_` | `connected_accounts` |
| `cal_` | `calendar_sources` |
| `evt_` | `calendar_events` |
| `tsk_` | `tasks` |
| `rem_` | `reminders` |
| `conv_` | `conversations` |
| `msg_` | `messages` |
| `mem_` | `memories` |
| `ntf_` | `notifications` |
| `job_` | `scheduler_jobs` |
| `set_` | `user_settings` |
| `audit_` | `audit_logs` |
| `ais_` | `ai_sessions` (wave 2) |
| `int_` | `integration_bindings` (wave 2) |

**Migration waves**

| Wave | Tables |
| --- | --- |
| **Wave 1** (first domain schema) | users, auth_identities, user_settings, connected_accounts, calendar_sources, calendar_events, tasks, reminders, conversations, messages, memories, notifications, scheduler_jobs, audit_logs |
| **Wave 2** | ai_sessions, integration_bindings |
| **Wave 3** (accountability) | goals, daily_plans, daily_reviews, check_ins, notes — design deferred; do not misuse `memories` |

---

# users

## PostgreSQL Table Name

`users`

## Purpose

Donna account / tenancy root.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — (app UUIDv7) | Internal PK |
| `public_id` | `text` | NO | — (app) | `usr_…` |
| `email` | `text` | NO | — | Primary email, lowercased by app |
| `email_verified` | `boolean` | NO | `false` | IdP verified flag |
| `display_name` | `text` | YES | `NULL` | Display name |
| `avatar_url` | `text` | YES | `NULL` | Avatar URL |
| `timezone` | `text` | NO | `'UTC'` | IANA TZ |
| `locale` | `text` | YES | `NULL` | BCP 47 |
| `status` | `text` | NO | `'active'` | Lifecycle status |
| `last_login_at` | `timestamptz` | YES | `NULL` | Last login |
| `created_at` | `timestamptz` | NO | `now()` | Created |
| `updated_at` | `timestamptz` | NO | `now()` | Updated (app maintains) |
| `deleted_at` | `timestamptz` | YES | `NULL` | Soft delete |

## Primary Key

`id` (UUIDv7).

## Public ID

- Column: `public_id`
- Prefix: `usr_`
- Uniqueness: global unique
- Generation: application at insert; immutable

## Foreign Keys

None (root).

## Unique Constraints

- `public_id`
- Partial unique `(email) WHERE deleted_at IS NULL`

## Check Constraints

- `status IN ('active', 'disabled', 'pending_deletion')`
- `email <> ''`
- `public_id LIKE 'usr_%'`
- `timezone <> ''`

## Recommended Indexes

| Index | Why |
| --- | --- |
| PK `id` | Joins |
| Unique `public_id` | API resolve |
| Unique partial `email` live | Login / signup |
| `(status) WHERE deleted_at IS NULL` | Cleanup jobs |

## JSONB Fields

None.

## Lifecycle

`active` ↔ `disabled` → `pending_deletion` → soft-deleted (`deleted_at`) → hard purge job.

---

# auth_identities

## PostgreSQL Table Name

`auth_identities`

## Purpose

Login IdP binding (R-01). Separates **sign-in** Google from **Connected Account** calendar Google.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK UUIDv7 |
| `public_id` | `text` | NO | — | `aid_…` |
| `user_id` | `uuid` | NO | — | Owning user |
| `provider` | `text` | NO | — | `google`, `apple`, `microsoft`, … |
| `provider_subject` | `text` | NO | — | IdP `sub` |
| `email` | `text` | YES | `NULL` | Email at link time |
| `email_verified` | `boolean` | NO | `false` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | Soft unlink |

## Primary Key

`id` (UUIDv7).

## Public ID

- `public_id`, prefix `aid_`, unique, app-generated, immutable

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Soft-delete user first; never orphan identity via hard delete |

## Unique Constraints

- `public_id`
- Partial unique `(provider, provider_subject) WHERE deleted_at IS NULL`

## Check Constraints

- `provider <> ''`
- `provider_subject <> ''`
- `public_id LIKE 'aid_%'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id) WHERE deleted_at IS NULL` | List identities |
| Unique `(provider, provider_subject)` live | OAuth callback resolve |
| Unique `public_id` | |

## JSONB Fields

None.

## Lifecycle

Created on first OAuth → active → soft-deleted on unlink → purge with user.

---

# user_settings

## PostgreSQL Table Name

`user_settings`

## Purpose

1:1 preferences and default calendar routing.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `set_…` (kept for consistency R-09) |
| `user_id` | `uuid` | NO | — | Owner (1:1) |
| `default_personal_calendar_source_id` | `uuid` | YES | `NULL` | Default personal cal |
| `default_work_calendar_source_id` | `uuid` | YES | `NULL` | Default work cal |
| `default_reminder_calendar_source_id` | `uuid` | YES | `NULL` | Default reminder cal |
| `morning_briefing_time` | `time` | YES | `NULL` | Local time-of-day |
| `midday_checkin_time` | `time` | YES | `NULL` | |
| `evening_reflection_time` | `time` | YES | `NULL` | |
| `quiet_hours_start` | `time` | YES | `NULL` | |
| `quiet_hours_end` | `time` | YES | `NULL` | |
| `notifications_enabled` | `boolean` | NO | `true` | Master toggle |
| `preferences` | `jsonb` | NO | `'{}'` | Sparse non-core prefs |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |

No `deleted_at` — row lifetime follows user (app deletes/anonymizes with user).

## Primary Key

`id` (UUIDv7).

## Public ID

`set_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `CASCADE` | `NO ACTION` | Settings cannot exist without user; hard-delete user removes settings |
| `default_personal_calendar_source_id` | `calendar_sources.id` | `SET NULL` | `NO ACTION` | Source may be removed |
| `default_work_calendar_source_id` | `calendar_sources.id` | `SET NULL` | `NO ACTION` | |
| `default_reminder_calendar_source_id` | `calendar_sources.id` | `SET NULL` | `NO ACTION` | |

App must enforce default sources belong to same `user_id`.

## Unique Constraints

- `public_id`
- `user_id` (exactly one settings row)

## Check Constraints

- `public_id LIKE 'set_%'`
- `jsonb_typeof(preferences) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| Unique `user_id` | 1:1 fetch |
| Unique `public_id` | |

## JSONB Fields

**`preferences`:** allow-listed keys only (app validation). Examples: `{ "compact_dashboard": true }`. No secrets. No GIN unless needed.

## Lifecycle

Created with user → updated → removed with user anonymization/hard purge.

---

# connected_accounts

## PostgreSQL Table Name

`connected_accounts`

## Purpose

Integration OAuth accounts (calendar etc.), not login identity.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `acct_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `provider` | `text` | NO | — | `google`, `microsoft`, `apple`, … |
| `provider_account_id` | `text` | NO | — | Remote account id |
| `display_name` | `text` | YES | `NULL` | Label |
| `status` | `text` | NO | `'active'` | Connection status |
| `scopes` | `text[]` | NO | `'{}'` | OAuth scopes (R-10) |
| `credentials_ref` | `text` | NO | — | Vault/secret reference |
| `token_expires_at` | `timestamptz` | YES | `NULL` | |
| `last_synced_at` | `timestamptz` | YES | `NULL` | |
| `provider_metadata` | `jsonb` | NO | `'{}'` | Non-secret quirks |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | Soft disconnect |

## Primary Key

`id` (UUIDv7).

## Public ID

`acct_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Soft-delete account/user in app first |

## Unique Constraints

- `public_id`
- Partial unique `(provider, provider_account_id) WHERE deleted_at IS NULL`

## Check Constraints

- `status IN ('active', 'needs_reauth', 'revoked', 'disconnected')`
- `public_id LIKE 'acct_%'`
- `provider <> ''`
- `credentials_ref <> ''`
- `jsonb_typeof(provider_metadata) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, status) WHERE deleted_at IS NULL` | Connections UI |
| Unique `(provider, provider_account_id)` live | Prevent duplicate links |
| Unique `public_id` | |

## JSONB Fields

**`provider_metadata`:** opaque non-secret provider extras; size-capped in app; no secrets; no GIN v1.

## Lifecycle

`active` → `needs_reauth` → `revoked` / `disconnected` → soft-deleted → purge.

---

# calendar_sources

## PostgreSQL Table Name

`calendar_sources`

## Purpose

Synced calendar feed under a connected account.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `cal_…` |
| `user_id` | `uuid` | NO | — | Denormalized tenancy |
| `connected_account_id` | `uuid` | NO | — | Parent account |
| `provider_calendar_id` | `text` | NO | — | Remote calendar id |
| `name` | `text` | NO | — | Display name |
| `color` | `text` | YES | `NULL` | UI color |
| `is_primary_on_provider` | `boolean` | NO | `false` | Provider primary flag |
| `sync_enabled` | `boolean` | NO | `true` | Donna sync toggle |
| `sync_cursor` | `text` | YES | `NULL` | Incremental sync token |
| `last_synced_at` | `timestamptz` | YES | `NULL` | |
| `timezone` | `text` | YES | `NULL` | Source TZ hint |
| `provider_metadata` | `jsonb` | NO | `'{}'` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`cal_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Tenancy |
| `connected_account_id` | `connected_accounts.id` | `RESTRICT` | `NO ACTION` | App soft-cascades |

App invariant: `user_id` must equal parent account’s `user_id`.

## Unique Constraints

- `public_id`
- Partial unique `(connected_account_id, provider_calendar_id) WHERE deleted_at IS NULL`

## Check Constraints

- `name <> ''`
- `public_id LIKE 'cal_%'`
- `jsonb_typeof(provider_metadata) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, sync_enabled) WHERE deleted_at IS NULL` | Sync selection |
| `(connected_account_id) WHERE deleted_at IS NULL` | List sources |
| Unique pair live | Sync upsert parent |

## JSONB Fields

**`provider_metadata`:** same policy as accounts.

## Lifecycle

Created → sync_enabled toggled → soft-deleted on unsubscribe/disconnect.

---

# calendar_events

## PostgreSQL Table Name

`calendar_events`

## Purpose

Donna canonical calendar event (no `provider` column — derive via source → account; R-07).

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `evt_…` |
| `user_id` | `uuid` | NO | — | Tenancy |
| `calendar_source_id` | `uuid` | NO | — | Parent source |
| `title` | `text` | NO | — | Title |
| `description` | `text` | YES | `NULL` | Body |
| `location` | `text` | YES | `NULL` | Location |
| `starts_at` | `timestamptz` | NO | — | Start |
| `ends_at` | `timestamptz` | NO | — | End |
| `is_all_day` | `boolean` | NO | `false` | All-day |
| `status` | `text` | NO | `'confirmed'` | Event status |
| `visibility` | `text` | YES | `NULL` | Visibility |
| `attendees_summary` | `jsonb` | NO | `'[]'` | Thin attendee list |
| `recurrence_rule` | `text` | YES | `NULL` | RRULE / summary |
| `recurring_event_id` | `uuid` | YES | `NULL` | Series parent |
| `provider_event_id` | `text` | YES | `NULL` | Remote id if synced |
| `provider_etag` | `text` | YES | `NULL` | Sync version |
| `provider_payload` | `jsonb` | YES | `NULL` | Opaque snapshot |
| `origin` | `text` | NO | `'donna'` | `donna` \| `provider_sync` |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | Tombstone |

## Primary Key

`id` (UUIDv7).

## Public ID

`evt_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Tenancy |
| `calendar_source_id` | `calendar_sources.id` | `RESTRICT` | `NO ACTION` | App soft-cascades |
| `recurring_event_id` | `calendar_events.id` | `SET NULL` | `NO ACTION` | Series parent optional |

## Unique Constraints

- `public_id`
- Partial unique `(calendar_source_id, provider_event_id) WHERE deleted_at IS NULL AND provider_event_id IS NOT NULL`

## Check Constraints

- `ends_at >= starts_at`
- `status IN ('confirmed', 'tentative', 'cancelled')`
- `visibility IS NULL OR visibility IN ('default', 'private', 'public')`
- `origin IN ('donna', 'provider_sync')`
- `title <> ''`
- `public_id LIKE 'evt_%'`
- `jsonb_typeof(attendees_summary) = 'array'`
- `provider_payload IS NULL OR jsonb_typeof(provider_payload) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, starts_at) WHERE deleted_at IS NULL` | Dashboard upcoming |
| `(calendar_source_id, starts_at, ends_at) WHERE deleted_at IS NULL` | Source range / sync |
| Unique sync key live | Upsert |
| Unique `public_id` | API |
| `(user_id, status) WHERE deleted_at IS NULL` | Filter cancelled |

## JSONB Fields

**`attendees_summary`:** array of `{ "email"?: string, "name"?: string, "status"?: string }`, max N (e.g. 50) enforced in app.  

**`provider_payload`:** opaque object; size cap; never secrets; never canonical.

## Lifecycle

Created → updated → `cancelled` and/or soft-deleted → purge.

---

# tasks

## PostgreSQL Table Name

`tasks`

## Purpose

Actionable work items (status vocabulary locked R-08).

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `tsk_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `title` | `text` | NO | — | |
| `description` | `text` | YES | `NULL` | |
| `status` | `text` | NO | `'open'` | `open` \| `completed` \| `cancelled` |
| `priority` | `text` | YES | `NULL` | `low` \| `medium` \| `high` |
| `due_at` | `timestamptz` | YES | `NULL` | |
| `completed_at` | `timestamptz` | YES | `NULL` | |
| `is_backlog` | `boolean` | NO | `false` | |
| `recurrence_rule` | `text` | YES | `NULL` | |
| `provider` | `text` | YES | `NULL` | Future Tasks/To Do |
| `provider_task_id` | `text` | YES | `NULL` | |
| `provider_payload` | `jsonb` | YES | `NULL` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`tsk_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Soft-delete cascade in app |

## Unique Constraints

- `public_id`
- Partial unique `(user_id, provider, provider_task_id) WHERE deleted_at IS NULL AND provider_task_id IS NOT NULL`

## Check Constraints

- `status IN ('open', 'completed', 'cancelled')`
- `priority IS NULL OR priority IN ('low', 'medium', 'high')`
- `title <> ''`
- `public_id LIKE 'tsk_%'`
- `(status = 'completed' AND completed_at IS NOT NULL) OR (status <> 'completed')` — completed requires timestamp
- `provider_payload IS NULL OR jsonb_typeof(provider_payload) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, status, due_at) WHERE deleted_at IS NULL` | Today’s / upcoming tasks |
| `(user_id, is_backlog, status) WHERE deleted_at IS NULL` | Backlog widget |
| Unique `public_id` | Tools/API |
| Provider unique live | Sync |

## JSONB Fields

**`provider_payload`:** opaque sync only.

## Lifecycle

`open` → `completed` / `cancelled` → soft-deleted → purge. Reopen: `completed` → `open` clears `completed_at`.

---

# reminders

## PostgreSQL Table Name

`reminders`

## Purpose

Fire-at-time nudges. **`remind_at` is source of truth** for schedule (R-03). Scheduler jobs copy `run_at` from `remind_at`.

**Write ownership (R-02):** Reminder service owns writes. Task soft-delete → app soft-deletes child reminders. Event soft-delete → app sets `calendar_event_id` null **or** soft-deletes reminder if event-only; prefer **SET NULL** on hard path and soft-delete reminders when event soft-deleted if reminder has no task.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `rem_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `task_id` | `uuid` | YES | `NULL` | Optional task |
| `calendar_event_id` | `uuid` | YES | `NULL` | Optional event |
| `title` | `text` | NO | — | Reminder text |
| `remind_at` | `timestamptz` | NO | — | Fire time (SoT) |
| `status` | `text` | NO | `'scheduled'` | |
| `channel_preference` | `text` | YES | `NULL` | Override channel |
| `last_error` | `text` | YES | `NULL` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`rem_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Tenancy |
| `task_id` | `tasks.id` | `RESTRICT` | `NO ACTION` | App soft-cascades reminders when task soft-deleted |
| `calendar_event_id` | `calendar_events.id` | `SET NULL` | `NO ACTION` | Event removal should not destroy standalone/task reminders |

## Unique Constraints

- `public_id`

## Check Constraints

- `status IN ('scheduled', 'sent', 'cancelled', 'failed')`
- `title <> ''`
- `public_id LIKE 'rem_%'`
- At least standalone title always required; `task_id`/`calendar_event_id` both null allowed

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(status, remind_at) WHERE deleted_at IS NULL` | Due processing |
| `(user_id, status, remind_at) WHERE deleted_at IS NULL` | User views |
| `(task_id) WHERE deleted_at IS NULL` | Task cascade lookup |
| `(calendar_event_id) WHERE deleted_at IS NULL` | Event association |
| Unique `public_id` | |

## JSONB Fields

None.

## Lifecycle

`scheduled` → `sent` \| `failed` \| `cancelled`; failed may return to `scheduled` on reschedule; soft-delete.

---

# conversations

## PostgreSQL Table Name

`conversations`

## Purpose

Chat threads with Donna.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `conv_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `title` | `text` | YES | `NULL` | |
| `purpose` | `text` | YES | `NULL` | morning/midday/… |
| `status` | `text` | NO | `'active'` | |
| `unread_count` | `integer` | NO | `0` | Denormalized |
| `last_message_at` | `timestamptz` | YES | `NULL` | List sort |
| `channel` | `text` | NO | `'web'` | Surface |
| `channel_thread_id` | `text` | YES | `NULL` | External thread |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`conv_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Soft cascade in app |

## Unique Constraints

- `public_id`

## Check Constraints

- `status IN ('active', 'archived')`
- `channel IN ('web', 'telegram', 'whatsapp')` — extend as needed; Phase 1 uses `web`
- `purpose IS NULL OR purpose IN ('general', 'morning', 'midday', 'evening', 'system')`
- `unread_count >= 0`
- `public_id LIKE 'conv_%'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, last_message_at DESC NULLS LAST) WHERE deleted_at IS NULL` | Conversation list |
| `(user_id, status) WHERE deleted_at IS NULL` | Active filter |
| Unique `public_id` | |

## JSONB Fields

None.

## Lifecycle

`active` → `archived` → soft-deleted.

---

# messages

## PostgreSQL Table Name

`messages`

## Purpose

Conversation turns. Citations remain `jsonb` for v1 (R-05 accepted).

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `msg_…` |
| `user_id` | `uuid` | NO | — | Tenancy |
| `conversation_id` | `uuid` | NO | — | Parent |
| `role` | `text` | NO | — | user/assistant/system |
| `content` | `text` | NO | — | Body |
| `content_format` | `text` | NO | `'plain'` | plain/markdown |
| `client_message_id` | `text` | YES | `NULL` | Idempotency |
| `citations` | `jsonb` | NO | `'{}'` | Soft refs |
| `ai_session_id` | `uuid` | YES | `NULL` | Optional link wave 2 |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`msg_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | Tenancy |
| `conversation_id` | `conversations.id` | `RESTRICT` | `NO ACTION` | App soft-cascades messages |
| `ai_session_id` | `ai_sessions.id` | `SET NULL` | `NO ACTION` | Wave 2; omit FK until table exists |

**Wave 1 note:** omit `ai_session_id` column **or** include nullable without FK until wave 2 — **prefer include nullable column without FK in wave 1**, add FK in wave 2.

## Unique Constraints

- `public_id`
- Partial unique `(conversation_id, client_message_id) WHERE deleted_at IS NULL AND client_message_id IS NOT NULL`

## Check Constraints

- `role IN ('user', 'assistant', 'system')`
- `content_format IN ('plain', 'markdown')`
- `content <> ''`
- `public_id LIKE 'msg_%'`
- `jsonb_typeof(citations) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(conversation_id, created_at) WHERE deleted_at IS NULL` | History |
| `(user_id, created_at DESC) WHERE deleted_at IS NULL` | Cross-thread rare |
| Unique `public_id` | |
| Unique client id live | Idempotent send |

## JSONB Fields

**`citations`:** `{ "task_ids": ["uuid"|public], "event_ids": [], "memory_ids": [] }` — prefer internal UUIDs; app validates. Future: `message_citations` junction.

## Lifecycle

Created (immutable content preferred) → soft-deleted. Edits rare; if allowed, bump `updated_at`.

---

# memories

## PostgreSQL Table Name

`memories`

## Purpose

Durable recalled knowledge.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `mem_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `content` | `text` | NO | — | Fact text |
| `category` | `text` | YES | `NULL` | Category |
| `importance` | `integer` | YES | `NULL` | Rank 1–100 optional |
| `source` | `text` | NO | — | Provenance kind |
| `source_conversation_id` | `uuid` | YES | `NULL` | |
| `source_message_id` | `uuid` | YES | `NULL` | |
| `embedding_model` | `text` | YES | `NULL` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

**Embedding:** add `embedding vector(...)` in a later migration when pgvector search ships — not required in wave 1.

## Primary Key

`id` (UUIDv7).

## Public ID

`mem_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | |
| `source_conversation_id` | `conversations.id` | `SET NULL` | `NO ACTION` | Provenance optional |
| `source_message_id` | `messages.id` | `SET NULL` | `NO ACTION` | |

## Unique Constraints

- `public_id`

## Check Constraints

- `source IN ('explicit', 'chat_extract', 'review', 'system')`
- `category IS NULL OR category IN ('preference', 'person', 'project', 'commitment', 'idea', 'other')`
- `importance IS NULL OR (importance >= 1 AND importance <= 100)`
- `content <> ''`
- `public_id LIKE 'mem_%'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, updated_at DESC) WHERE deleted_at IS NULL` | Recent memories |
| `(user_id, category) WHERE deleted_at IS NULL` | Filter |
| Unique `public_id` | Tools |
| Later: vector index | Semantic search |

## JSONB Fields

None in wave 1.

## Lifecycle

Created → updated → soft-deleted → purge.

---

# notifications

## PostgreSQL Table Name

`notifications`

## Purpose

User-visible alerts. **R-04 locked:** separate delivery from read/dismiss.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `ntf_…` |
| `user_id` | `uuid` | NO | — | Owner |
| `channel` | `text` | NO | `'browser_push'` | Delivery channel |
| `title` | `text` | NO | — | |
| `body` | `text` | NO | — | |
| `priority` | `text` | NO | `'normal'` | |
| `delivery_status` | `text` | NO | `'pending'` | pending/sent/failed |
| `reminder_id` | `uuid` | YES | `NULL` | Correlation |
| `calendar_event_id` | `uuid` | YES | `NULL` | |
| `scheduler_job_id` | `uuid` | YES | `NULL` | |
| `conversation_id` | `uuid` | YES | `NULL` | |
| `payload` | `jsonb` | NO | `'{}'` | Channel extras |
| `sent_at` | `timestamptz` | YES | `NULL` | |
| `read_at` | `timestamptz` | YES | `NULL` | User read |
| `dismissed_at` | `timestamptz` | YES | `NULL` | User dismiss |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Primary Key

`id` (UUIDv7).

## Public ID

`ntf_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | |
| `reminder_id` | `reminders.id` | `SET NULL` | `NO ACTION` | Keep notification history |
| `calendar_event_id` | `calendar_events.id` | `SET NULL` | `NO ACTION` | |
| `scheduler_job_id` | `scheduler_jobs.id` | `SET NULL` | `NO ACTION` | |
| `conversation_id` | `conversations.id` | `SET NULL` | `NO ACTION` | |

## Unique Constraints

- `public_id`

## Check Constraints

- `channel IN ('browser_push', 'email', 'telegram')` — extend later
- `priority IN ('low', 'normal', 'high')`
- `delivery_status IN ('pending', 'sent', 'failed')`
- `title <> ''`
- `body <> ''`
- `public_id LIKE 'ntf_%'`
- `jsonb_typeof(payload) = 'object'`
- `(delivery_status <> 'sent') OR (sent_at IS NOT NULL)`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(user_id, created_at DESC) WHERE deleted_at IS NULL` | Notification center |
| `(user_id, created_at DESC) WHERE deleted_at IS NULL AND read_at IS NULL` | Unread |
| `(user_id, delivery_status) WHERE deleted_at IS NULL` | Ops |
| `(reminder_id)` | Correlation |
| Unique `public_id` | |

## JSONB Fields

**`payload`:** non-secret channel data only; no substitute for FK ids when FK exists.

## Lifecycle

`pending` → `sent`/`failed`; independently `read_at` / `dismissed_at`; soft-delete.

---

# scheduler_jobs

## PostgreSQL Table Name

`scheduler_jobs`

## Purpose

Durable job ledger. For `reminder_fire`, set `run_at = reminders.remind_at` at creation (R-03).

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `job_…` |
| `user_id` | `uuid` | NO | — | Scope |
| `job_type` | `text` | NO | — | Type |
| `status` | `text` | NO | `'pending'` | |
| `run_at` | `timestamptz` | NO | — | Due time |
| `attempt_count` | `integer` | NO | `0` | |
| `max_attempts` | `integer` | NO | `5` | |
| `payload` | `jsonb` | NO | `'{}'` | Extras only |
| `reminder_id` | `uuid` | YES | `NULL` | |
| `connected_account_id` | `uuid` | YES | `NULL` | Sync jobs |
| `last_error` | `text` | YES | `NULL` | |
| `started_at` | `timestamptz` | YES | `NULL` | |
| `finished_at` | `timestamptz` | YES | `NULL` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |

No soft delete — retention hard-delete.

## Primary Key

`id` (UUIDv7).

## Public ID

`job_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | |
| `reminder_id` | `reminders.id` | `SET NULL` | `NO ACTION` | Keep job history |
| `connected_account_id` | `connected_accounts.id` | `SET NULL` | `NO ACTION` | |

## Unique Constraints

- `public_id`
- Optional later: idempotency key in payload + unique expression — app-enforced v1

## Check Constraints

- `job_type IN ('morning_briefing', 'midday_checkin', 'evening_reflection', 'reminder_fire', 'calendar_sync')` — extend carefully
- `status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')`
- `attempt_count >= 0`
- `max_attempts >= 1`
- `public_id LIKE 'job_%'`
- `jsonb_typeof(payload) = 'object'`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(status, run_at)` WHERE status IN pending/running — use partial `(run_at) WHERE status = 'pending'` | Poller |
| `(user_id, job_type, status)` | User/debug |
| `(reminder_id)` | Lookup |
| Unique `public_id` | Observability |

## JSONB Fields

**`payload`:** parameters not covered by FKs; never the only copy of `reminder_id`.

## Lifecycle

`pending` → `running` → `succeeded`/`failed`/`cancelled`; retry increments `attempt_count`; purge succeeded after retention.

---

# audit_logs

## PostgreSQL Table Name

`audit_logs`

## Purpose

Append-only security/compliance log.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `audit_…` |
| `actor_user_id` | `uuid` | YES | `NULL` | Actor (null = system) |
| `action` | `text` | NO | — | Stable action key |
| `subject_type` | `text` | YES | `NULL` | Entity kind |
| `subject_id` | `uuid` | YES | `NULL` | Internal id |
| `subject_public_id` | `text` | YES | `NULL` | Snapshot |
| `request_id` | `text` | YES | `NULL` | Correlation |
| `metadata` | `jsonb` | NO | `'{}'` | Non-secret |
| `created_at` | `timestamptz` | NO | `now()` | Immutable |

No `updated_at`, no `deleted_at`.

## Primary Key

`id` (UUIDv7).

## Public ID

`audit_`, unique, app-generated.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `actor_user_id` | `users.id` | `SET NULL` | `NO ACTION` | Preserve audit if user purged |

No FK on subject (polymorphic by design).

## Unique Constraints

- `public_id`

## Check Constraints

- `action <> ''`
- `public_id LIKE 'audit_%'`
- `jsonb_typeof(metadata) = 'object'`
- `subject_type IS NULL OR subject_type IN ('user', 'auth_identity', 'connected_account', 'calendar_source', 'calendar_event', 'task', 'reminder', 'conversation', 'message', 'memory', 'notification', 'scheduler_job', 'user_settings', 'billing')`

## Recommended Indexes

| Index | Why |
| --- | --- |
| `(actor_user_id, created_at DESC)` | Per-user audit |
| `(subject_type, subject_id, created_at DESC)` | Subject history |
| `(action, created_at DESC)` | Action reports |
| `(request_id) WHERE request_id IS NOT NULL` | Trace join |
| Unique `public_id` | |

## JSONB Fields

**`metadata`:** redacted context only; never tokens.

## Lifecycle

Insert only. Retention purge is ops-only hard delete of old partitions (future).

---

# ai_sessions (wave 2)

## PostgreSQL Table Name

`ai_sessions`

## Purpose

Durable LLM turn usage (optional hardening of observability).

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `ais_…` |
| `user_id` | `uuid` | NO | — | |
| `conversation_id` | `uuid` | YES | `NULL` | |
| `trigger_message_id` | `uuid` | YES | `NULL` | |
| `provider` | `text` | NO | — | |
| `model` | `text` | NO | — | |
| `prompt_version` | `text` | YES | `NULL` | |
| `input_tokens` | `integer` | YES | `NULL` | |
| `output_tokens` | `integer` | YES | `NULL` | |
| `latency_ms` | `integer` | YES | `NULL` | |
| `estimated_cost_usd` | `numeric(12,6)` | YES | `NULL` | |
| `tools_used` | `text[]` | NO | `'{}'` | |
| `status` | `text` | NO | `'started'` | |
| `error_summary` | `text` | YES | `NULL` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `finished_at` | `timestamptz` | YES | `NULL` | |

## Primary Key / Public ID

UUIDv7; `ais_` unique.

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | |
| `conversation_id` | `conversations.id` | `SET NULL` | `NO ACTION` | |
| `trigger_message_id` | `messages.id` | `SET NULL` | `NO ACTION` | |

## Checks / Indexes / Lifecycle

- `status IN ('started','succeeded','failed')`
- Indexes: `(user_id, created_at DESC)`, `(conversation_id, created_at)`
- Retention hard purge; no soft delete required

---

# integration_bindings (wave 2)

## PostgreSQL Table Name

`integration_bindings`

## Purpose

Non-calendar capabilities on a connected account.

## Columns

| Name | PostgreSQL Type | Nullable | Default | Description |
| --- | --- | --- | --- | --- |
| `id` | `uuid` | NO | — | PK |
| `public_id` | `text` | NO | — | `int_…` |
| `user_id` | `uuid` | NO | — | |
| `connected_account_id` | `uuid` | NO | — | |
| `capability` | `text` | NO | — | calendar/tasks/messaging/… |
| `status` | `text` | NO | `'active'` | |
| `provider_resource_id` | `text` | YES | `NULL` | |
| `config` | `jsonb` | NO | `'{}'` | |
| `created_at` | `timestamptz` | NO | `now()` | |
| `updated_at` | `timestamptz` | NO | `now()` | |
| `deleted_at` | `timestamptz` | YES | `NULL` | |

## Foreign Keys

| Column | References | ON DELETE | ON UPDATE | Reasoning |
| --- | --- | --- | --- | --- |
| `user_id` | `users.id` | `RESTRICT` | `NO ACTION` | |
| `connected_account_id` | `connected_accounts.id` | `RESTRICT` | `NO ACTION` | Soft cascade in app |

## Unique

Partial `(connected_account_id, capability) WHERE deleted_at IS NULL` (add resource id if multi-resource).

---

# Foreign Key Matrix

| From | To | ON DELETE | ON UPDATE | Reason |
| --- | --- | --- | --- | --- |
| `auth_identities.user_id` | `users.id` | RESTRICT | NO ACTION | Soft-delete first |
| `user_settings.user_id` | `users.id` | CASCADE | NO ACTION | 1:1 dies with user hard-delete |
| `user_settings.default_*_calendar_source_id` | `calendar_sources.id` | SET NULL | NO ACTION | Optional defaults |
| `connected_accounts.user_id` | `users.id` | RESTRICT | NO ACTION | Soft cascade |
| `calendar_sources.user_id` | `users.id` | RESTRICT | NO ACTION | Tenancy |
| `calendar_sources.connected_account_id` | `connected_accounts.id` | RESTRICT | NO ACTION | Soft cascade |
| `calendar_events.user_id` | `users.id` | RESTRICT | NO ACTION | Tenancy |
| `calendar_events.calendar_source_id` | `calendar_sources.id` | RESTRICT | NO ACTION | Soft cascade |
| `calendar_events.recurring_event_id` | `calendar_events.id` | SET NULL | NO ACTION | Optional series |
| `tasks.user_id` | `users.id` | RESTRICT | NO ACTION | Soft cascade |
| `reminders.user_id` | `users.id` | RESTRICT | NO ACTION | |
| `reminders.task_id` | `tasks.id` | RESTRICT | NO ACTION | App soft-cascades reminders |
| `reminders.calendar_event_id` | `calendar_events.id` | SET NULL | NO ACTION | Keep task reminders |
| `conversations.user_id` | `users.id` | RESTRICT | NO ACTION | Soft cascade |
| `messages.user_id` | `users.id` | RESTRICT | NO ACTION | |
| `messages.conversation_id` | `conversations.id` | RESTRICT | NO ACTION | Soft cascade messages |
| `memories.user_id` | `users.id` | RESTRICT | NO ACTION | |
| `memories.source_conversation_id` | `conversations.id` | SET NULL | NO ACTION | Provenance |
| `memories.source_message_id` | `messages.id` | SET NULL | NO ACTION | Provenance |
| `notifications.user_id` | `users.id` | RESTRICT | NO ACTION | |
| `notifications.reminder_id` | `reminders.id` | SET NULL | NO ACTION | History |
| `notifications.calendar_event_id` | `calendar_events.id` | SET NULL | NO ACTION | |
| `notifications.scheduler_job_id` | `scheduler_jobs.id` | SET NULL | NO ACTION | |
| `notifications.conversation_id` | `conversations.id` | SET NULL | NO ACTION | |
| `scheduler_jobs.user_id` | `users.id` | RESTRICT | NO ACTION | |
| `scheduler_jobs.reminder_id` | `reminders.id` | SET NULL | NO ACTION | |
| `scheduler_jobs.connected_account_id` | `connected_accounts.id` | SET NULL | NO ACTION | |
| `audit_logs.actor_user_id` | `users.id` | SET NULL | NO ACTION | Keep audit |
| `ai_sessions.user_id` | `users.id` | RESTRICT | NO ACTION | Wave 2 |
| `ai_sessions.conversation_id` | `conversations.id` | SET NULL | NO ACTION | Wave 2 |
| `ai_sessions.trigger_message_id` | `messages.id` | SET NULL | NO ACTION | Wave 2 |
| `integration_bindings.user_id` | `users.id` | RESTRICT | NO ACTION | Wave 2 |
| `integration_bindings.connected_account_id` | `connected_accounts.id` | RESTRICT | NO ACTION | Wave 2 |

**Application soft-delete cascades (not DB CASCADE):** User → accounts → sources → events; User → tasks → reminders; User → conversations → messages; Account disconnect → sources/events soft-delete policy.

---

# Index Matrix (query → index)

| Query / Feature | Entity | Columns | Recommended Index | Reason |
| --- | --- | --- | --- | --- |
| Sign-in resolve IdP | auth_identities | provider, provider_subject | Unique partial live | OAuth callback |
| Sign-in / signup email | users | email | Unique partial live | Lookup |
| API public id resolve | * | public_id | Unique per table | Edge resolve |
| Dashboard today’s tasks | tasks | user_id, status, due_at | Composite partial live | Due today / open |
| Backlog widget | tasks | user_id, is_backlog, status | Composite partial | Backlog |
| Upcoming events | calendar_events | user_id, starts_at | Composite partial | Dashboard |
| Calendar month grid | calendar_events | calendar_source_id, starts_at, ends_at | Composite partial | Range |
| Calendar sync upsert | calendar_events | calendar_source_id, provider_event_id | Unique partial | Idempotent sync |
| Connection list | connected_accounts | user_id, status | Composite partial | Settings |
| Conversation list | conversations | user_id, last_message_at | Composite partial DESC | Phone list |
| Conversation history | messages | conversation_id, created_at | Composite partial | Scrollback |
| Idempotent message send | messages | conversation_id, client_message_id | Unique partial | Dedupe |
| Notification center | notifications | user_id, created_at | Composite partial | Inbox |
| Unread notifications | notifications | user_id, created_at WHERE read_at IS NULL | Partial | Badge |
| Reminder processing | reminders | status, remind_at | Composite / partial scheduled | Worker |
| Scheduler poll | scheduler_jobs | run_at WHERE status='pending' | Partial | Due jobs |
| Memory browse | memories | user_id, updated_at | Composite partial | UI |
| Memory semantic (later) | memories | embedding | Vector index | Search |
| Audit by user | audit_logs | actor_user_id, created_at | Composite | Security |
| Audit by subject | audit_logs | subject_type, subject_id, created_at | Composite | Forensics |
| Job by reminder | scheduler_jobs | reminder_id | Single | Debug |

---

# PostgreSQL Best Practices (Donna)

| Topic | Recommendation |
| --- | --- |
| **UUIDv7** | Generate in Go before INSERT; do not use `gen_random_uuid()` (v4) for PKs |
| **timestamptz** | All absolute times; interpret wall-clock prefs with `users.timezone` |
| **jsonb** | Objects/arrays only as specified; validate `jsonb_typeof`; size-cap in app; no secrets |
| **text vs varchar** | Prefer `text`; avoid arbitrary `varchar(255)` |
| **enum vs text+check** | Use `text` + `CHECK` for evolvable statuses; avoid PG ENUM alter pain |
| **soft delete** | `deleted_at`; all list queries filter null; partial uniques |
| **partial indexes** | Default for live-row access paths |
| **covering indexes** | Defer until EXPLAIN proves need (INCLUDE columns) |
| **generated columns** | Not required v1; consider later for `starts_on date` from `starts_at` if heavy |
| **updated_at** | App sets on write; optional trigger later — pick one and stay consistent (**app-owned**) |
| **connection pooling** | pgx pool already; short transactions; no remote calls inside TX |
| **migrations** | Expand/contract; never edit applied files |

---

# Implementation Readiness Checklist

| Item | Status |
| --- | --- |
| Every wave-1 column defined with PG type | ✓ |
| Every FK defined with ON DELETE/UPDATE | ✓ |
| Unique + check constraints defined | ✓ |
| Indexes mapped to queries | ✓ |
| Delete / soft-delete policies documented | ✓ |
| Public IDs + prefixes locked | ✓ |
| PostgreSQL types selected | ✓ |
| Schema review findings R-01–R-04, R-07, R-10 applied | ✓ |
| R-05 citations JSON accepted for v1 | ✓ |
| R-06 Planning/Notes deferred to wave 3 (explicit) | ✓ |
| Wave 2 AI/Integration specified enough to add later | ✓ |
| **Ready for migration generation** | ✓ **Yes — for wave 1 tables** |

### Explicit non-goals for first SQL PR

- No `CREATE TABLE` in this milestone (next milestone generates migrations from **this** doc)
- No Planning/Notes tables yet
- No pgvector column until memory search milestone
- No Go models until after migrations exist (or alongside — product choice)

**Engineer instruction:** Implement wave 1 migrations as a 1:1 translation of this document. Do not reopen aggregate or provider-ownership debates without updating [SCHEMA_DECISIONS.md](./SCHEMA_DECISIONS.md) and this file first.
