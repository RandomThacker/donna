package business

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// PersonalityService manages user personality preferences and acts as a ProfileStore.
type PersonalityService struct {
	repo  repository.UserPersonalityRepository
	users repository.UserRepository
	cat   personality.Catalog
	now   func() time.Time
}

// NewPersonalityService constructs a PersonalityService.
func NewPersonalityService(
	repo repository.UserPersonalityRepository,
	users repository.UserRepository,
	cat personality.Catalog,
) *PersonalityService {
	return &PersonalityService{repo: repo, users: users, cat: cat, now: time.Now}
}

// Get implements personality.ProfileStore.
func (s *PersonalityService) Get(ctx context.Context, userID uuid.UUID) (personality.Profile, error) {
	if userID == uuid.Nil {
		return personality.Profile{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	row, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return s.ensureDefault(ctx, userID)
		}
		return personality.Profile{}, err
	}
	return s.toProfile(ctx, row), nil
}

// Upsert implements personality.ProfileStore.
func (s *PersonalityService) Upsert(ctx context.Context, profile personality.Profile) (personality.Profile, error) {
	if profile.UserID == uuid.Nil {
		return personality.Profile{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	normalized, err := s.normalizeProfile(profile)
	if err != nil {
		return personality.Profile{}, err
	}

	existing, err := s.repo.GetByUserID(ctx, normalized.UserID)
	now := s.now().UTC()
	row := entity.UserPersonality{
		UserID:             normalized.UserID,
		PersonalityID:      normalized.PersonalityID,
		EmojiLevel:         string(normalized.EmojiLevel),
		HumorLevel:         string(normalized.HumorLevel),
		GreetingStyle:      normalized.GreetingStyle,
		EncouragementLevel: string(normalized.EncouragementLevel),
		ResponseStyle:      normalized.ResponseStyle,
		UpdatedAt:          now,
	}
	if dn := strings.TrimSpace(normalized.DisplayName); dn != "" {
		row.DisplayName = &dn
	}
	if nn := strings.TrimSpace(normalized.Nickname); nn != "" {
		row.Nickname = &nn
	}

	if err == nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
	} else if errors.Is(err, apperr.ErrNotFound) {
		id, genErr := idgen.NewUUIDv7()
		if genErr != nil {
			return personality.Profile{}, genErr
		}
		row.ID = id
		row.CreatedAt = now
	} else {
		return personality.Profile{}, err
	}

	saved, err := s.repo.Upsert(ctx, row)
	if err != nil {
		return personality.Profile{}, err
	}
	return s.toProfile(ctx, saved), nil
}

// UpdateInput patches personality preferences.
type PersonalityUpdateInput struct {
	PersonalityID      *string
	DisplayName        *string
	Nickname           *string
	EmojiLevel         *string
	HumorLevel         *string
	GreetingStyle      *string
	EncouragementLevel *string
	ResponseStyle      *string
}

// Update patches the user's personality preferences.
func (s *PersonalityService) Update(ctx context.Context, userID uuid.UUID, in PersonalityUpdateInput) (personality.Profile, error) {
	current, err := s.Get(ctx, userID)
	if err != nil {
		return personality.Profile{}, err
	}
	if in.PersonalityID != nil {
		current.PersonalityID = strings.TrimSpace(*in.PersonalityID)
	}
	if in.DisplayName != nil {
		current.DisplayName = strings.TrimSpace(*in.DisplayName)
	}
	if in.Nickname != nil {
		current.Nickname = strings.TrimSpace(*in.Nickname)
	}
	if in.EmojiLevel != nil {
		current.EmojiLevel = personality.Level(strings.TrimSpace(*in.EmojiLevel))
	}
	if in.HumorLevel != nil {
		current.HumorLevel = personality.Level(strings.TrimSpace(*in.HumorLevel))
	}
	if in.GreetingStyle != nil {
		current.GreetingStyle = strings.TrimSpace(*in.GreetingStyle)
	}
	if in.EncouragementLevel != nil {
		current.EncouragementLevel = personality.Level(strings.TrimSpace(*in.EncouragementLevel))
	}
	if in.ResponseStyle != nil {
		current.ResponseStyle = strings.TrimSpace(*in.ResponseStyle)
	}
	return s.Upsert(ctx, current)
}

// ListDefinitions returns built-in personalities.
func (s *PersonalityService) ListDefinitions() ([]personality.Definition, error) {
	if s.cat == nil {
		return nil, fmt.Errorf("%w: personality catalog is required", apperr.ErrInvalid)
	}
	return s.cat.List(), nil
}

// Preview renders sample outputs for the given (or current) profile without saving.
func (s *PersonalityService) Preview(
	ctx context.Context,
	userID uuid.UUID,
	renderer personality.Renderer,
	override *personality.Profile,
	timezone string,
) (map[string]string, error) {
	if renderer == nil {
		return nil, fmt.Errorf("%w: renderer is required", apperr.ErrInvalid)
	}
	profile := override
	if profile == nil {
		loaded, err := s.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		profile = &loaded
	} else {
		normalized, err := s.normalizeProfile(*profile)
		if err != nil {
			return nil, err
		}
		normalized.UserID = userID
		profile = &normalized
	}
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		tz = constant.DefaultUserTimezone
	}
	now := s.now()
	samples := map[string]string{
		"greeting":        "",
		"reminder":        "Your meeting begins in 10 minutes.",
		"task_complete":   "Task marked complete.",
		"error":           "Something went wrong on my end.",
		"notification":    "Standup — Starts in 10 minutes",
		"automation":      "You have 3 meetings today.\n\nNothing due today.",
		"morning_brief":   "You have 3 meetings today.\n\n2 tasks still open.",
		"chat":            "You have 3 meetings today.",
	}
	kinds := map[string]personality.Kind{
		"greeting":      personality.KindGreeting,
		"reminder":      personality.KindReminder,
		"task_complete": personality.KindTaskComplete,
		"error":         personality.KindError,
		"notification":  personality.KindNotification,
		"automation":    personality.KindAutomation,
		"morning_brief": personality.KindMorningBrief,
		"chat":          personality.KindChat,
	}
	out := make(map[string]string, len(samples))
	for key, canonical := range samples {
		rendered, err := renderer.Render(ctx, personality.RenderInput{
			UserID:    userID,
			Canonical: canonical,
			Kind:      kinds[key],
			Now:       now,
			Timezone:  tz,
			Profile:   profile,
		})
		if err != nil {
			return nil, err
		}
		out[key] = rendered.Text
	}
	return out, nil
}

