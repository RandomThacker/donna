package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/automationcatalog"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// AutomationService orchestrates user automations.
type AutomationService struct {
	autos repository.AutomationRepository
	now   func() time.Time
}

// NewAutomationService constructs an AutomationService.
func NewAutomationService(autos repository.AutomationRepository) *AutomationService {
	return &AutomationService{autos: autos, now: time.Now}
}

// CreateAutomationInput creates an automation.
type CreateAutomationInput struct {
	Name             string
	Description      *string
	Enabled          *bool
	TriggerType      string
	TriggerTime      string
	Timezone         string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
	TemplateID       *string
}

// UpdateAutomationInput patches an automation.
type UpdateAutomationInput struct {
	Name             *string
	Description      *string
	Enabled          *bool
	TriggerType      *string
	TriggerTime      *string
	Timezone         *string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
}

// List returns live automations for the user.
func (s *AutomationService) List(ctx context.Context, userID uuid.UUID) ([]entity.Automation, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.autos.ListByUser(ctx, userID)
}

// ListEnabled returns all enabled automations (scheduler scan).
func (s *AutomationService) ListEnabled(ctx context.Context) ([]entity.Automation, error) {
	return s.autos.ListEnabled(ctx)
}

// ListTemplates returns Intent Catalog templates.
func (s *AutomationService) ListTemplates() ([]automationcatalog.Template, error) {
	return automationcatalog.Load()
}

