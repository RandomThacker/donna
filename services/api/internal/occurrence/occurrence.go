// Package occurrence defines the scheduling-domain Occurrence model.
//
// Occurrence is the input shape for notification scheduling and other
// time-based pipeline work. It is intentionally separate from TimelineItem,
// which carries UI/display concerns.
package occurrence

import (
	"time"

	"github.com/google/uuid"
)

// OccurrenceType classifies what kind of scheduled item this is.
type OccurrenceType string

const (
	TypeEvent    OccurrenceType = "EVENT"
	TypeReminder OccurrenceType = "REMINDER"
)

// Valid reports whether t is a known OccurrenceType.
func (t OccurrenceType) Valid() bool {
	switch t {
	case TypeEvent, TypeReminder:
		return true
	default:
		return false
	}
}

// OccurrenceSource identifies where the occurrence originated.
type OccurrenceSource string

const (
	SourceGoogle       OccurrenceSource = "GOOGLE"
	SourceMicrosoftICS OccurrenceSource = "MICROSOFT_ICS"
	SourceDonna        OccurrenceSource = "DONNA"
)

// Valid reports whether s is a known OccurrenceSource.
func (s OccurrenceSource) Valid() bool {
	switch s {
	case SourceGoogle, SourceMicrosoftICS, SourceDonna:
		return true
	default:
		return false
	}
}

// OccurrenceStatus is the scheduling lifecycle state (not a UI presentation state).
type OccurrenceStatus string

const (
	StatusActive    OccurrenceStatus = "ACTIVE"
	StatusCompleted OccurrenceStatus = "COMPLETED"
	StatusCancelled OccurrenceStatus = "CANCELLED"
	StatusMissed    OccurrenceStatus = "MISSED"
)

// Valid reports whether s is a known OccurrenceStatus.
func (s OccurrenceStatus) Valid() bool {
	switch s {
	case StatusActive, StatusCompleted, StatusCancelled, StatusMissed:
		return true
	default:
		return false
	}
}

// Occurrence is a lightweight scheduling unit: one concrete (or expandable)
// moment in time that the notification pipeline can reason about.
//
// It excludes UI-only fields (color, icon, read-only flags, display labels).
type Occurrence struct {
	ID           string
	ParentID     *string
	OccurrenceID string
	UserID       uuid.UUID
	Source       OccurrenceSource
	Type         OccurrenceType
	Title        string
	Description  *string
	StartAt      time.Time
	EndAt        time.Time
	Timezone     string
	RecurrenceRule *string
	Status       OccurrenceStatus
	// Metadata holds minimal non-display extras (e.g. provider event id, meet URL).
	// Callers must not put colors, icons, or frontend-only fields here.
	Metadata map[string]any
}
