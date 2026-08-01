package occurrence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

type stubProvider struct {
	items []Occurrence
	err   error
}

func (s stubProvider) ListOccurrences(
	_ context.Context,
	_ uuid.UUID,
	_, _ time.Time,
) ([]Occurrence, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]Occurrence, len(s.items))
	copy(out, s.items)
	return out, nil
}

var _ Provider = stubProvider{}

func testUser() uuid.UUID {
	return uuid.MustParse("01900000-0000-7000-8000-000000000401")
}

func baseOcc(id string, start time.Time, source OccurrenceSource) Occurrence {
	return Occurrence{
		ID:           id,
		OccurrenceID: id,
		UserID:       testUser(),
		Source:       source,
		Type:         TypeEvent,
		Title:        id,
		StartAt:      start,
		EndAt:        start.Add(time.Hour),
		Timezone:     "UTC",
		Status:       StatusActive,
	}
}

func TestListUpcomingNoProviders(t *testing.T) {
	svc := NewService(ServiceDeps{})
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d items, want 0", len(got))
	}
}

func TestListUpcomingOneProvider(t *testing.T) {
	from := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	item := baseOcc("evt-1", from.Add(time.Hour), SourceGoogle)

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{item}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 1 || got[0].OccurrenceID != "evt-1" {
		t.Fatalf("got %#v", got)
	}
}

func TestListUpcomingMultipleProviders(t *testing.T) {
	from := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{baseOcc("g", from.Add(2*time.Hour), SourceGoogle)}},
			stubProvider{items: []Occurrence{baseOcc("d", from.Add(time.Hour), SourceDonna)}},
			nil,
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	if got[0].OccurrenceID != "d" || got[1].OccurrenceID != "g" {
		t.Fatalf("sort order = [%s, %s], want [d, g]", got[0].OccurrenceID, got[1].OccurrenceID)
	}
}

func TestListUpcomingSortingBySourcePriority(t *testing.T) {
	from := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	start := from.Add(time.Hour)

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{
				baseOcc("donna", start, SourceDonna),
				baseOcc("ms", start, SourceMicrosoftICS),
				baseOcc("google", start, SourceGoogle),
			}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d", len(got))
	}
	want := []string{"google", "ms", "donna"}
	for i, id := range want {
		if got[i].OccurrenceID != id {
			t.Fatalf("index %d = %s, want %s", i, got[i].OccurrenceID, id)
		}
	}
}

func TestListUpcomingSortingByParentID(t *testing.T) {
	from := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	start := from.Add(time.Hour)
	parentA := "parent-a"
	parentB := "parent-b"

	a := baseOcc("occ-b", start, SourceDonna)
	a.ParentID = &parentB
	b := baseOcc("occ-a", start, SourceDonna)
	b.ParentID = &parentA

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{a, b}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if got[0].OccurrenceID != "occ-a" || got[1].OccurrenceID != "occ-b" {
		t.Fatalf("got [%s, %s]", got[0].OccurrenceID, got[1].OccurrenceID)
	}
}

func TestListUpcomingDuplicateHandling(t *testing.T) {
	from := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	start := from.Add(time.Hour)

	first := baseOcc("same-id", start, SourceGoogle)
	first.Title = "first"
	second := baseOcc("same-id", start.Add(time.Minute), SourceDonna)
	second.Title = "second"

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{second, first}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d items, want 1", len(got))
	}
	if got[0].Title != "first" {
		t.Fatalf("kept %q, want first (earlier StartAt)", got[0].Title)
	}
}

func TestListUpcomingMixedRecurringAndOneTime(t *testing.T) {
	from := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC) // Monday
	to := from.Add(7 * 24 * time.Hour)
	rule := "FREQ=DAILY;COUNT=3"
	parent := "series-1"

	series := Occurrence{
		ID:             parent,
		OccurrenceID:   parent,
		UserID:         testUser(),
		Source:         SourceDonna,
		Type:           TypeEvent,
		Title:          "standup",
		StartAt:        from.Add(10 * time.Hour),
		EndAt:          from.Add(10*time.Hour + 30*time.Minute),
		Timezone:       "UTC",
		RecurrenceRule: &rule,
		Status:         StatusActive,
	}
	oneTime := baseOcc("one-shot", from.Add(12*time.Hour), SourceGoogle)

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{series, oneTime}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d items, want 4 (3 recurring + 1 one-time)", len(got))
	}

	var recurring, oneShot int
	for _, item := range got {
		if item.OccurrenceID == "one-shot" {
			oneShot++
			if item.ParentID != nil {
				t.Fatalf("one-shot should have no parent: %#v", item)
			}
			continue
		}
		recurring++
		if item.ParentID == nil || *item.ParentID != parent {
			t.Fatalf("recurring missing parent: %#v", item)
		}
	}
	if recurring != 3 || oneShot != 1 {
		t.Fatalf("recurring=%d oneShot=%d", recurring, oneShot)
	}
	if !got[0].StartAt.Before(got[1].StartAt) && !got[0].StartAt.Equal(got[1].StartAt) {
		t.Fatalf("not chronological: %v then %v", got[0].StartAt, got[1].StartAt)
	}
}

