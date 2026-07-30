package business

import (
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
)

func TestValidateRecurrenceRule(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		got, err := ValidateRecurrenceRule(nil)
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v)", got, err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		empty := "  "
		got, err := ValidateRecurrenceRule(&empty)
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v)", got, err)
		}
	})

	t.Run("valid with prefix", func(t *testing.T) {
		raw := "RRULE:FREQ=DAILY"
		got, err := ValidateRecurrenceRule(&raw)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil || *got != "FREQ=DAILY" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		raw := "FREQ=NEVER"
		_, err := ValidateRecurrenceRule(&raw)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestExpandRecurrenceDaily(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=DAILY", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 3 {
		t.Fatalf("len = %d, want 3", len(occs))
	}
	for i, occ := range occs {
		want := start.AddDate(0, 0, i)
		if !occ.Start.Equal(want) {
			t.Fatalf("occ[%d].Start = %v, want %v", i, occ.Start, want)
		}
		if !occ.End.Equal(want.Add(time.Hour)) {
			t.Fatalf("occ[%d].End = %v", i, occ.End)
		}
	}
}

func TestExpandRecurrenceWeeklyMultiDay(t *testing.T) {
	t.Parallel()
	// Wednesday 1 July 2026 19:00 UTC
	start := time.Date(2026, 7, 1, 19, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=WEEKLY;BYDAY=WE,FR", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) < 8 {
		t.Fatalf("len = %d, want at least 8 WE/FR in July", len(occs))
	}
	for _, occ := range occs {
		wd := occ.Start.Weekday()
		if wd != time.Wednesday && wd != time.Friday {
			t.Fatalf("unexpected weekday %v at %v", wd, occ.Start)
		}
		if occ.Start.Month() != time.July {
			t.Fatalf("occurrence outside July: %v", occ.Start)
		}
	}
}

func TestExpandRecurrenceMonthly(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=MONTHLY", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 4 {
		t.Fatalf("len = %d, want 4 (Jan–Apr)", len(occs))
	}
	for i, occ := range occs {
		if occ.Start.Day() != 15 {
			t.Fatalf("occ[%d] day = %d", i, occ.Start.Day())
		}
	}
}

func TestExpandRecurrenceHourly(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC)
	end := start.Add(15 * time.Minute)
	from := time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=HOURLY;INTERVAL=2", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	// 08:00, 10:00, 12:00
	if len(occs) != 3 {
		t.Fatalf("len = %d, want 3", len(occs))
	}
	if !occs[1].Start.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("second occ = %v", occs[1].Start)
	}
}

func TestExpandRecurrenceNoInfiniteOutsideRange(t *testing.T) {
	t.Parallel()
	start := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=DAILY", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 7 {
		t.Fatalf("len = %d, want 7 days in window", len(occs))
	}
}

func TestExpandRecurrenceTimezone(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	// 19:00 IST on Wed 1 Jul 2026
	localStart := time.Date(2026, 7, 1, 19, 0, 0, 0, loc)
	startUTC := localStart.UTC()
	endUTC := startUTC.Add(time.Hour)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)

	occs, err := ExpandRecurrence("FREQ=WEEKLY;BYDAY=WE", startUTC, endUTC, "Asia/Kolkata", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 1 {
		t.Fatalf("len = %d, want 1", len(occs))
	}
	gotLocal := occs[0].Start.In(loc)
	if gotLocal.Hour() != 19 || gotLocal.Weekday() != time.Wednesday {
		t.Fatalf("got local %v", gotLocal)
	}
}

func TestExpandDonnaEventNoRecurrence(t *testing.T) {
	t.Parallel()
	id := mustParseUUID(t, "018f0000-0000-7000-8000-000000000010")
	start := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)
	event := donnaEventFixture(id, start, start.Add(time.Hour), nil)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	items, err := expandDonnaEvent(event, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d", len(items))
	}
	if items[0].IsRecurring {
		t.Fatal("expected non-recurring")
	}
	if items[0].OccurrenceID != id.String() {
		t.Fatalf("occurrence id = %s", items[0].OccurrenceID)
	}
}

func TestExpandDonnaEventWeekly(t *testing.T) {
	t.Parallel()
	id := mustParseUUID(t, "018f0000-0000-7000-8000-000000000011")
	start := time.Date(2026, 7, 1, 19, 0, 0, 0, time.UTC)
	rule := "FREQ=WEEKLY;BYDAY=WE,FR"
	event := donnaEventFixture(id, start, start.Add(time.Hour), &rule)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)

	items, err := expandDonnaEvent(event, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 4 {
		t.Fatalf("len = %d", len(items))
	}
	for _, item := range items {
		if !item.IsRecurring {
			t.Fatal("expected recurring")
		}
		if item.ParentID == nil || *item.ParentID != id.String() {
			t.Fatalf("parent = %v", item.ParentID)
		}
		if item.ID != item.OccurrenceID {
			t.Fatalf("id/occurrence mismatch %s vs %s", item.ID, item.OccurrenceID)
		}
		if item.OccurrenceStart == nil {
			t.Fatal("missing occurrence_start")
		}
	}
}

func TestOccurrenceIDStable(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 2, 19, 0, 0, 0, time.UTC)
	a := OccurrenceID("parent", start)
	b := OccurrenceID("parent", start)
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
	if a != "parent_20260702T190000Z" {
		t.Fatalf("got %s", a)
	}
}
