package business

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/google/uuid"
)

type greetingOnlyChat struct{}

func (greetingOnlyChat) Execute(_ context.Context, in ChatCommandInput) ChatCommandResult {
	if !in.SkipPersonality {
		t := "chat did not skip personality"
		return ChatCommandResult{Error: t}
	}
	return ChatCommandResult{Reply: "", Intent: "greeting"}
}

type fixedPersonalityStore struct {
	profile personality.Profile
}

func (s fixedPersonalityStore) Get(_ context.Context, _ uuid.UUID) (personality.Profile, error) {
	return s.profile, nil
}

func (s fixedPersonalityStore) Upsert(_ context.Context, profile personality.Profile) (personality.Profile, error) {
	return profile, nil
}

func TestAutomationRunnerGreetingUsesPersonality(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000a01")
	store := fixedPersonalityStore{profile: personality.Profile{
		UserID:        userID,
		PersonalityID: personality.IDFlirty,
		DisplayName:   "Aryan Thacker",
		Nickname:      "Rockstar",
		EmojiLevel:    personality.LevelNone,
	}}
	renderer := personality.NewTemplateRenderer(cat, store)
	runner := NewAutomationRunner(nil, greetingOnlyChat{}, nil, nil, nil)
	runner.SetPersonality(renderer)
	runner.now = func() time.Time {
		return time.Date(2026, 8, 1, 18, 30, 0, 0, time.UTC)
	}

	auto := entity.Automation{
		ID:               uuid.MustParse("018f0000-0000-7000-8000-000000000a02"),
		PublicID:         "atm_test",
		UserID:           userID,
		Name:             "Evening Hello",
		Enabled:          true,
		TriggerType:      constant.AutomationTriggerDaily,
		TriggerTime:      "18:30",
		Timezone:         "UTC",
		Commands:         []entity.AutomationCommand{{Command: constant.AutomationCommandGreeting}},
		DeliveryChannels: []string{constant.AutomationDeliveryChat},
	}

	out, err := runner.Run(context.Background(), auto, AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourceManual,
		RecordHistory:  false,
		DeliverToChat:  false,
		UpdateSchedule: false,
		Now:            time.Date(2026, 8, 1, 18, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.Response, "Good evening, there") {
		t.Fatalf("generic fallback leaked into automation reply: %q", out.Response)
	}
	if !strings.Contains(out.Response, "Rockstar") {
		t.Fatalf("expected flirty nickname in %q", out.Response)
	}
}
