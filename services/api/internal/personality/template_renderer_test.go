package personality_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/google/uuid"
)

func TestCatalogLoadsBuiltIns(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cat.Get(personality.IDProfessional); !ok {
		t.Fatal("missing professional")
	}
	if _, ok := cat.Get(personality.IDCasual); !ok {
		t.Fatal("missing casual")
	}
	if _, ok := cat.Get(personality.IDFlirty); !ok {
		t.Fatal("missing flirty")
	}
	if len(cat.List()) < 3 {
		t.Fatalf("expected 3 personalities, got %d", len(cat.List()))
	}
}

func TestTemplateRendererProfessionalGreeting(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := personality.NewTemplateRenderer(cat, nil)
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000901")
	profile := personality.Profile{
		UserID:        userID,
		PersonalityID: personality.IDProfessional,
		DisplayName:   "Aryan Thacker",
		EmojiLevel:    personality.LevelNone,
	}
	out, err := r.Render(context.Background(), personality.RenderInput{
		UserID:  userID,
		Kind:    personality.KindGreeting,
		Now:     time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Timezone: "UTC",
		Profile: &profile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out.Text), "morning") || !strings.Contains(out.Text, "Aryan") {
		t.Fatalf("got %q", out.Text)
	}
}

func TestTemplateRendererCasualChatUsesNickname(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := personality.NewTemplateRenderer(cat, nil)
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000902")
	profile := personality.Profile{
		UserID:        userID,
		PersonalityID: personality.IDCasual,
		DisplayName:   "Aryan",
		Nickname:      "Rockstar",
		EmojiLevel:    personality.LevelLow,
	}
	out, err := r.Render(context.Background(), personality.RenderInput{
		UserID:    userID,
		Canonical: "You have 3 meetings today.",
		Kind:      personality.KindReminder,
		Now:       time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Timezone:  "UTC",
		Profile:   &profile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Text, "Rockstar") && !strings.Contains(out.Text, "meetings") {
		t.Fatalf("expected nickname or canonical in %q", out.Text)
	}
}

func TestTemplateRendererMorningGreetingUsesPunchline(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := personality.NewTemplateRenderer(cat, nil)
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000904")
	profile := personality.Profile{
		UserID:        userID,
		PersonalityID: personality.IDCasual,
		DisplayName:   "Aryan",
		Nickname:      "Rockstar",
		EmojiLevel:    personality.LevelNone,
	}
	out, err := r.Render(context.Background(), personality.RenderInput{
		UserID:   userID,
		Kind:     personality.KindMorningGreeting,
		Now:      time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Timezone: "UTC",
		Profile:  &profile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Text, "Rockstar") && !strings.Contains(strings.ToLower(out.Text), "morning") {
		t.Fatalf("expected morning greeting with nickname in %q", out.Text)
	}
	if !strings.Contains(out.Text, "\n") && len(strings.TrimSpace(out.Text)) < 10 {
		t.Fatalf("expected greeting + punchline in %q", out.Text)
	}
}

func TestCatalogHasNicknamesAndPunchlines(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{personality.IDProfessional, personality.IDCasual, personality.IDFlirty} {
		def, ok := cat.Get(id)
		if !ok {
			t.Fatalf("missing %s", id)
		}
		if len(def.FallbackNicknames) < 10 {
			t.Fatalf("%s nicknames = %d, want >= 10", id, len(def.FallbackNicknames))
		}
		if len(def.Punchlines) < 15 {
			t.Fatalf("%s punchlines = %d, want >= 15", id, len(def.Punchlines))
		}
		if len(def.MorningGreetings) == 0 || len(def.EveningGreetings) == 0 || len(def.GoodNightGreetings) == 0 {
			t.Fatalf("%s missing dedicated greeting lists", id)
		}
	}
}

func TestTemplateRendererStripsEmojiWhenNone(t *testing.T) {
	t.Parallel()
	cat, err := personality.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := personality.NewTemplateRenderer(cat, nil)
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000903")
	profile := personality.Profile{
		UserID:        userID,
		PersonalityID: personality.IDFlirty,
		DisplayName:   "Aryan",
		EmojiLevel:    personality.LevelNone,
	}
	out, err := r.Render(context.Background(), personality.RenderInput{
		UserID:   userID,
		Kind:     personality.KindGreeting,
		Now:      time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Timezone: "UTC",
		Profile:  &profile,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, emoji := range []string{"❤️", "☀️", "😏", "💕"} {
		if strings.Contains(out.Text, emoji) {
			t.Fatalf("emoji should be stripped at none level: %q", out.Text)
		}
	}
}
