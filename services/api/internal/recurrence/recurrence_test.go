package recurrence

import (
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
)

func TestValidateRule(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		got, err := ValidateRule(nil)
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v)", got, err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		empty := "  "
		got, err := ValidateRule(&empty)
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v)", got, err)
		}
	})

	t.Run("valid with prefix", func(t *testing.T) {
		raw := "RRULE:FREQ=DAILY"
		got, err := ValidateRule(&raw)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil || *got != "FREQ=DAILY" {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		raw := "FREQ=NEVER"
		_, err := ValidateRule(&raw)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestExpandDaily(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)

	occs, err := Expand("FREQ=DAILY", start, end, "UTC", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(occs) != 3 {
		t.Fatalf("len = %d, want 3", len(occs))
	}
	if !occs[0].Start.Equal(start) || !occs[0].End.Equal(end) {
		t.Fatalf("first = %+v", occs[0])
	}
}

func TestExpandInvalidRule(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	from := start
	to := start.Add(24 * time.Hour)
	_, err := Expand("FREQ=NEVER", start, start, "UTC", from, to)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestIDStable(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	a := ID("parent", start)
	b := ID("parent", start)
	if a != b || a == "" {
		t.Fatalf("ids = %q %q", a, b)
	}
}