// Create adds an automation. Optional template_id fills name/commands/defaults.
func (s *AutomationService) Create(ctx context.Context, userID uuid.UUID, in CreateAutomationInput) (entity.Automation, error) {
	if userID == uuid.Nil {
		return entity.Automation{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}

	name := strings.TrimSpace(in.Name)
	description := trimOptional(in.Description)
	commands, err := NormalizeAutomationCommands(in.Commands)
	if err != nil {
		return entity.Automation{}, err
	}
	triggerType := strings.TrimSpace(in.TriggerType)
	triggerTime := strings.TrimSpace(in.TriggerTime)
	tz := strings.TrimSpace(in.Timezone)
	var templateID *string

	if in.TemplateID != nil {
		tid := strings.TrimSpace(*in.TemplateID)
		if tid != "" && tid != "custom" {
			tmpl, ok := automationcatalog.Get(tid)
			if !ok {
				return entity.Automation{}, fmt.Errorf("%w: unknown template_id", apperr.ErrValidation)
			}
			templateID = &tid
			if name == "" {
				name = tmpl.Name
			}
			if description == nil && strings.TrimSpace(tmpl.Description) != "" {
				d := tmpl.Description
				description = &d
			}
			if len(commands) == 0 {
				commands = append([]entity.AutomationCommand{}, tmpl.Commands...)
			}
			if triggerType == "" {
				triggerType = tmpl.DefaultSchedule.Type
			}
			if triggerTime == "" {
				triggerTime = tmpl.DefaultSchedule.Time
			}
		} else if tid == "custom" {
			templateID = &tid
		}
	}

	if name == "" {
		return entity.Automation{}, fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	if len(commands) == 0 {
		return entity.Automation{}, fmt.Errorf("%w: at least one command is required", apperr.ErrValidation)
	}
	commands, err = NormalizeAutomationCommands(commands)
	if err != nil {
		return entity.Automation{}, err
	}
	if triggerType == "" {
		triggerType = constant.AutomationTriggerDaily
	}
	if triggerType != constant.AutomationTriggerDaily {
		return entity.Automation{}, fmt.Errorf("%w: only daily triggers are supported", apperr.ErrValidation)
	}
	normalizedTime, err := normalizeLocalTime(triggerTime)
	if err != nil {
		return entity.Automation{}, fmt.Errorf("%w: trigger time must be HH:MM", apperr.ErrValidation)
	}
	if tz == "" {
		tz = "UTC"
	}
	if _, err := loadTimezone(tz); err != nil {
		return entity.Automation{}, fmt.Errorf("%w: invalid timezone", apperr.ErrValidation)
	}
	delivery, err := normalizeDeliveryChannels(in.DeliveryChannels)
	if err != nil {
		return entity.Automation{}, err
	}

	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}

	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.Automation{}, err
	}
	now := s.now().UTC()
	next := NextDailyRunAt(now, tz, normalizedTime)
	return s.autos.Create(ctx, entity.Automation{
		ID:               id,
		PublicID:         idgen.PublicID(constant.PublicIDPrefixAutomation, id),
		UserID:           userID,
		Name:             name,
		Description:      description,
		Enabled:          enabled,
		TriggerType:      triggerType,
		TriggerTime:      normalizedTime,
		Timezone:         tz,
		Commands:         commands,
		DeliveryChannels: delivery,
		TemplateID:       templateID,
		NextRunAt:        &next,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

// Update patches an automation owned by the user.
func (s *AutomationService) Update(ctx context.Context, userID, autoID uuid.UUID, in UpdateAutomationInput) (entity.Automation, error) {
	if userID == uuid.Nil || autoID == uuid.Nil {
		return entity.Automation{}, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	existing, err := s.autos.GetByID(ctx, autoID)
	if err != nil {
		return entity.Automation{}, err
	}
	if existing.UserID != userID {
		return entity.Automation{}, apperr.ErrForbidden
	}

	fields := repository.AutomationUpdateFields{
		Enabled: in.Enabled,
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return entity.Automation{}, fmt.Errorf("%w: name is required", apperr.ErrValidation)
		}
		fields.Name = &name
	}
	if in.Description != nil {
		fields.Description = trimOptional(in.Description)
	}
	if in.TriggerType != nil {
		tt := strings.TrimSpace(*in.TriggerType)
		if tt != constant.AutomationTriggerDaily {
			return entity.Automation{}, fmt.Errorf("%w: only daily triggers are supported", apperr.ErrValidation)
		}
		fields.TriggerType = &tt
	}
	tz := existing.Timezone
	triggerTime := existing.TriggerTime
	if in.Timezone != nil {
		tz = strings.TrimSpace(*in.Timezone)
		if tz == "" {
			return entity.Automation{}, fmt.Errorf("%w: timezone is required", apperr.ErrValidation)
		}
		if _, err := loadTimezone(tz); err != nil {
			return entity.Automation{}, fmt.Errorf("%w: invalid timezone", apperr.ErrValidation)
		}
		fields.Timezone = &tz
	}
	if in.TriggerTime != nil {
		normalized, err := normalizeLocalTime(*in.TriggerTime)
		if err != nil {
			return entity.Automation{}, fmt.Errorf("%w: trigger time must be HH:MM", apperr.ErrValidation)
		}
		triggerTime = normalized
		fields.TriggerTime = &normalized
	}
	if in.Commands != nil {
		commands, err := NormalizeAutomationCommands(in.Commands)
		if err != nil {
			return entity.Automation{}, err
		}
		if len(commands) == 0 {
			return entity.Automation{}, fmt.Errorf("%w: at least one command is required", apperr.ErrValidation)
		}
		fields.Commands = commands
	}
	if in.DeliveryChannels != nil {
		delivery, err := normalizeDeliveryChannels(in.DeliveryChannels)
		if err != nil {
			return entity.Automation{}, err
		}
		fields.DeliveryChannels = delivery
	}

	next := NextDailyRunAt(s.now().UTC(), tz, triggerTime)
	fields.NextRunAt = &next
	return s.autos.Update(ctx, autoID, userID, fields, s.now().UTC())
}

// GetOwned returns an automation owned by the user.
func (s *AutomationService) GetOwned(ctx context.Context, userID, autoID uuid.UUID) (entity.Automation, error) {
	if userID == uuid.Nil || autoID == uuid.Nil {
		return entity.Automation{}, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	existing, err := s.autos.GetByID(ctx, autoID)
	if err != nil {
		return entity.Automation{}, err
	}
	if existing.UserID != userID {
		return entity.Automation{}, apperr.ErrForbidden
	}
	return existing, nil
}

// Delete soft-deletes an automation owned by the user.
func (s *AutomationService) Delete(ctx context.Context, userID, autoID uuid.UUID) error {
	if userID == uuid.Nil || autoID == uuid.Nil {
		return fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	existing, err := s.autos.GetByID(ctx, autoID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.autos.SoftDelete(ctx, autoID, userID, s.now().UTC())
}

// MarkRun records a successful run and the next scheduled time.
func (s *AutomationService) MarkRun(ctx context.Context, autoID uuid.UUID, ranAt time.Time, nextRunAt *time.Time) (entity.Automation, error) {
	if autoID == uuid.Nil {
		return entity.Automation{}, fmt.Errorf("%w: automation id is required", apperr.ErrValidation)
	}
	if ranAt.IsZero() {
		ranAt = s.now().UTC()
	}
	return s.autos.MarkRun(ctx, autoID, ranAt.UTC(), nextRunAt)
}

// AutomationDue reports whether the automation should run at now (minute granularity, once per civil day).
func AutomationDue(auto entity.Automation, now time.Time) bool {
	if !auto.Enabled {
		return false
	}
	if auto.TriggerType != constant.AutomationTriggerDaily {
		return false
	}
	loc, err := time.LoadLocation(strings.TrimSpace(auto.Timezone))
	if err != nil || loc == nil {
		loc = time.UTC
	}
	local := now.In(loc)
	if local.Format("15:04") != normalizeLocalTimeOrEmpty(auto.TriggerTime) {
		return false
	}
	if auto.LastRunAt != nil {
		lastLocal := auto.LastRunAt.In(loc)
		if sameCivilDay(lastLocal, local) {
			return false
		}
	}
	return true
}

// ClientMessageIDForAutomationRun builds the idempotent client_message_id for a run.
func ClientMessageIDForAutomationRun(publicID string, localDay time.Time) string {
	return fmt.Sprintf("automation:%s:%s", publicID, localDay.Format("2006-01-02"))
}

// NextDailyRunAt returns the next daily trigger instant in UTC.
func NextDailyRunAt(now time.Time, timezone, triggerTime string) time.Time {
	loc, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil || loc == nil {
		loc = time.UTC
	}
	local := now.In(loc)
	hhmm := normalizeLocalTimeOrEmpty(triggerTime)
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		t, _ = time.Parse("15:04", "09:00")
	}
	candidate := time.Date(local.Year(), local.Month(), local.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	if !candidate.After(local) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate.UTC()
}

func normalizeDeliveryChannels(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return []string{constant.AutomationDeliveryChat}, nil
	}
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, ch := range raw {
		c := strings.ToLower(strings.TrimSpace(ch))
		if c == "" {
			continue
		}
		if _, ok := constant.AllowedAutomationDeliveryChannels[c]; !ok {
			return nil, fmt.Errorf("%w: delivery channel %q is not supported yet", apperr.ErrValidation, c)
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	if len(out) == 0 {
		return []string{constant.AutomationDeliveryChat}, nil
	}
	return out, nil
}

func normalizeLocalTime(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%w: local_time is required", apperr.ErrValidation)
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.Format("15:04"), nil
		}
	}
	return "", fmt.Errorf("%w: local_time must be HH:MM", apperr.ErrValidation)
}

func normalizeLocalTimeOrEmpty(raw string) string {
	out, err := normalizeLocalTime(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return out
}

func trimOptional(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
