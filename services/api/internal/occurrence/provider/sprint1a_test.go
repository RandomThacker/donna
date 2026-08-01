package provider

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

func TestSchedulerProjectionOmitsPayloadColumns(t *testing.T) {
	t.Parallel()
	for _, col := range repository.CalendarEventSchedulerColumnNames {
		if col == "provider_payload" || col == "attendees_summary" || col == "organizer_summary" {
			t.Fatalf("scheduler calendar projection must not include %s", col)
		}
	}
	for _, col := range repository.DonnaEventSchedulerColumnNames {
		if col == "color" || col == "created_at" || col == "all_day" {
			t.Fatalf("donna event scheduler projection must not include %s", col)
		}
	}
	for _, col := range repository.DonnaReminderSchedulerColumnNames {
		if col == "color" || col == "created_at" {
			t.Fatalf("donna reminder scheduler projection must not include %s", col)
		}
	}
}

func TestNarrowCalendarRowProducesIdenticalOccurrence(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000601")
	eventID := uuid.MustParse("018f0000-0000-7000-8000-000000000602")
	sourceID := uuid.MustParse("018f0000-0000-7000-8000-000000000603")
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	tz := "America/New_York"
	loc := "Room A"
	providerEventID := "gcal-abc"
	desc := "notes"

	wide := entity.CalendarEventWithProvider{
		Event: entity.CalendarEvent{
			ID: eventID, PublicID: "evt_x", UserID: userID, CalendarSourceID: sourceID,
			Title: "Standup", Description: &desc, Location: &loc,
			StartsAt: start, EndsAt: end, Status: constant.CalendarEventStatusConfirmed,
			Timezone: &tz, ProviderEventID: &providerEventID, Origin: "provider_sync",
			ProviderPayload:  []byte(`{"huge":true,"attendees":[1,2,3]}`),
			AttendeesSummary: []byte(`[{"email":"a@b.c"}]`),
			OrganizerSummary: []byte(`{"email":"o@b.c"}`),
			IsAllDay:         true,
			Visibility:       strPtr("private"),
		},
		Provider:    constant.AuthProviderGoogle,
		SourceColor: strPtr("#ff0000"),
	}
	narrow := entity.CalendarEventWithProvider{
		Event: entity.CalendarEvent{
			ID: eventID, PublicID: "evt_x", UserID: userID, CalendarSourceID: sourceID,
			Title: "Standup", Description: &desc, Location: &loc,
			StartsAt: start, EndsAt: end, Status: constant.CalendarEventStatusConfirmed,
			Timezone: &tz, ProviderEventID: &providerEventID, Origin: "provider_sync",
		},
		Provider: constant.AuthProviderGoogle,
	}

	assertOccurrenceEqual(t, mapCalendarEvent(wide), mapCalendarEvent(narrow))
}

