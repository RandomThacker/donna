package provider

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

type countingCalendarEvents struct {
	memCalendarEvents
	calls atomic.Int32
}

func (c *countingCalendarEvents) ListForSchedulerByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	c.calls.Add(1)
	return c.memCalendarEvents.ListForSchedulerByUserInRange(ctx, userID, from, to, providers)
}

func (c *countingCalendarEvents) ListCalendarOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	return c.ListForSchedulerByUserInRange(ctx, userID, from, to, providers)
}

func (c *countingCalendarEvents) ListByUserInRangeWithProvider(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.CalendarEventWithProvider, error) {
	c.calls.Add(1)
	return c.memCalendarEvents.ListByUserInRangeWithProvider(ctx, userID, from, to)
}

func TestSharedCalendarSingleQuery(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000701")
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	start := from.Add(10 * time.Hour)

	repo := &countingCalendarEvents{memCalendarEvents: memCalendarEvents{rows: []entity.CalendarEventWithProvider{
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000702"),
				UserID: userID, Title: "G", PublicID: "evt_g",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status:          constant.CalendarEventStatusConfirmed,
				ProviderPayload: []byte(`{"x":1}`),
			},
			Provider: constant.AuthProviderGoogle,
		},
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000703"),
				UserID: userID, Title: "M", PublicID: "evt_m",
				StartsAt: start.Add(time.Hour), EndsAt: start.Add(2 * time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderMicrosoft,
		},
	}}}

	// Before: Google narrow + Microsoft wide = 2 queries.
	google := NewGoogleOccurrenceProvider(repo, nil)
	ms := NewMicrosoftICSOccurrenceProvider(repo)
	if _, err := google.ListOccurrences(context.Background(), userID, from, to); err != nil {
		t.Fatal(err)
	}
	if _, err := ms.ListOccurrences(context.Background(), userID, from, to); err != nil {
		t.Fatal(err)
	}
	beforeCalls := repo.calls.Load()
	if beforeCalls != 2 {
		t.Fatalf("before queries = %d, want 2", beforeCalls)
	}

	repo.calls.Store(0)
	shared := NewSharedCalendarOccurrenceProvider(repo, ActiveCalendarOccurrenceProviders, nil)
	got, err := shared.ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	afterCalls := repo.calls.Load()
	if afterCalls != 1 {
		t.Fatalf("after queries = %d, want 1", afterCalls)
	}
	if len(got) != 1 || got[0].Title != "G" {
		t.Fatalf("active providers should return google only, got %#v", got)
	}
}

func TestSharedCalendarMatchesGoogleProviderOutput(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000704")
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	start := from.Add(11 * time.Hour)
	tz := "UTC"
	pid := "g-1"

	repo := &memCalendarEvents{rows: []entity.CalendarEventWithProvider{{
		Event: entity.CalendarEvent{
			ID: uuid.MustParse("018f0000-0000-7000-8000-000000000705"),
			PublicID: "evt_1", UserID: userID, Title: "Sync",
			StartsAt: start, EndsAt: start.Add(30 * time.Minute),
			Status: constant.CalendarEventStatusConfirmed, Timezone: &tz,
			ProviderEventID: &pid, Origin: "provider_sync",
			ProviderPayload: []byte(`{"ignored":true}`),
		},
		Provider: constant.AuthProviderGoogle,
	}}}

	legacy, err := NewGoogleOccurrenceProvider(repo, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := NewSharedCalendarOccurrenceProvider(repo, []string{constant.AuthProviderGoogle}, nil).
		ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) != 1 || len(shared) != 1 {
		t.Fatalf("len legacy=%d shared=%d", len(legacy), len(shared))
	}
	assertOccurrenceEqual(t, legacy[0], shared[0])
}

func TestSharedCalendarMultiProviderFilterInSQL(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000706")
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	start := from.Add(9 * time.Hour)

	repo := &countingCalendarEvents{memCalendarEvents: memCalendarEvents{rows: []entity.CalendarEventWithProvider{
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000707"),
				UserID: userID, Title: "G", PublicID: "evt_g",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderGoogle,
		},
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000708"),
				UserID: userID, Title: "M", PublicID: "evt_m",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderMicrosoft,
		},
		{
			Event: entity.CalendarEvent{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000709"),
				UserID: userID, Title: "I", PublicID: "evt_i",
				StartsAt: start, EndsAt: start.Add(time.Hour),
				Status: constant.CalendarEventStatusConfirmed,
			},
			Provider: constant.AuthProviderICS,
		},
	}}}

	got, err := NewSharedCalendarOccurrenceProvider(repo, []string{
		constant.AuthProviderGoogle,
		constant.AuthProviderMicrosoft,
		constant.AuthProviderICS,
	}, nil).ListOccurrences(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if repo.calls.Load() != 1 {
		t.Fatalf("queries = %d", repo.calls.Load())
	}
	if len(got) != 3 {
		t.Fatalf("got %d rows", len(got))
	}
	sources := map[occurrence.OccurrenceSource]int{}
	for _, o := range got {
		sources[o.Source]++
	}
	if sources[occurrence.SourceGoogle] != 1 || sources[occurrence.SourceMicrosoftICS] != 2 {
		t.Fatalf("sources = %#v", sources)
	}
}

func TestSharedCalendarDefaultsToActiveProviders(t *testing.T) {
	t.Parallel()
	p := NewSharedCalendarOccurrenceProvider(&memCalendarEvents{}, nil, nil)
	if len(p.providers) != 1 || p.providers[0] != constant.AuthProviderGoogle {
		t.Fatalf("providers = %#v", p.providers)
	}
}
