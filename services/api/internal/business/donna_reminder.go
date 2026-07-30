package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// DonnaReminderService orchestrates Donna-owned timeline reminders.
type DonnaReminderService struct {
	reminders repository.DonnaReminderRepository
	now       func() time.Time
}

// NewDonnaReminderService constructs a DonnaReminderService.
func NewDonnaReminderService(reminders repository.DonnaReminderRepository) *DonnaReminderService {
	return &DonnaReminderService{reminders: reminders, now: time.Now}
}

// CreateDonnaReminderInput creates a Donna reminder.
type CreateDonnaReminderInput struct {
	Title          string
	Description    *string
	TriggerAt      time.Time
	Timezone       string
	RecurrenceRule *string
	Color          *string
}

// UpdateDonnaReminderInput patches a Donna reminder.
type UpdateDonnaReminderInput struct {
	Title          *string
	Description    *string
	TriggerAt      *time.Time
	Timezone       *string
	RecurrenceRule *string
	Status         *string
	Color          *string
}

// List returns live Donna reminders for the user.
func (s *DonnaReminderService) List(ctx context.Context, userID uuid.UUID) ([]entity.DonnaReminder, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.reminders.ListByUser(ctx, userID)
}

// ListInRange returns Donna reminders with trigger_at in [from, to).
func (s *DonnaReminderService) ListInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.DonnaReminder, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("%w: valid from/to range is required", apperr.ErrValidation)
	}
	return s.reminders.ListByUserInRange(ctx, userID, from.UTC(), to.UTC())
}

// Create adds a Donna reminder.
func (s *DonnaReminderService) Create(ctx context.Context, userID uuid.UUID, in CreateDonnaReminderInput) (entity.DonnaReminder, error) {
	if userID == uuid.Nil {
		return entity.DonnaReminder{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return entity.DonnaReminder{}, fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}
	if in.TriggerAt.IsZero() {
		return entity.DonnaReminder{}, fmt.Errorf("%w: trigger_at is required", apperr.ErrValidation)
	}
	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	if _, err := loadTimezone(tz); err != nil {
		return entity.DonnaReminder{}, err
	}
	rule, err := ValidateRecurrenceRule(in.RecurrenceRule)
	if err != nil {
		return entity.DonnaReminder{}, err
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.DonnaReminder{}, err
	}
	now := s.now().UTC()
	return s.reminders.Create(ctx, entity.DonnaReminder{
		ID:             id,
		PublicID:       idgen.PublicID(constant.PublicIDPrefixDonnaReminder, id),
		UserID:         userID,
		Title:          title,
		Description:    trimPtr(in.Description),
		TriggerAt:      in.TriggerAt.UTC(),
		Timezone:       tz,
		RecurrenceRule: rule,
		Status:         constant.DonnaReminderStatusScheduled,
		Color:          trimPtr(in.Color),
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

// Update patches a Donna reminder owned by the user.
func (s *DonnaReminderService) Update(ctx context.Context, userID, reminderID uuid.UUID, in UpdateDonnaReminderInput) (entity.DonnaReminder, error) {
	if userID == uuid.Nil || reminderID == uuid.Nil {
		return entity.DonnaReminder{}, fmt.Errorf("%w: user and reminder id are required", apperr.ErrValidation)
	}
	existing, err := s.reminders.GetByID(ctx, reminderID)
	if err != nil {
		return entity.DonnaReminder{}, err
	}
	if existing.UserID != userID {
		return entity.DonnaReminder{}, apperr.ErrForbidden
	}
	fields := repository.DonnaReminderUpdateFields{
		Description: trimPtr(in.Description),
		Color:       trimPtr(in.Color),
	}
	if in.RecurrenceRule != nil {
		rule, err := ValidateRecurrenceRule(in.RecurrenceRule)
		if err != nil {
			return entity.DonnaReminder{}, err
		}
		if rule == nil {
			empty := ""
			fields.RecurrenceRule = &empty
		} else {
			fields.RecurrenceRule = rule
		}
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return entity.DonnaReminder{}, fmt.Errorf("%w: title is required", apperr.ErrValidation)
		}
		fields.Title = &title
	}
	if in.TriggerAt != nil {
		utc := in.TriggerAt.UTC()
		fields.TriggerAt = &utc
	}
	if in.Timezone != nil {
		tz := strings.TrimSpace(*in.Timezone)
		if tz == "" {
			tz = "UTC"
		}
		if _, err := loadTimezone(tz); err != nil {
			return entity.DonnaReminder{}, err
		}
		fields.Timezone = &tz
	}
	if in.Status != nil {
		status := strings.TrimSpace(*in.Status)
		if status != constant.DonnaReminderStatusScheduled && status != constant.DonnaReminderStatusCancelled {
			return entity.DonnaReminder{}, fmt.Errorf("%w: invalid status", apperr.ErrValidation)
		}
		fields.Status = &status
	}
	return s.reminders.Update(ctx, reminderID, userID, fields, s.now().UTC())
}

// Delete soft-deletes a Donna reminder owned by the user.
func (s *DonnaReminderService) Delete(ctx context.Context, userID, reminderID uuid.UUID) error {
	if userID == uuid.Nil || reminderID == uuid.Nil {
		return fmt.Errorf("%w: user and reminder id are required", apperr.ErrValidation)
	}
	existing, err := s.reminders.GetByID(ctx, reminderID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.reminders.SoftDelete(ctx, reminderID, userID, s.now().UTC())
}
