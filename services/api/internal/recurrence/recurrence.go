// Package recurrence provides shared RRULE normalization and expansion helpers
// used by Timeline providers and Occurrence providers.
package recurrence

import (
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/teambition/rrule-go"
)

// MaxOccurrencesPerSeries caps expansion inside one query window.
const MaxOccurrencesPerSeries = 2000

// Occurrence is one virtual instance of a recurring series.
type Occurrence struct {
	Start time.Time
	End   time.Time
}

// NormalizeRule strips an optional RRULE: prefix and trims space.
// Empty input returns ("", false).
func NormalizeRule(raw string) (string, bool) {
	rule := strings.TrimSpace(raw)
	if rule == "" {
		return "", false
	}
	upper := strings.ToUpper(rule)
	if strings.HasPrefix(upper, "RRULE:") {
		rule = strings.TrimSpace(rule[len("RRULE:"):])
	}
	if rule == "" {
		return "", false
	}
	return rule, true
}

// ValidateRule rejects invalid iCalendar RRULE strings.
// nil / empty means no recurrence (valid).
func ValidateRule(raw *string) (normalized *string, err error) {
	if raw == nil {
		return nil, nil
	}
	rule, ok := NormalizeRule(*raw)
	if !ok {
		return nil, nil
	}
	if _, err := rrule.StrToRRule(rule); err != nil {
		return nil, fmt.Errorf("%w: invalid recurrence_rule: %v", apperr.ErrValidation, err)
	}
	return &rule, nil
}

// Expand generates occurrence starts/ends inside [from, to) using RRULE.
// dtStart/dtEnd are the series anchor in UTC; expansion respects timezoneName.
// Duration of each occurrence equals dtEnd - dtStart (zero for reminders).
func Expand(
	rule string,
	dtStart, dtEnd time.Time,
	timezoneName string,
	from, to time.Time,
) ([]Occurrence, error) {
	normalized, ok := NormalizeRule(rule)
	if !ok {
		return nil, fmt.Errorf("%w: recurrence_rule is required for expansion", apperr.ErrValidation)
	}
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("%w: valid from/to range is required", apperr.ErrValidation)
	}

	loc, err := loadTimezone(timezoneName)
	if err != nil {
		return nil, err
	}

	anchorStart := dtStart.In(loc)
	duration := dtEnd.Sub(dtStart)
	if duration < 0 {
		duration = 0
	}

	r, err := rrule.StrToRRule(normalized)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid recurrence_rule: %v", apperr.ErrValidation, err)
	}
	r.DTStart(anchorStart)

	// Inclusive window slightly wider, then filter to half-open [from, to).
	candidates := r.Between(from.Add(-time.Second), to.Add(time.Second), true)
	out := make([]Occurrence, 0, len(candidates))
	for _, start := range candidates {
		startUTC := start.UTC()
		if startUTC.Before(from.UTC()) || !startUTC.Before(to.UTC()) {
			continue
		}
		out = append(out, Occurrence{
			Start: startUTC,
			End:   startUTC.Add(duration),
		})
		if len(out) >= MaxOccurrencesPerSeries {
			break
		}
	}
	return out, nil
}

// ID builds a stable virtual id for a parent + occurrence start.
func ID(parentID string, occurrenceStart time.Time) string {
	return parentID + "_" + occurrenceStart.UTC().Format("20060102T150405Z")
}

// LoadTimezone resolves an IANA timezone name (empty / UTC → time.UTC).
func LoadTimezone(name string) (*time.Location, error) {
	return loadTimezone(name)
}

func loadTimezone(name string) (*time.Location, error) {
	tz := strings.TrimSpace(name)
	if tz == "" || strings.EqualFold(tz, "UTC") {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid timezone %q", apperr.ErrValidation, tz)
	}
	return loc, nil
}
