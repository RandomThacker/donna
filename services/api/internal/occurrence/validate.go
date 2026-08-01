package occurrence

import (
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// Validate checks that o is a usable scheduling Occurrence.
func (o Occurrence) Validate() error {
	if strings.TrimSpace(o.ID) == "" {
		return fmt.Errorf("%w: id is required", apperr.ErrValidation)
	}
	if strings.TrimSpace(o.OccurrenceID) == "" {
		return fmt.Errorf("%w: occurrence id is required", apperr.ErrValidation)
	}
	if o.UserID == uuid.Nil {
		return fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if !o.Source.Valid() {
		return fmt.Errorf("%w: source is invalid", apperr.ErrValidation)
	}
	if !o.Type.Valid() {
		return fmt.Errorf("%w: type is invalid", apperr.ErrValidation)
	}
	if strings.TrimSpace(o.Title) == "" {
		return fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}
	if o.StartAt.IsZero() {
		return fmt.Errorf("%w: start_at is required", apperr.ErrValidation)
	}
	if o.EndAt.IsZero() {
		return fmt.Errorf("%w: end_at is required", apperr.ErrValidation)
	}
	if o.EndAt.Before(o.StartAt) {
		return fmt.Errorf("%w: end_at must not be before start_at", apperr.ErrValidation)
	}
	if strings.TrimSpace(o.Timezone) == "" {
		return fmt.Errorf("%w: timezone is required", apperr.ErrValidation)
	}
	if !o.Status.Valid() {
		return fmt.Errorf("%w: status is invalid", apperr.ErrValidation)
	}
	if o.ParentID != nil && strings.TrimSpace(*o.ParentID) == "" {
		return fmt.Errorf("%w: parent_id must not be empty when set", apperr.ErrValidation)
	}
	if o.RecurrenceRule != nil && strings.TrimSpace(*o.RecurrenceRule) == "" {
		return fmt.Errorf("%w: recurrence_rule must not be empty when set", apperr.ErrValidation)
	}
	if o.Description != nil && strings.TrimSpace(*o.Description) == "" {
		return fmt.Errorf("%w: description must not be empty when set", apperr.ErrValidation)
	}
	return nil
}

// Normalize returns a copy with trimmed string fields. Optional empty pointers
// become nil. Does not mutate the receiver.
func (o Occurrence) Normalize() Occurrence {
	out := o
	out.ID = strings.TrimSpace(o.ID)
	out.OccurrenceID = strings.TrimSpace(o.OccurrenceID)
	out.Title = strings.TrimSpace(o.Title)
	out.Timezone = strings.TrimSpace(o.Timezone)
	out.ParentID = trimOptional(o.ParentID)
	out.Description = trimOptional(o.Description)
	out.RecurrenceRule = trimOptional(o.RecurrenceRule)
	if len(o.Metadata) == 0 {
		out.Metadata = nil
	}
	return out
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// Duration returns EndAt − StartAt. Zero if times are unset or inverted.
func (o Occurrence) Duration() time.Duration {
	if o.StartAt.IsZero() || o.EndAt.IsZero() || o.EndAt.Before(o.StartAt) {
		return 0
	}
	return o.EndAt.Sub(o.StartAt)
}
