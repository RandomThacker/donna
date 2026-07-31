package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IntentKind classifies a user command.
type IntentKind string

// MVP-supported intents (keep this list small and reliable).
const (
	IntentCreateTask     IntentKind = "CREATE_TASK"
	IntentCompleteTask   IntentKind = "COMPLETE_TASK"
	IntentCreateReminder IntentKind = "CREATE_REMINDER"
	IntentCreateEvent    IntentKind = "CREATE_EVENT"
	IntentQueryToday     IntentKind = "QUERY_TODAY"
	IntentQueryTomorrow  IntentKind = "QUERY_TOMORROW"
	IntentQueryDueToday  IntentKind = "QUERY_DUE_TODAY"
	IntentGreeting       IntentKind = "GREETING"
	IntentUnknown        IntentKind = "UNKNOWN"
)

// Intent is the parser output — transport-agnostic.
// All IntentParser implementations (rule-based, OpenAI, Claude, …) return this shape.
type Intent struct {
	Kind           IntentKind
	Raw            string
	Title          string
	Description    *string
	StartAt        *time.Time
	EndAt          *time.Time
	TriggerAt      *time.Time
	Timezone       string
	AllDay         bool
	RecurrenceRule *string
	// TargetTitle resolves complete-by-name.
	TargetTitle string
	Completed   *bool
	Confidence  float64
}

// IntentParser classifies natural language into an Intent.
// Chat depends only on this interface — swap RuleBasedParser for OpenAIParser later
// without changing the executor or handlers.
type IntentParser interface {
	Parse(ctx context.Context, input string) (*Intent, error)
}

type parseCtxKey int

const (
	parseKeyNow parseCtxKey = iota + 1
	parseKeyTimezone
)

// WithParseNow attaches the clock used for relative dates ("tomorrow", "at 6 PM").
func WithParseNow(ctx context.Context, now time.Time) context.Context {
	return context.WithValue(ctx, parseKeyNow, now)
}

// WithParseTimezone attaches the user timezone for local date/time extraction.
func WithParseTimezone(ctx context.Context, tz string) context.Context {
	return context.WithValue(ctx, parseKeyTimezone, tz)
}

func parseNowFrom(ctx context.Context) time.Time {
	if v, ok := ctx.Value(parseKeyNow).(time.Time); ok && !v.IsZero() {
		return v
	}
	return time.Now()
}

func parseTimezoneFrom(ctx context.Context) string {
	if v, ok := ctx.Value(parseKeyTimezone).(string); ok {
		return v
	}
	return ""
}

// UnknownHelp is the fixed reply when parsing fails.
const UnknownHelp = `I couldn't understand that.

Try something like
• Add task Finish API
• Complete task Finish API
• Remind me tomorrow at 6 PM
• Schedule meeting Standup tomorrow at 10 AM
• What do I have today?
• What do I have tomorrow?
• What's due today?`

// CommandRequest is POST /chat/command.
type CommandRequest struct {
	Message         string `json:"message"`
	ClientMessageID string `json:"client_message_id,omitempty"`
}

// CommandResult is returned to any chat UI.
type CommandResult struct {
	Reply  string     `json:"reply"`
	Intent IntentKind `json:"intent"`
}

// ExecuteInput is required by the executor (not part of Intent).
type ExecuteInput struct {
	UserID      uuid.UUID
	Timezone    string
	Now         time.Time
	Message     string
	DisplayName string
}
