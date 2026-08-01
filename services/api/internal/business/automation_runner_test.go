package business

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
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
	if out.Response == "" {
		t.Fatal("expected personalized greeting")
	}
	// Flirty evening templates may omit the nickname on some picks.
	if !strings.Contains(out.Response, "Rockstar") &&
		!strings.Contains(out.Response, "Hey you") &&
		!strings.Contains(out.Response, "Evening") {
		t.Fatalf("expected flirty greeting in %q", out.Response)
	}
}

type countingPushSender struct {
	calls int
}

func (s *countingPushSender) Configured() bool { return true }

func (s *countingPushSender) Send(_ context.Context, _ entity.PushSubscription, _ webpush.Payload) (webpush.Result, error) {
	s.calls++
	return webpush.Result{StatusCode: 201}, nil
}

type noopNotices struct{}

func (noopNotices) PostAssistantNotice(_ context.Context, _ uuid.UUID, _ string, _ string) (entity.Message, bool, error) {
	return entity.Message{}, false, nil
}

func TestAutomationRunnerDeliversWebPush(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000a11")
	pushRepo := newMemPushSubRepo()
	_, err := pushRepo.Upsert(context.Background(), entity.PushSubscription{
		ID:       uuid.MustParse("018f0000-0000-7000-8000-000000000a12"),
		PublicID: "psub_auto",
		UserID:   userID,
		Endpoint: "https://push.example/auto",
		P256dh:   "p256",
		Auth:     "auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	sender := &countingPushSender{}
	runner := NewAutomationRunner(nil, greetingOnlyChat{}, noopNotices{}, nil, nil)
	runner.SetWebPush(NewPushSubscriptionService(pushRepo), sender)

	auto := entity.Automation{
		ID:               uuid.MustParse("018f0000-0000-7000-8000-000000000a13"),
		PublicID:         "atm_push",
		UserID:           userID,
		Name:             "Morning Brief",
		Enabled:          true,
		TriggerType:      constant.AutomationTriggerDaily,
		TriggerTime:      "09:00",
		Timezone:         "UTC",
		Commands:         []entity.AutomationCommand{{Command: constant.AutomationCommandGreeting}},
		DeliveryChannels: []string{constant.AutomationDeliveryChat, constant.AutomationDeliveryPush},
	}

	out, err := runner.Run(context.Background(), auto, AutomationRunOptions{
		TriggerSource: constant.AutomationTriggerSourceManual,
		DeliverToChat: true,
		Now:           time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sender.calls != 1 {
		t.Fatalf("expected 1 push send, got %d", sender.calls)
	}
	if out.DeliveryStatus != constant.AutomationDeliverySent {
		t.Fatalf("delivery status = %s", out.DeliveryStatus)
	}
}
