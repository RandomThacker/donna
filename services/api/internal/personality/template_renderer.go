package personality

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// TemplateRenderer is the Phase 1 personality renderer (config-driven templates).
type TemplateRenderer struct {
	catalog Catalog
	store   ProfileStore
	now     func() time.Time
}

// NewTemplateRenderer constructs a template-based PersonalityRenderer.
func NewTemplateRenderer(catalog Catalog, store ProfileStore) *TemplateRenderer {
	return &TemplateRenderer{
		catalog: catalog,
		store:   store,
		now:     time.Now,
	}
}

// Render implements Renderer.
func (r *TemplateRenderer) Render(ctx context.Context, input RenderInput) (RenderOutput, error) {
	if r == nil || r.catalog == nil {
		return RenderOutput{Text: strings.TrimSpace(input.Canonical)}, nil
	}

	profile := input.Profile
	if profile == nil && r.store != nil && input.UserID != uuid.Nil {
		loaded, err := r.store.Get(ctx, input.UserID)
		if err == nil {
			profile = &loaded
		}
	}
	if profile == nil {
		def := DefaultProfile(input.UserID)
		profile = &def
	}

	def, ok := r.catalog.Get(profile.PersonalityID)
	if !ok {
		def, _ = r.catalog.Get(DefaultID)
	}

	now := input.Now
	if now.IsZero() {
		now = r.now()
	}
	tz := strings.TrimSpace(input.Timezone)
	if tz == "" {
		tz = "UTC"
	}

	canonical := strings.TrimSpace(input.Canonical)
	vars := r.buildVars(*profile, def, now, tz, canonical)
	text := r.renderKind(def, input.Kind, vars, canonical)
	text = applyEmojiLevel(text, profile.EmojiLevel)
	text = strings.TrimSpace(text)
	if text == "" {
		text = canonical
	}
	return RenderOutput{Text: text}, nil
}

func (r *TemplateRenderer) renderKind(def Definition, kind Kind, vars map[string]string, canonical string) string {
	switch kind {
	case KindGreeting:
		return vars["greeting"]
	case KindReminder:
		return fill(pick(def.Reminders, "{canonical}"), vars)
	case KindNotification:
		return fill(pick(def.Notifications, "{canonical}"), vars)
	case KindAutomation:
		intro := strings.TrimSpace(fill(pick(def.AutomationIntros, ""), vars))
		closing := strings.TrimSpace(fill(pick(def.Closings, ""), vars))
		parts := make([]string, 0, 3)
		if intro != "" {
			parts = append(parts, intro)
		}
		if canonical != "" {
			parts = append(parts, canonical)
		}
		if closing != "" {
			parts = append(parts, closing)
		}
		return strings.Join(parts, "\n\n")
	case KindMorningBrief:
		return fill(pick(def.MorningBriefs, "{greeting}\n\n{canonical}"), vars)
	case KindTaskComplete:
		return fill(pick(def.TaskComplete, "{canonical}"), vars)
	case KindError:
		return fill(pick(def.Errors, "{canonical}"), vars)
	case KindAcknowledgement:
		tpl := pick(def.Acknowledgements, "{canonical}")
		if !strings.Contains(tpl, "{canonical}") && canonical != "" {
			return fill(tpl, vars) + "\n\n" + canonical
		}
		return fill(tpl, vars)
	case KindChat:
		fallthrough
	default:
		return fill(pick(def.ChatWrappers, "{canonical}"), vars)
	}
}

func (r *TemplateRenderer) buildVars(profile Profile, def Definition, now time.Time, tz, canonical string) map[string]string {
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		loc = time.UTC
	}
	local := now.In(loc)
	period := dayPeriod(local.Hour())

	name := firstName(profile.DisplayName)
	if name == "" {
		name = "there"
	}
	nickname := strings.TrimSpace(profile.Nickname)
	if nickname == "" {
		nickname = pick(def.FallbackNicknames, name)
	}
	if nickname == "" {
		nickname = name
	}

	greetingTpl := pick(def.Greetings[period], "Hello, {name}.")
	emoji := pickEmoji(def, profile.EmojiLevel)

	vars := map[string]string{
		"name":       name,
		"nickname":   nickname,
		"canonical":  canonical,
		"period":     period,
		"emoji":      emoji,
		"greeting":   "", // filled below after recursive-safe fill
	}
	vars["greeting"] = fill(greetingTpl, vars)
	return vars
}

// DefaultProfile returns Professional defaults for a user.
func DefaultProfile(userID uuid.UUID) Profile {
	return Profile{
		UserID:             userID,
		PersonalityID:      DefaultID,
		EmojiLevel:         LevelNone,
		HumorLevel:         LevelNone,
		EncouragementLevel: LevelLow,
		ResponseStyle:      "concise",
		GreetingStyle:      "formal",
	}
}

func dayPeriod(hour int) string {
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

func firstName(displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return ""
	}
	parts := strings.Fields(displayName)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func pick(options []string, fallback string) string {
	clean := make([]string, 0, len(options))
	for _, o := range options {
		// Allow intentionally empty automation intros.
		clean = append(clean, o)
	}
	if len(clean) == 0 {
		return fallback
	}
	// Prefer non-empty when fallback matters; still allow empty picks for intros.
	nonEmpty := make([]string, 0, len(clean))
	for _, o := range clean {
		if strings.TrimSpace(o) != "" {
			nonEmpty = append(nonEmpty, o)
		}
	}
	pool := nonEmpty
	if len(pool) == 0 {
		if fallback != "" {
			return fallback
		}
		pool = clean
	}
	return pool[rand.IntN(len(pool))]
}

func pickEmoji(def Definition, level Level) string {
	if level == LevelNone {
		return ""
	}
	key := string(level)
	list := def.Emojis[key]
	if len(list) == 0 && level == LevelHigh {
		list = def.Emojis[string(LevelMedium)]
	}
	if len(list) == 0 && (level == LevelHigh || level == LevelMedium) {
		list = def.Emojis[string(LevelLow)]
	}
	if len(list) == 0 {
		return ""
	}
	return list[rand.IntN(len(list))]
}

func fill(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	// Collapse leftover empty emoji spacing.
	out = strings.ReplaceAll(out, "  ", " ")
	return strings.TrimSpace(out)
}

func applyEmojiLevel(text string, level Level) string {
	if level != LevelNone {
		return text
	}
	return stripEmojis(text)
}

func stripEmojis(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isEmojiRune(r) {
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}
	return strings.TrimSpace(out)
}

func isEmojiRune(r rune) bool {
	if r == 0xFE0F || r == 0x200D { // variation selector / ZWJ
		return true
	}
	// Common emoji blocks (good enough for template scrubbing).
	switch {
	case r >= 0x1F300 && r <= 0x1FAFF:
		return true
	case r >= 0x2600 && r <= 0x27BF:
		return true
	case r >= 0x1F1E6 && r <= 0x1F1FF:
		return true
	case unicode.Is(unicode.So, r) && r > 0x2000:
		return true
	default:
		return false
	}
}