func (s *PersonalityService) ensureDefault(ctx context.Context, userID uuid.UUID) (personality.Profile, error) {
	profile := personality.DefaultProfile(userID)
	if s.users != nil {
		if user, err := s.users.GetByID(ctx, userID); err == nil && user.DisplayName != nil {
			profile.DisplayName = strings.TrimSpace(*user.DisplayName)
		}
	}
	if s.cat != nil {
		if def, ok := s.cat.Get(personality.DefaultID); ok {
			profile.EmojiLevel = def.EmojiLevelDefault
			profile.HumorLevel = def.HumorLevelDefault
			profile.EncouragementLevel = def.EncouragementLevelDefault
			profile.ResponseStyle = def.ResponseStyleDefault
		}
	}
	return s.Upsert(ctx, profile)
}

func (s *PersonalityService) normalizeProfile(in personality.Profile) (personality.Profile, error) {
	id := strings.ToLower(strings.TrimSpace(in.PersonalityID))
	if id == "" {
		id = personality.DefaultID
	}
	if s.cat != nil {
		if _, ok := s.cat.Get(id); !ok {
			return personality.Profile{}, fmt.Errorf("%w: unknown personality_id", apperr.ErrValidation)
		}
	}
	out := in
	out.PersonalityID = id
	out.EmojiLevel = normalizeLevel(string(in.EmojiLevel), personality.LevelNone)
	out.HumorLevel = normalizeLevel(string(in.HumorLevel), personality.LevelNone)
	out.EncouragementLevel = normalizeLevel(string(in.EncouragementLevel), personality.LevelLow)
	if strings.TrimSpace(out.GreetingStyle) == "" {
		out.GreetingStyle = "formal"
	}
	if strings.TrimSpace(out.ResponseStyle) == "" {
		out.ResponseStyle = "concise"
	}
	return out, nil
}

func (s *PersonalityService) toProfile(ctx context.Context, row entity.UserPersonality) personality.Profile {
	p := personality.Profile{
		UserID:             row.UserID,
		PersonalityID:      row.PersonalityID,
		EmojiLevel:         personality.Level(row.EmojiLevel),
		HumorLevel:         personality.Level(row.HumorLevel),
		GreetingStyle:      row.GreetingStyle,
		EncouragementLevel: personality.Level(row.EncouragementLevel),
		ResponseStyle:      row.ResponseStyle,
	}
	if row.DisplayName != nil {
		p.DisplayName = strings.TrimSpace(*row.DisplayName)
	}
	if row.Nickname != nil {
		p.Nickname = strings.TrimSpace(*row.Nickname)
	}
	if p.DisplayName == "" && s.users != nil {
		if user, err := s.users.GetByID(ctx, row.UserID); err == nil && user.DisplayName != nil {
			p.DisplayName = strings.TrimSpace(*user.DisplayName)
		}
	}
	return p
}

func normalizeLevel(raw string, fallback personality.Level) personality.Level {
	switch personality.Level(strings.ToLower(strings.TrimSpace(raw))) {
	case personality.LevelNone, personality.LevelLow, personality.LevelMedium, personality.LevelHigh:
		return personality.Level(strings.ToLower(strings.TrimSpace(raw)))
	default:
		return fallback
	}
}
