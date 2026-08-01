# Personality Engine

Donna's Personality Engine controls **how** she speaks — not what she does.

```text
Business Logic
      ↓
Canonical Response
      ↓
Personality Renderer
      ↓
Final Response
```

Business logic, Action Layer mutations, scheduling, and execution history stay unchanged.

## Package

`services/api/internal/personality`

| Piece | Role |
| --- | --- |
| `Renderer` | Interface: `Render(ctx, RenderInput) (RenderOutput, error)` |
| `TemplateRenderer` | Phase 1 config-driven implementation |
| `Catalog` | Loads built-in YAML personalities |
| `Profile` | Per-user preferences (name, nickname, levels) |

Future AI / LLM / Markdown renderers implement the same `Renderer` interface.

## Built-in personalities

Configuration (embedded):

- `internal/personality/config/personalities/professional.yaml`
- `internal/personality/config/personalities/casual.yaml`
- `internal/personality/config/personalities/flirty.yaml`

Each file defines:

- `fallback_nicknames` (10–15+)
- `punchlines` (15–30 tone-matched lines)
- period `greetings` (morning / afternoon / evening / night)
- dedicated `morning_greetings`, `evening_greetings`, `goodnight_greetings`
- acknowledgements, task-complete, errors, reminders, notifications, automation intros, closings, chat wrappers

Placeholders: `{name}`, `{nickname}`, `{emoji}`, `{punchline}`, `{greeting}`, `{canonical}`.

Greeting kinds: `greeting`, `morning_greeting`, `evening_greeting`, `good_night`.

## Persistence

Table: `user_personality` (migration `000035`)

One row per user. Existing users default to **Professional**.

## Wire-up

| Surface | Kind |
| --- | --- |
| Chat commands | greeting / chat / acknowledgement / task_complete / error |
| Automations | automation / morning_brief (combined reply) |
| Notification chat bubbles | notification / reminder |

Canonical notification center / push payloads stay factual.

## API

| Method | Path |
| --- | --- |
| GET | `/api/v1/settings/personality` |
| PATCH | `/api/v1/settings/personality` |
| GET | `/api/v1/settings/personality/catalog` |
| POST | `/api/v1/settings/personality/preview` |

## UI

**Settings → Personality** — choose personality, preferred name, nickname, emoji level, with live preview samples.

## Explicit non-goals

- AI-generated personalities
- Voice / avatar
- Long-term memory
