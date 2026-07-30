package business

import (
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/teambition/rrule-go"
)

// MaxOccurrencesPerSeries caps expansion inside one timeline query window.
const MaxOccurrencesPerSeries = 2000

// RecurrenceOccurrence is one virtual occurrence of a recurring series.
type RecurrenceOccurrence struct {
	Start time.Time
	End   time.Time
}

// NormalizeRecurrenceRule strips an optional RRULE: prefix and trims space.
// Empty input returns ("", false).
func NormalizeRecurrenceRule(raw string) (string, bool) {
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

// ValidateRecurrenceRule rejects invalid iCalendar RRULE strings.
// nil / empty means no recurrence (valid).
func ValidateRecurrenceRule(raw *string) (normalized *string, err error) {
	if raw == nil {
		return nil, nil
	}
	rule, ok := NormalizeRecurrenceRule(*raw)
	if !ok {
		return nil, nil
	}
	if _, err := rrule.StrToRRule(rule); err != nil {
		return nil, fmt.Errorf("%w: invalid recurrence_rule: %v", apperr.ErrValidation, err)
	}
	return &rule, nil
}

// ExpandRecurrence generates occurrence starts/ends inside [from, to) using RRULE.
// dtStart/dtEnd are the series anchor in UTC; expansion respects timezoneName.
// Duration of each occurrence equals dtEnd - dtStart (zero for reminders).
func ExpandRecurrence(
	rule string,
	dtStart, dtEnd time.Time,
	timezoneName string,
	from, to time.Time,
) ([]RecurrenceOccurrence, error) {
	normalized, ok := NormalizeRecurrenceRule(rule)
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
	out := make([]RecurrenceOccurrence, 0, len(candidates))
	for _, start := range candidates {
		startUTC := start.UTC()
		if startUTC.Before(from.UTC()) || !startUTC.Before(to.UTC()) {
			continue
		}
		out = append(out, RecurrenceOccurrence{
			Start: startUTC,
			End:   startUTC.Add(duration),
		})
		if len(out) >= MaxOccurrencesPerSeries {
			break
		}
	}
	return out, nil
}

// OccurrenceID builds a stable virtual id for a parent + occurrence start.
func OccurrenceID(parentID string, occurrenceStart time.Time) string {
	return parentID + "_" + occurrenceStart.UTC().Format("20060102T150405Z")
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