func TestListUpcomingAlreadyExpandedSkipsReexpand(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	parent := "series"
	rule := "FREQ=DAILY;COUNT=10"
	item := baseOcc("series#20260801T100000Z", from.Add(10*time.Hour), SourceDonna)
	item.ParentID = &parent
	item.RecurrenceRule = &rule

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{item}},
		},
	})

	got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatalf("ListUpcoming: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 (no re-expand)", len(got))
	}
	if got[0].OccurrenceID != item.OccurrenceID {
		t.Fatalf("id changed: %s", got[0].OccurrenceID)
	}
}

func TestListUpcomingValidation(t *testing.T) {
	svc := NewService(ServiceDeps{})
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)

	if _, err := svc.ListUpcoming(context.Background(), uuid.Nil, from, to); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("nil user: %v", err)
	}
	if _, err := svc.ListUpcoming(context.Background(), testUser(), time.Time{}, to); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("zero from: %v", err)
	}
	if _, err := svc.ListUpcoming(context.Background(), testUser(), to, from); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("inverted range: %v", err)
	}
}

func TestListUpcomingProviderError(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{err: errors.New("boom")},
		},
	})
	if _, err := svc.ListUpcoming(context.Background(), testUser(), from, to); err == nil {
		t.Fatal("expected error")
	}
}

func TestListUpcomingWithStats(t *testing.T) {
	from := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	to := from.Add(7 * 24 * time.Hour)
	rule := "FREQ=DAILY;COUNT=3"
	parent := "series-1"

	series := Occurrence{
		ID:             parent,
		OccurrenceID:   parent,
		UserID:         testUser(),
		Source:         SourceDonna,
		Type:           TypeEvent,
		Title:          "standup",
		StartAt:        from.Add(10 * time.Hour),
		EndAt:          from.Add(10*time.Hour + 30*time.Minute),
		Timezone:       "UTC",
		RecurrenceRule: &rule,
		Status:         StatusActive,
	}

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: []Occurrence{series, baseOcc("one", from.Add(12*time.Hour), SourceGoogle)}},
			nil,
		},
	})

	items, stats, err := ListUpcomingWithStats(svc, context.Background(), testUser(), from, to)
	if err != nil {
		t.Fatal(err)
	}
	if stats.ProvidersQueried != 1 {
		t.Fatalf("providers = %d", stats.ProvidersQueried)
	}
	if stats.OccurrencesReturned != 2 {
		t.Fatalf("returned = %d", stats.OccurrencesReturned)
	}
	if stats.AfterExpansion != 4 {
		t.Fatalf("expanded = %d", stats.AfterExpansion)
	}
	if stats.AfterDedup != 4 || len(items) != 4 {
		t.Fatalf("dedup = %d len=%d", stats.AfterDedup, len(items))
	}
}

func BenchmarkListUpcomingLarge(b *testing.B) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(30 * 24 * time.Hour)
	items := make([]Occurrence, 0, 5000)
	for i := 0; i < 5000; i++ {
		id := uuid.NewString()
		start := from.Add(time.Duration(i) * time.Minute)
		source := SourceGoogle
		switch i % 3 {
		case 1:
			source = SourceMicrosoftICS
		case 2:
			source = SourceDonna
		}
		items = append(items, baseOcc(id, start, source))
	}
	items = append(items, items[0], items[1], items[2])

	svc := NewService(ServiceDeps{
		Providers: []Provider{
			stubProvider{items: items[:2000]},
			stubProvider{items: items[2000:4000]},
			stubProvider{items: items[4000:]},
		},
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := svc.ListUpcoming(context.Background(), testUser(), from, to)
		if err != nil {
			b.Fatal(err)
		}
		if len(got) < 5000 {
			b.Fatalf("got %d", len(got))
		}
	}
}
