# Automations

Scheduled executions of Donna chat commands. Evolved from Scheduled Intent Check-ins into a generic **Automation Engine**.

The scheduler does **not** know about Tasks, Calendar, or specific intents. It only knows:

```text
Trigger → Command(s) → Chat Executor → Delivery
```

Exactly the same path as if the user typed those commands in chat.

Automations are **templates**. Each run creates an **execution** (history) — except dry-run previews.

## Architecture

```text
Settings UI / Templates
        ↓
    Automation Actions → AutomationService → automations table
        ↓
AutomationScheduler (minute tick)  OR  Manual Run  OR  Preview
        ↓
AutomationRunner
        ↓
Begin automation_executions (RUNNING)   [skipped for Preview]
        ↓
for each structured command → resolve → chat.Executor → record command rows
        ↓
combine replies → PostAssistantNotice (chat)   [skipped for Preview]
        ↓
Complete execution (SUCCESS / PARTIAL_SUCCESS / FAILED)
        ↓
last_run_at / next_run_at   [scheduled runs only]
```

Intent Catalog (`internal/automationcatalog/templates.yaml`) is **configuration**, not business logic.

## Domain

### Automation (template)

| Field | Notes |
| --- | --- |
| `name`, `description` | Display |
| `enabled` | Soft pause |
| `trigger` | `{ type: "daily"\|"weekly", time: "HH:MM", days?: ["MO"…"SU"] }` — `days` required for weekly |
| `commands` | Ordered structured steps: `{ command, variables }` |
| `delivery` | `{ channels: ["chat","push"] }` — chat bubble + Web Push |
| `last_run_at` / `next_run_at` | Schedule markers (updated by scheduled runs only) |

Public ids: `atm_`.

### Structured commands

| Key | Variables | Resolves to |
| --- | --- | --- |
| `greeting` | — | `Hi` |
| `todays_agenda` | `range: today\|tomorrow` | agenda query |
| `tasks_due` | `priority: all` (reserved) | due-today query |
| `chat_message` | `message` | free-text chat line |

API still accepts legacy string commands (stored as `chat_message`). UI shows readable `label`s.

### Execution (history)

| Field | Notes |
| --- | --- |
| `status` | RUNNING, SUCCESS, PARTIAL_SUCCESS, FAILED, CANCELLED |
| `duration_ms` | Wall time |
| `commands_*` | Counts |
| `trigger_source` | `scheduler`, `manual` (`preview` is never persisted) |
| `delivery_status` | PENDING, SENT, FAILED, SKIPPED |
| `response` | Combined assistant message |

Public ids: `aex_`. Per-command rows use `ace_`.

## API

| Method | Path |
| --- | --- |
| GET | `/api/v1/automations` (includes success rate, last status, avg duration) |
| GET | `/api/v1/automations/templates` |
| GET | `/api/v1/automations/history` |
| GET | `/api/v1/automations/analytics` |
| GET | `/api/v1/automations/:id/history` |
| GET | `/api/v1/automations/executions/:id` |
| POST | `/api/v1/automations/:id/run` — Manual Run (`trigger_source=manual`) |
| POST | `/api/v1/automations/:id/preview` — Dry run (no history / delivery / mutations) |
| POST / PATCH / DELETE | `/api/v1/automations` |

Development builds include a `debug` object on execution detail (raw command outputs, timing, payload). Hidden in production.

## UI

- **Settings → Automations** — templates, Run Now, Preview, enable/delete, last run metrics
- **Dashboard → Automations** — history list with expandable command results + analytics strip

## Migrations

- `000032` — automations table (+ migrate legacy intent_rules)
- `000033` — automation_executions + automation_command_executions
- `000034` — structured commands (`command` + `variables`)
- `000036` — weekly trigger days (`trigger_days`, `daily|weekly`)
- `000037` — add `push` delivery channel (default chat+push)

## Personality

Combined automation replies go through the Personality Engine (`KindAutomation` / `KindGreeting` / `KindMorningBrief`). Per-command chat execution skips personality so the wrap owns tone, name, and nickname.

## Delivery

After commands finish, Donna posts one combined reply to **chat** and fans out the same text via **Web Push** (when VAPID is configured and the user has a subscribed device). Preview skips both.

## Explicit non-goals (this phase)

- Retries / replay / conditions
- Telegram / WhatsApp / Email delivery
- AI-authored automations
- Cron expressions / time windows beyond daily & weekly weekdays

## Future compatibility

Execution records and trigger sources are shaped for retry, replay, multi-channel delivery, AI summaries, and timeline UIs without rewriting the run model.