func TestNarrowDonnaEventProducesIdenticalOccurrences(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000604")
	eventID := uuid.MustParse("018f0000-0000-7000-8000-000000000605")
	from := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	to := from.Add(3 * 24 * time.Hour)
	rule := "FREQ=DAILY;COUNT=2"
	color := "#abc"
	offset := 15
	loc := "Studio"

	wide := entity.DonnaEvent{
		ID: eventID, PublicID: "dev_1", UserID: userID, Title: "Class",
		StartAt: from.Add(10 * time.Hour), EndAt: from.Add(11 * time.Hour),
		Timezone: "UTC", AllDay: true, Location: &loc,
		ReminderOffsetMinutes: &offset, RecurrenceRule: &rule,
		Status: constant.DonnaEventStatusConfirmed, Color: &color,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	narrow := wide
	narrow.Color = nil
	narrow.AllDay = false
	narrow.CreatedAt = time.Time{}
	narrow.UpdatedAt = time.Time{}

	wideOut, err := expandDonnaEvent(wide, from, to)
	if err != nil {
		t.Fatal(err)
	}
	narrowOut, err := expandDonnaEvent(narrow, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(wideOut) != len(narrowOut) {
		t.Fatalf("len wide=%d narrow=%d", len(wideOut), len(narrowOut))
	}
	for i := range wideOut {
		assertOccurrenceEqual(t, wideOut[i], narrowOut[i])
	}
}

func TestNarrowDonnaReminderProducesIdenticalOccurrences(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000610")
	remID := uuid.MustParse("018f0000-0000-7000-8000-000000000611")
	from := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	to := from.Add(3 * 24 * time.Hour)
	rule := "FREQ=DAILY;COUNT=2"
	color := "#def"

	wide := entity.DonnaReminder{
		ID: remID, PublicID: "drm_1", UserID: userID, Title: "Water",
		TriggerAt: from.Add(9 * time.Hour), Timezone: "UTC",
		RecurrenceRule: &rule, Status: constant.DonnaReminderStatusScheduled,
		Color: &color, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	narrow := wide
	narrow.Color = nil
	narrow.CreatedAt = time.Time{}
	narrow.UpdatedAt = time.Time{}

	wideOut, err := expandDonnaReminder(wide, from, to)
	if err != nil {
		t.Fatal(err)
	}
	narrowOut, err := expandDonnaReminder(narrow, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(wideOut) != len(narrowOut) {
		t.Fatalf("len wide=%d narrow=%d", len(wideOut), len(narrowOut))
	}
	for i := range wideOut {
		assertOccurrenceEqual(t, wideOut[i], narrowOut[i])
	}
}

func TestGoogleSchedulerPathMatchesWideFilterBehavior(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000606")
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	start := from.Add(10 * time.Hour)

	repo := &memCalendarEvents{rows: []entity.CalendarEventWithProvider{
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000607"),
				UserID: userID, Title: "G", PublicID: "evt_g",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status:          constant.CalendarEventStatusConfirmed,
				ProviderPayload: []byte(`{"x":1}`),
			},
			Provider: constant.AuthProviderGoogle,
		},
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000608"),
				UserID: userID, Title: "M",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderMicrosoft,
		},
	}}

	wideRows, err := repo.ListByUserInRangeWithProvider(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	var legacy []occurrence.Occurrence
	for _, row := range wideRows {
		if row.Provider != constant.AuthProviderGoogle {
			continue
		}
		legacy = append(legacy, mapCalendarEvent(row))
	}

	optimized, err := NewGoogleOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) != 1 || len(optimized) != 1 {
		t.Fatalf("legacy=%d optimized=%d", len(legacy), len(optimized))
	}
	assertOccurrenceEqual(t, legacy[0], optimized[0])
}

func strPtr(s string) *string { return &s }

func assertOccurrenceEqual(t *testing.T, a, b occurrence.Occurrence) {
	t.Helper()
	if a.ID != b.ID || a.OccurrenceID != b.OccurrenceID {
		t.Fatalf("ids %s/%s vs %s/%s", a.ID, a.OccurrenceID, b.ID, b.OccurrenceID)
	}
	if a.Source != b.Source || a.Type != b.Type || a.Status != b.Status {
		t.Fatalf("source/type/status mismatch")
	}
	if a.Title != b.Title || a.Timezone != b.Timezone {
		t.Fatalf("title/tz mismatch")
	}
	if !a.StartAt.Equal(b.StartAt) || !a.EndAt.Equal(b.EndAt) {
		t.Fatalf("times mismatch")
	}
	if (a.ParentID == nil) != (b.ParentID == nil) {
		t.Fatalf("parent nil mismatch")
	}
	if a.ParentID != nil && *a.ParentID != *b.ParentID {
		t.Fatalf("parent %s vs %s", *a.ParentID, *b.ParentID)
	}
	if (a.RecurrenceRule == nil) != (b.RecurrenceRule == nil) {
		t.Fatalf("rrule nil mismatch")
	}
	if a.RecurrenceRule != nil && *a.RecurrenceRule != *b.RecurrenceRule {
		t.Fatalf("rrule mismatch")
	}
	if !reflect.DeepEqual(a.Description, b.Description) {
		t.Fatalf("description mismatch")
	}
	if !reflect.DeepEqual(a.Metadata, b.Metadata) {
		t.Fatalf("metadata %v vs %v", a.Metadata, b.Metadata)
	}
}
