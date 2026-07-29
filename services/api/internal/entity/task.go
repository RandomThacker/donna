package entity

import (
	"time"

	"github.com/google/uuid"
)

// Task is the permanent work item aggregate root.
type Task struct {
	ID              uuid.UUID
	PublicID        string
	UserID          uuid.UUID
	Title           string
	Description     *string
	Status          string
	Priority        *string
	Project         *string
	Labels          []string
	DueAt           *time.Time
	CompletedAt     *time.Time
	IsBacklog       bool
	RecurrenceRule  *string
	Provider        *string
	ProviderTaskID  *string
	ProviderPayload []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// TaskOccurrence is a daily journal row for a task on one civil day.
type TaskOccurrence struct {
	ID              uuid.UUID
	PublicID        string
	TaskID          uuid.UUID
	UserID          uuid.UUID
	Date            time.Time
	SortOrder       int
	Completed       bool
	CompletedAt     *time.Time
	CarriedForward  bool
	Source          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TaskOccurrenceWithTask joins occurrence with its permanent task for display.
type TaskOccurrenceWithTask struct {
	TaskOccurrence
	Title          string
	Description    *string
	Priority       *string
	Project        *string
	Labels         []string
	RecurrenceRule *string
	Tags           []TaskTag
}

// DailyNote is markdown content for one journal day.
type DailyNote struct {
	ID        uuid.UUID
	PublicID  string
	UserID    uuid.UUID
	Date      time.Time
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TaskDaySummary is aggregate stats for one civil day (history / mini calendar).
type TaskDaySummary struct {
	Date      time.Time
	Total     int
	Completed int
	Pending   int
	Carried   int
}

// TaskDayStatistics holds analytics for one journal day.
type TaskDayStatistics struct {
	Total                int
	Completed            int
	Pending              int
	Carried              int
	CompletionPct        float64
	CompletedToday       int
	CarriedForward       int
	LongestCarriedStreak int
	AverageCompletionMin *float64
	Streak               int
}

// TaskJournalDay is the full journal page for one civil day.
type TaskJournalDay struct {
	Date        time.Time
	Note        DailyNote
	Statistics  TaskDayStatistics
	Occurrences []TaskOccurrenceWithTask
}
