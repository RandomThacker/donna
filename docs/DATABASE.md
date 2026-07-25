# Donna Database Standards

**Status:** Source of truth for persistence conventions  
**Companion:** [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) (entities & relationships — no SQL)  

This document defines **how** Donna will persist data when schema milestones begin. It does **not** define tables, columns, or migrations. Field-level logical design: [DATA_MODEL.md](./DATA_MODEL.md).

Related: [SYSTEM_DESIGN.md](./SYSTEM_DESIGN.md), [OBSERVABILITY.md](./OBSERVABILITY.md), [ARCHITECTURE.md](./ARCHITECTURE.md), [CURSOR_RULES.md](./CURSOR_RULES.md).

---

## Principles

1. **PostgreSQL is the source of truth.** The AI service never writes to the database.
2. **Donna domain first.** Tables model [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) aggregates — not Google/OpenAI API resources.
3. **Provider data is linkage + sync metadata**, not the canonical row shape for events/tasks/messages.
4. **Migrations are the only schema change mechanism** (golang-migrate already wired in M1).
5. **Optimize for 5–10 years of evolution:** clear naming, UUIDs, soft deletes where product needs recovery, explicit audit.

---

## UUID policy

| Rule | Detail |
| --- | --- |
| Primary keys | UUID for all domain entities |
| Generation | **UUIDv7** (time-sortable) generated in application or DB — see [DATA_MODEL.md](./DATA_MODEL.md); stay consistent once chosen |
| External ids | Provider remote ids are **separate columns** (text), never used as Donna PKs |
| API exposure | Prefer **public ids** (`usr_…`, `evt_…`, …) in APIs/URLs; internal UUIDs for FKs |

---

## Naming conventions

| Concern | Convention |
| --- | --- |
| Tables | `snake_case`, plural nouns: `users`, `calendar_events`, `connected_accounts` |
| Columns | `snake_case` |
| Primary key | `id` (uuid) |
| Foreign keys | `<singular_table>_id` — e.g. `user_id`, `conversation_id` |
| Booleans | `is_` / `has_` prefix when clarity helps: `is_completed`, `is_dismissed` |
| Timestamps | `created_at`, `updated_at`, `deleted_at`, `scheduled_at`, `sent_at` |
| Enums | Prefer PostgreSQL enums **or** text + check constraint; type names `snake_case` |
| Join tables | `entity_a_entity_b` when many-to-many is required (avoid until necessary) |
| Provider link columns | `provider`, `provider_account_id`, `provider_object_id`, `provider_etag` / `provider_version` |

Avoid encoding provider names into table names (`google_events`). Use `calendar_events` + provider link fields.

---

## Timestamp conventions

| Column | Meaning |
| --- | --- |
| `created_at` | Row creation (timestamptz, required) |
| `updated_at` | Last mutation (timestamptz, required on mutable entities) |
| `deleted_at` | Soft delete marker (nullable timestamptz) |
| Domain times | Event starts/ends, reminder fire times: **timestamptz** always |

Rules:

- Store absolute instants in UTC (`timestamptz`)
- User-facing timezones come from profile/settings, not from stripping TZ in the DB
- Prefer `now()` at write time via application or DB default — be consistent within a table

---

## Soft delete policy

| Use soft delete when | Prefer hard delete when |
| --- | --- |
| User-visible recovery matters (tasks, memories, conversations) | Pure ephemeral jobs/runs after retention |
| Audit/compliance needs existence history | Deduplicated cache rows |
| Sync must tombstone to providers | Mistaken draft data never shared |

Rules:

- Soft-deleted rows set `deleted_at`; unique indexes must account for soft delete (partial unique indexes where needed)
- Default repository reads **exclude** soft-deleted rows unless explicitly querying archive
- Cascade policy: define per aggregate — prefer soft-delete children with parent or block delete when children exist
- Hard delete of PII follows a future retention/GDPR process — not ad-hoc DELETEs from handlers

---

## Foreign key conventions

| Rule | Detail |
| --- | --- |
| Always declare FKs | For ownership and aggregate integrity |
| ON DELETE | Prefer `RESTRICT` / `NO ACTION` for safety; use `CASCADE` only inside a clear aggregate where child cannot exist alone |
| Cross-aggregate refs | Nullable FKs or association tables; do not cascade-delete across aggregates |
| Tenancy | Child tables include `user_id` (or workspace_id later) even when reachable via parent, when it simplifies RLS and queries — document denormalization consciously |

---

## Indexing guidelines

Index for **actual access paths**, not speculation.

| Pattern | Index |
| --- | --- |
| FK columns used in joins/filters | B-tree on FK |
| Per-user lists | `(user_id, created_at DESC)` or status-specific composites |
| Soft delete | Partial indexes with `WHERE deleted_at IS NULL` |
| Provider sync upsert | Unique `(provider, provider_object_id)` or `(calendar_source_id, provider_object_id)` where applicable |
| Time ranges | Event `(calendar_source_id, starts_at, ends_at)` as needed |
| Scheduler polling | `(status, run_at)` for due jobs |

Avoid:

- Indexing every column
- Redundant indexes that duplicate PK/unique constraints

Review slow queries via observability budgets ([OBSERVABILITY.md](./OBSERVABILITY.md) — DB &lt; 50ms target).

---

## JSONB usage guidelines

JSONB is for **flexible side-data**, not for core relational meaning.

