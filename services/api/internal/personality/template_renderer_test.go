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
	if !strings.Contains(out.Text, "Good morning, Aryan") {
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
		UserID:  userID,
		Kind:    personality.KindGreeting,
		Now:     time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Timezone: "UTC",
		Profile: &profile,
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
