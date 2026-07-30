# Command Chat (Phase 3.0)

Donna Chat is a **conversational command interface** over the Action Layer.

It is not an AI assistant. No planning, memory, RAG, or LLM agents in this phase.

```text
Chat UI
   ↓
IntentParser  ← interface (swap implementations later)
   ↓
Actions
   ↓
Services
```

## IntentParser

Chat never hard-codes a concrete parser.

```go
type IntentParser interface {
    Parse(ctx context.Context, input string) (*Intent, error)
}
```

| Today | Later (same Intent DTO) |
| --- | --- |
| `RuleBasedParser` | `OpenAIParser`, `ClaudeParser`, `GeminiParser` |

Relative dates use context helpers: `WithParseNow`, `WithParseTimezone`.

Wire in `app.go`:

```go
chat.NewExecutor(chat.NewRuleBasedParser(), actionRegistry)
```

## MVP commands (intentionally small)

| Say | Intent |
| --- | --- |
| Add task Finish API | `CREATE_TASK` |
| Complete task Finish API | `COMPLETE_TASK` |
| Remind me tomorrow at 6 PM | `CREATE_REMINDER` |
| Schedule meeting Standup tomorrow at 10 AM | `CREATE_EVENT` |
| What do I have today? | `QUERY_TODAY` |
| What do I have tomorrow? | `QUERY_TOMORROW` |
| What's due today? | `QUERY_DUE_TODAY` |

Everything else → help reply. Expand only after these feel effortless.

## API

`POST /api/v1/chat/command` → `{ "reply", "intent" }`

History is browser session only — not persisted.

## Command guide (web)

`/dashboard/commands` lists every MVP phrase with copy + “try in chat”
(`?prefill=`). Also linked from Settings → Commands.

## Out of scope

AI · Memory · Embeddings · Delete/Update commands · Week/notifications queries · Telegram · Voice