| Appropriate | Inappropriate |
| --- | --- |
| Provider raw payload snapshot (opaque) | Primary title/start/end of an event |
| Sparse metadata / feature flags bag | User id / foreign keys |
| Template params for notifications | Anything you need strong FK integrity on |

Rules:

- Document the JSON shape in code (typed struct) even if the DB is schemaless
- Constrain with CHECK or application validation when keys are required
- Prefer columns when a field is filtered, sorted, or joined often
- Never store secrets in JSONB (tokens belong in encrypted secret storage / sealed columns)

---

## Enum usage guidelines

| Approach | When |
| --- | --- |
| PostgreSQL `ENUM` | Closed, stable sets (e.g. message role) changed rarely |
| Text + CHECK | Sets expected to evolve often in early product |
| Lookup table | User-extensible or localized labels |

Rules:

- Name enums after domain language (`task_status`, `reminder_status`, `connection_status`)
- Migrations that alter enums need care — prefer additive values
- API exposes string enums matching domain language in [DOMAIN_MODEL.md](./DOMAIN_MODEL.md)

---

## Migration strategy

| Rule | Detail |
| --- | --- |
| Tool | golang-migrate (already in API image/entrypoint) |
| Location | `services/api/migrations` |
| Naming | `NNNNNN_description.up.sql` / `.down.sql` |
| Expand/contract | Prefer backward-compatible expands (add column/table) before contract (drop) |
| No auto-migrate in prod binary | Ops/entrypoint or explicit job runs migrations (local compose may migrate-on-start) |
| Domain schema | Introduced in a dedicated schema milestone — **not** in M2 domain-modeling docs |
| Review | Every migration reviewed for locks, indexes on large tables, and backfill plan |

Never edit an already-applied migration on shared environments; add a new migration instead.

---

## Transaction guidelines

| Rule | Detail |
| --- | --- |
| Boundary | One use case / aggregate command when possible |
| Cross-aggregate | Explicit application workflow; avoid giant transactions spanning calendar + chat + AI |
| AI tool flows | Persist user message → execute tools in API services (each with clear transactions) → persist assistant message |
| Idempotency | Provider sync and webhook handlers must be idempotent (unique provider keys) |
| Timeouts | Short DB transactions; do not hold locks while calling Google/OpenAI |
| Outbox (future) | For reliable side-effects after commit — not required to define tables now |

Handler layer never opens transactions directly; Business/Repository orchestration owns them.

---

## Audit logging strategy

Aligned with [OBSERVABILITY.md](./OBSERVABILITY.md) and the Audit domain in [DOMAIN_MODEL.md](./DOMAIN_MODEL.md).

| Concern | Standard |
| --- | --- |
| Store | Dedicated audit log entity/table (future schema) — append-oriented |
| When | Auth, connection link/unlink, reminder create/delete, settings changes, future billing |
| Contents | Actor user id, action, subject type/id, non-secret metadata, request_id, created_at |
| Retention | Longer than application logs; define policy at schema milestone |
| Dual write | Business operation succeeds then audit write; if audit fails, log ERROR and alert — product decision on fail-open vs fail-closed per action class |
| Not audit | Routine GETs, high-volume AI token ticks (those are AI usage / metrics) |

Application INFO logs may mirror audit events; **audit storage is authoritative** for compliance queries.

---

## Multi-provider integration philosophy

| Layer | Responsibility |
| --- | --- |
| Donna tables | Canonical User, Event, Task, Message, Memory, … |
| Link columns / satellite rows | `provider`, remote ids, etags, last_sync_at |
| Adapters (Go) | Map provider API ↔ Donna entities |
| Credentials | Connected Account secret material — encrypted at rest, never in frontend |

Rules:

1. **No `google_events` as source of truth.**
2. Upserts key on provider object id **within** a calendar source/connection.
3. Sync may store opaque `provider_payload` JSONB for debug/repair — UI and AI read Donna fields.
4. Disconnecting a Connected Account defines product behavior: keep Donna copies, mark read-only, or soft-delete — decide at schema/product milestone; domain allows all three.
5. Adding Outlook/Apple = new adapter + provider enum value, not a parallel schema.

---

## Multi-tenancy (Phase 1 and later)

| Phase | Model |
| --- | --- |
| Phase 1 | Single-user tenancy: every row traces to `user_id` |
| Later | Optional `workspace_id` for shared resources; personal data remains user-scoped |

Anticipate nullable `workspace_id` only when Teams work starts — do not invent workspace tables in the first personal schema unless required.

---

## Security & secrets

- Provider refresh tokens and API keys: encrypted column or external secret manager reference — **never** plain JSONB logs
- Row Level Security (optional later on Supabase): policies keyed by `user_id`
- Migrations must not print secrets; seed data uses placeholders only

---

## Alignment with Clean Architecture

```text
Handler        → no SQL
Business       → domain rules, transactions across repositories
Repository     → SQL / pgx; maps rows ↔ entities
Entity         → Donna domain types (DOMAIN_MODEL)
Model (DTO)    → API wire format
Adapters       → Google/Outlook/OpenAI outside the domain core
```

Database standards enforce that repositories speak Donna’s language, while adapters translate foreign languages at the edge.

---

## Checklist for future schema PRs

- [ ] Entities match [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) names and ownership
- [ ] UUID PKs; timestamptz; soft delete policy applied deliberately
- [ ] FKs and indexes justified
- [ ] JSONB limited to side-data
- [ ] Provider links are not canonical meaning
- [ ] Up + down migrations
- [ ] No secrets in fixtures or comments
- [ ] Audit/AI usage needs considered for sensitive or LLM paths
