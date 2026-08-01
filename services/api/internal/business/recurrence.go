package business

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/recurrence"
)

// MaxOccurrencesPerSeries caps expansion inside one timeline query window.
const MaxOccurrencesPerSeries = recurrence.MaxOccurrencesPerSeries

// RecurrenceOccurrence is one virtual occurrence of a recurring series.
type RecurrenceOccurrence = recurrence.Occurrence

// NormalizeRecurrenceRule strips an optional RRULE: prefix and trims space.
// Empty input returns ("", false).
func NormalizeRecurrenceRule(raw string) (string, bool) {
	return recurrence.NormalizeRule(raw)
}

// ValidateRecurrenceRule rejects invalid iCalendar RRULE strings.
// nil / empty means no recurrence (valid).
func ValidateRecurrenceRule(raw *string) (normalized *string, err error) {
	return recurrence.ValidateRule(raw)
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
	return recurrence.Expand(rule, dtStart, dtEnd, timezoneName, from, to)
}

// OccurrenceID builds a stable virtual id for a parent + occurrence start.
func OccurrenceID(parentID string, occurrenceStart time.Time) string {
	return recurrence.ID(parentID, occurrenceStart)
}

func loadTimezone(name string) (*time.Location, error) {
	return recurrence.LoadTimezone(name)
}
