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

// DonnaEventService orchestrates Donna-owned timeline events.
type DonnaEventService struct {
	events repository.DonnaEventRepository
	now    func() time.Time
}

// NewDonnaEventService constructs a DonnaEventService.
func NewDonnaEventService(events repository.DonnaEventRepository) *DonnaEventService {
	return &DonnaEventService{events: events, now: time.Now}
}

// CreateDonnaEventInput creates a Donna event.
type CreateDonnaEventInput struct {
	Title                 string
	Description           *string
	StartAt               time.Time
	EndAt                 time.Time
	Timezone              string
	AllDay                bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Color                 *string
}

// UpdateDonnaEventInput patches a Donna event.
type UpdateDonnaEventInput struct {
	Title                 *string
	Description           *string
	StartAt               *time.Time
	EndAt                 *time.Time
	Timezone              *string
	AllDay                *bool
	Location              *string
	ReminderOffsetMinutes *int
	RecurrenceRule        *string
	Status                *string
	Color                 *string
}

// List returns live Donna events for the user.
func (s *DonnaEventService) List(ctx context.Context, userID uuid.UUID) ([]entity.DonnaEvent, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.events.ListByUser(ctx, userID)
}

// ListInRange returns Donna events overlapping [from, to).
func (s *DonnaEventService) ListInRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.DonnaEvent, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("%w: valid from/to range is required", apperr.ErrValidation)
	}
	return s.events.ListByUserInRange(ctx, userID, from.UTC(), to.UTC())
}

// Create adds a Donna event.
func (s *DonnaEventService) Create(ctx context.Context, userID uuid.UUID, in CreateDonnaEventInput) (entity.DonnaEvent, error) {
	if userID == uuid.Nil {
		return entity.DonnaEvent{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return entity.DonnaEvent{}, fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}
	if in.StartAt.IsZero() || in.EndAt.IsZero() {
		return entity.DonnaEvent{}, fmt.Errorf("%w: start_at and end_at are required", apperr.ErrValidation)
	}
	if in.EndAt.Before(in.StartAt) {
		return entity.DonnaEvent{}, fmt.Errorf("%w: end_at must be on or after start_at", apperr.ErrValidation)
	}
	if in.ReminderOffsetMinutes != nil && *in.ReminderOffsetMinutes < 0 {
		return entity.DonnaEvent{}, fmt.Errorf("%w: reminder_offset_minutes must be >= 0", apperr.ErrValidation)
	}
	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	if _, err := loadTimezone(tz); err != nil {
		return entity.DonnaEvent{}, err
	}
	rule, err := ValidateRecurrenceRule(in.RecurrenceRule)
	if err != nil {
		return entity.DonnaEvent{}, err
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.DonnaEvent{}, err
	}
	now := s.now().UTC()
	return s.events.Create(ctx, entity.DonnaEvent{
		ID:                    id,
		PublicID:              idgen.PublicID(constant.PublicIDPrefixDonnaEvent, id),
		UserID:                userID,
		Title:                 title,
		Description:           trimPtr(in.Description),
		StartAt:               in.StartAt.UTC(),
		EndAt:                 in.EndAt.UTC(),
		Timezone:              tz,
		AllDay:                in.AllDay,
		Location:              trimPtr(in.Location),
		ReminderOffsetMinutes: in.ReminderOffsetMinutes,
		RecurrenceRule:        rule,
		Status:                constant.DonnaEventStatusConfirmed,
		Color:                 trimPtr(in.Color),
		CreatedAt:             now,
		UpdatedAt:             now,
	})
}

// Update patches a Donna event owned by the user.
func (s *DonnaEventService) Update(ctx context.Context, userID, eventID uuid.UUID, in UpdateDonnaEventInput) (entity.DonnaEvent, error) {
	if userID == uuid.Nil || eventID == uuid.Nil {
		return entity.DonnaEvent{}, fmt.Errorf("%w: user and event id are required", apperr.ErrValidation)
	}
	existing, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return entity.DonnaEvent{}, err
	}
	if existing.UserID != userID {
		return entity.DonnaEvent{}, apperr.ErrForbidden
	}
	fields := repository.DonnaEventUpdateFields{
		Description:           trimPtr(in.Description),
		StartAt:               in.StartAt,
		EndAt:                 in.EndAt,
		AllDay:                in.AllDay,
		Location:              trimPtr(in.Location),
		ReminderOffsetMinutes: in.ReminderOffsetMinutes,
		Color:                 trimPtr(in.Color),
	}
	if in.RecurrenceRule != nil {
		rule, err := ValidateRecurrenceRule(in.RecurrenceRule)
		if err != nil {
			return entity.DonnaEvent{}, err
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
			return entity.DonnaEvent{}, fmt.Errorf("%w: title is required", apperr.ErrValidation)
		}
		fields.Title = &title
	}
	if in.Timezone != nil {
		tz := strings.TrimSpace(*in.Timezone)
		if tz == "" {
			tz = "UTC"
		}
		if _, err := loadTimezone(tz); err != nil {
			return entity.DonnaEvent{}, err
		}
		fields.Timezone = &tz
	}
	if in.Status != nil {
		status := strings.TrimSpace(*in.Status)
		if status != constant.DonnaEventStatusConfirmed && status != constant.DonnaEventStatusCancelled {
			return entity.DonnaEvent{}, fmt.Errorf("%w: invalid status", apperr.ErrValidation)
		}
		fields.Status = &status
	}
	startAt := existing.StartAt
	endAt := existing.EndAt
	if fields.StartAt != nil {
		utc := fields.StartAt.UTC()
		fields.StartAt = &utc
		startAt = utc
	}
	if fields.EndAt != nil {
		utc := fields.EndAt.UTC()
		fields.EndAt = &utc
		endAt = utc
	}
	if endAt.Before(startAt) {
		return entity.DonnaEvent{}, fmt.Errorf("%w: end_at must be on or after start_at", apperr.ErrValidation)
	}
	return s.events.Update(ctx, eventID, userID, fields, s.now().UTC())
}

// Delete soft-deletes a Donna event owned by the user.
func (s *DonnaEventService) Delete(ctx context.Context, userID, eventID uuid.UUID) error {
	if userID == uuid.Nil || eventID == uuid.Nil {
		return fmt.Errorf("%w: user and event id are required", apperr.ErrValidation)
	}
	existing, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.events.SoftDelete(ctx, eventID, userID, s.now().UTC())
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	return &trimmed
}
