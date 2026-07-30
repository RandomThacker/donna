package entity

import "time"

// TimelineItem is the unified DTO for calendar, agenda, brief, and future scheduling.
// It is built in memory from provider events and Donna-owned rows — never persisted as a table.
type TimelineItem struct {
	ID          string
	Source      string
	Type        string
	Status      string
	Title       string
	Description *string
	StartAt     time.Time
	EndAt       time.Time
	Timezone    string
	AllDay      bool
	Color       *string
	ReadOnly    bool
	// Metadata holds source-specific extras (meet links, organizer, priority, …)
	// so TimelineItem stays generic across providers.
	Metadata map[string]any

	// Recurrence (virtual occurrences — never persisted).
	IsRecurring     bool
	RecurrenceRule  *string
	ParentID        *string
	OccurrenceID    string
	OccurrenceStart *time.Time
	OccurrenceEnd   *time.Time
}
