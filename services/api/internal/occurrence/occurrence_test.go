package occurrence

import (
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

func TestOccurrenceTypeValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		t    OccurrenceType
		want bool
	}{
		{TypeEvent, true},
		{TypeReminder, true},
		{OccurrenceType("TASK"), false},
		{OccurrenceType(""), false},
	}
	for _, tc := range cases {
		if got := tc.t.Valid(); got != tc.want {
			t.Fatalf("%q Valid() = %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestOccurrenceSourceValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    OccurrenceSource
		want bool
	}{
		{SourceGoogle, true},
		{SourceMicrosoftICS, true},
		{SourceDonna, true},
		{OccurrenceSource("OUTLOOK"), false},
		{OccurrenceSource(""), false},
	}
	for _, tc := range cases {
		if got := tc.s.Valid(); got != tc.want {
			t.Fatalf("%q Valid() = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestOccurrenceStatusValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    OccurrenceStatus
		want bool
	}{
		{StatusActive, true},
		{StatusCompleted, true},
		{StatusCancelled, true},
		{StatusMissed, true},
		{OccurrenceStatus("PENDING"), false},
	}
	for _, tc := range cases {
		if got := tc.s.Valid(); got != tc.want {
			t.Fatalf("%q Valid() = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func validOccurrence() Occurrence {
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	desc := "Stretch break"
	parent := "parent_1"
	rrule := "FREQ=WEEKLY;BYDAY=FR"
	return Occurrence{
		ID:             "occ_row_1",
		ParentID:       &parent,
		OccurrenceID:   "occ_20260801T100000Z",
		UserID:         uuid.MustParse("018f0000-0000-7000-8000-000000000401"),
		Source:         SourceDonna,
		Type:           TypeReminder,
		Title:          "Stretch",
		Description:    &desc,
		StartAt:        start,
		EndAt:          start.Add(15 * time.Minute),
		Timezone:       "UTC",
		RecurrenceRule: &rrule,
		Status:         StatusActive,
		Metadata:       map[string]any{"priority": "normal"},
	}
}

func TestOccurrenceValidateOK(t *testing.T) {
	t.Parallel()
	o := validOccurrence()
	if err := o.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestOccurrenceValidateRequiredFields(t *testing.T) {
	t.Parallel()
	base := validOccurrence()

	cases := []struct {
		name string
		mut  func(*Occurrence)
	}{
		{"empty id", func(o *Occurrence) { o.ID = "  " }},
		{"empty occurrence id", func(o *Occurrence) { o.OccurrenceID = "" }},
		{"nil user", func(o *Occurrence) { o.UserID = uuid.Nil }},
		{"bad source", func(o *Occurrence) { o.Source = "X" }},
		{"bad type", func(o *Occurrence) { o.Type = "TASK" }},
		{"empty title", func(o *Occurrence) { o.Title = "" }},
		{"zero start", func(o *Occurrence) { o.StartAt = time.Time{} }},
		{"zero end", func(o *Occurrence) { o.EndAt = time.Time{} }},
		{"end before start", func(o *Occurrence) { o.EndAt = o.StartAt.Add(-time.Minute) }},
		{"empty timezone", func(o *Occurrence) { o.Timezone = " " }},
		{"bad status", func(o *Occurrence) { o.Status = "PENDING" }},
		{"empty parent ptr", func(o *Occurrence) { empty := "  "; o.ParentID = &empty }},
		{"empty rrule ptr", func(o *Occurrence) { empty := ""; o.RecurrenceRule = &empty }},
		{"empty description ptr", func(o *Occurrence) { empty := " "; o.Description = &empty }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o := base
			tc.mut(&o)
			err := o.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
		})
	}
}

func TestOccurrenceNormalize(t *testing.T) {
	t.Parallel()
	empty := "  "
	o := validOccurrence()
	o.ID = "  id  "
	o.Title = "  Hello  "
	o.Timezone = " Asia/Kolkata "
	o.Description = &empty
	o.Metadata = map[string]any{}

	n := o.Normalize()
	if n.ID != "id" || n.Title != "Hello" || n.Timezone != "Asia/Kolkata" {
		t.Fatalf("normalize strings = %+v", n)
	}
	if n.Description != nil {
		t.Fatalf("empty description should become nil, got %q", *n.Description)
	}
	if n.Metadata != nil {
		t.Fatalf("empty metadata should become nil")
	}
	if err := n.Validate(); err != nil {
		t.Fatalf("normalized should validate: %v", err)
	}
}

func TestOccurrenceDuration(t *testing.T) {
	t.Parallel()
	o := validOccurrence()
	if got := o.Duration(); got != 15*time.Minute {
		t.Fatalf("Duration() = %v", got)
	}
	o.EndAt = o.StartAt.Add(-time.Second)
	if o.Duration() != 0 {
		t.Fatal("inverted range should return 0")
	}
}
