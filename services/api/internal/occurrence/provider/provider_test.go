package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

func TestGoogleOccurrenceProviderOneTime(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000501")
	eventID := uuid.MustParse("018f0000-0000-7000-8000-000000000502")
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	repo := &memCalendarEvents{rows: []entity.CalendarEventWithProvider{{
		Event: entity.CalendarEvent{
			ID: eventID, UserID: userID, Title: "Standup",
			StartsAt: start, EndsAt: end,
			Status: constant.CalendarEventStatusConfirmed,
		},
		Provider: constant.AuthProviderGoogle,
	}}}

	got, err := NewGoogleOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	o := got[0]
	if o.Source != occurrence.SourceGoogle || o.Type != occurrence.TypeEvent {
		t.Fatalf("source/type = %s/%s", o.Source, o.Type)
	}
	if o.Title != "Standup" || o.OccurrenceID != eventID.String() {
		t.Fatalf("occurrence = %+v", o)
	}
	if o.Metadata != nil {
		if _, ok := o.Metadata["color"]; ok {
			t.Fatal("must not include color metadata")
		}
	}
}

func TestGoogleOccurrenceProviderEmptyAndFiltersMicrosoft(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000503")
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	repo := &memCalendarEvents{rows: []entity.CalendarEventWithProvider{{
		Event: entity.CalendarEvent{
			ID: uuid.MustParse("018f0000-0000-7000-8000-000000000504"),
			UserID: userID, Title: "Outlook",
			StartsAt: start, EndsAt: start.Add(time.Hour),
			Status: constant.CalendarEventStatusConfirmed,
		},
		Provider: constant.AuthProviderMicrosoft,
	}}}

	got, err := NewGoogleOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("google provider should ignore microsoft rows, got %d", len(got))
	}

	empty, err := NewGoogleOccurrenceProvider(&memCalendarEvents{}, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty = %v %v", empty, err)
	}
}

func TestMicrosoftICSOccurrenceProvider(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000505")
	start := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	repo := &memCalendarEvents{rows: []entity.CalendarEventWithProvider{
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000506"),
				UserID: userID, Title: "ICS",
				StartsAt: start, EndsAt: start.Add(30 * time.Minute),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderICS,
		},
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000507"),
				UserID: userID, Title: "Google",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderGoogle,
		},
	}}

	got, err := NewMicrosoftICSOccurrenceProvider(repo).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Source != occurrence.SourceMicrosoftICS {
		t.Fatalf("got = %+v", got)
	}
}

func TestDonnaEventOccurrenceProviderOneTimeAndRange(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000508")
	start := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(45 * time.Minute)
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	repo := &memDonnaEvents{rows: []entity.DonnaEvent{{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000509"),
		UserID: userID, Title: "Focus",
		StartAt: start, EndAt: end, Timezone: "UTC",
		Status: constant.DonnaEventStatusConfirmed,
	}}}

	got, err := NewDonnaEventOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].Type != occurrence.TypeEvent || got[0].Source != occurrence.SourceDonna {
		t.Fatalf("got = %+v", got[0])
	}

	// Outside range → empty after expansion filter.
	out, err := NewDonnaEventOccurrenceProvider(repo, nil).ListOccurrences(
		context.Background(), userID,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("want empty outside range, got %d", len(out))
	}
}

func TestDonnaEventOccurrenceProviderRecurring(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000510")
	start := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	rule := "FREQ=DAILY"
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)

	repo := &memDonnaEvents{rows: []entity.DonnaEvent{{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000511"),
		UserID: userID, Title: "Daily standup",
		StartAt: start, EndAt: end, Timezone: "UTC",
		RecurrenceRule: &rule,
		Status:         constant.DonnaEventStatusConfirmed,
	}}}

	got, err := NewDonnaEventOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ParentID == nil {
		t.Fatal("expected ParentID on recurring occurrence")
	}
	if got[0].RecurrenceRule == nil || *got[0].RecurrenceRule != "FREQ=DAILY" {
		t.Fatalf("rule = %v", got[0].RecurrenceRule)
	}
}

func TestDonnaEventOccurrenceProviderInvalidRecurrence(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000512")
	start := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	rule := "FREQ=NEVER"
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	repo := &memDonnaEvents{rows: []entity.DonnaEvent{{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000513"),
		UserID: userID, Title: "Bad",
		StartAt: start, EndAt: start.Add(time.Hour), Timezone: "UTC",
		RecurrenceRule: &rule,
		Status:         constant.DonnaEventStatusConfirmed,
	}}}

	_, err := NewDonnaEventOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

func TestDonnaReminderOccurrenceProvider(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000514")
	trigger := time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	t.Run("one-time", func(t *testing.T) {
		t.Parallel()
		repo := &memDonnaReminders{rows: []entity.DonnaReminder{{
			ID: uuid.MustParse("018f0000-0000-7000-8000-000000000515"),
			UserID: userID, Title: "Stretch",
			TriggerAt: trigger, Timezone: "UTC",
			Status: constant.DonnaReminderStatusScheduled,
		}}}
		got, err := NewDonnaReminderOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Type != occurrence.TypeReminder {
			t.Fatalf("got = %+v", got)
		}
		if !got[0].StartAt.Equal(trigger) || !got[0].EndAt.Equal(trigger) {
			t.Fatalf("window = %v..%v", got[0].StartAt, got[0].EndAt)
		}
	})

	t.Run("recurring weekly", func(t *testing.T) {
		t.Parallel()
		rule := "FREQ=DAILY"
		repo := &memDonnaReminders{rows: []entity.DonnaReminder{{
			ID: uuid.MustParse("018f0000-0000-7000-8000-000000000516"),
			UserID: userID, Title: "Water",
			TriggerAt: trigger, Timezone: "UTC",
			RecurrenceRule: &rule,
			Status:         constant.DonnaReminderStatusScheduled,
		}}}
		windowTo := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
		got, err := NewDonnaReminderOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, windowTo)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Fatalf("len = %d", len(got))
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got, err := NewDonnaReminderOccurrenceProvider(&memDonnaReminders{}, nil).ListOccurrences(context.Background(), userID, from, to)
		if err != nil || len(got) != 0 {
			t.Fatalf("got %v %v", got, err)
		}
	})
}

func TestNilReadersReturnEmpty(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000517")
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	if got, err := NewGoogleOccurrenceProvider(nil, nil).ListOccurrences(context.Background(), userID, from, to); err != nil || got != nil {
		t.Fatalf("google nil: %v %v", got, err)
	}
	if got, err := NewDonnaEventOccurrenceProvider(nil, nil).ListOccurrences(context.Background(), userID, from, to); err != nil || got != nil {
		t.Fatalf("donna event nil: %v %v", got, err)
	}
	if got, err := NewDonnaReminderOccurrenceProvider(nil, nil).ListOccurrences(context.Background(), userID, from, to); err != nil || got != nil {
		t.Fatalf("donna reminder nil: %v %v", got, err)
	}
}
