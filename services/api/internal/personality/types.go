package personality

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Kind classifies the type of message being rendered.
type Kind string

const (
	KindChat            Kind = "chat"
	KindGreeting        Kind = "greeting"
	KindMorningGreeting Kind = "morning_greeting"
	KindEveningGreeting Kind = "evening_greeting"
	KindGoodNight       Kind = "good_night"
	KindReminder        Kind = "reminder"
	KindNotification    Kind = "notification"
	KindAutomation      Kind = "automation"
	KindAutomationBody   Kind = "automation_body"
	KindMorningBrief    Kind = "morning_brief"
	KindMorningBriefFrame Kind = "morning_brief_frame"
	KindTaskComplete    Kind = "task_complete"
	KindError           Kind = "error"
	KindAcknowledgement Kind = "acknowledgement"
)

// Level is a discrete intensity setting (emoji, humor, encouragement).
type Level string

const (
	LevelNone   Level = "none"
	LevelLow    Level = "low"
	LevelMedium Level = "medium"
	LevelHigh   Level = "high"
)

// Built-in personality ids.
const (
	IDProfessional = "professional"
	IDCasual       = "casual"
	IDFlirty       = "flirty"
	DefaultID      = IDProfessional
)

// Profile is the per-user personality preference (DB + effective overrides).
type Profile struct {
	UserID              uuid.UUID
	PersonalityID       string
	DisplayName         string
	Nickname            string
	EmojiLevel          Level
	HumorLevel          Level
	GreetingStyle       string
	EncouragementLevel  Level
	ResponseStyle       string
}

// RenderInput is the canonical → personalized transform request.
type RenderInput struct {
	UserID     uuid.UUID
	Canonical  string
	Kind       Kind
	Now        time.Time
	Timezone   string
	// Profile, when set, skips loading from storage (preview / tests).
	Profile *Profile
}

// RenderOutput is the final personalized text.
type RenderOutput struct {
	Text string
}

// Renderer transforms canonical responses into personalized responses.
// Future AI/LLM/Markdown renderers implement the same contract.
type Renderer interface {
	Render(ctx context.Context, input RenderInput) (RenderOutput, error)
}

// ProfileStore loads and persists user personality preferences.
type ProfileStore interface {
	Get(ctx context.Context, userID uuid.UUID) (Profile, error)
	Upsert(ctx context.Context, profile Profile) (Profile, error)
}

// Catalog lists built-in personality definitions.
type Catalog interface {
	List() []Definition
	Get(id string) (Definition, bool)
}

// Definition is a built-in personality loaded from configuration.
type Definition struct {
	ID                        string
	Name                      string
	Description               string
	EmojiLevelDefault         Level
	HumorLevelDefault         Level
	EncouragementLevelDefault Level
	ResponseStyleDefault      string
	FallbackNicknames         []string
	Punchlines                []string
	Greetings                 map[string][]string // morning|afternoon|evening|night
	MorningGreetings          []string
	EveningGreetings          []string
	GoodNightGreetings        []string
	Acknowledgements          []string
	TaskComplete              []string
	Errors                    []string
	Reminders                 []string
	Notifications             []string
	AutomationIntros          []string
	AutomationBodies          []string
	Closings                  []string
	Encouragements            []string
	ChatWrappers              []string
	MorningBriefs             []string
	MorningBriefFrames        []string
	Emojis                    map[string][]string // low|medium|high
}
