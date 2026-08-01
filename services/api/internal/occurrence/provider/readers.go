package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// CalendarEventReader is the wide calendar-event read surface (Timeline / Microsoft ICS).
type CalendarEventReader interface {
	ListByUserInRangeWithProvider(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.CalendarEventWithProvider, error)
}

// CalendarEventSchedulerReader is the narrow calendar read surface for Occurrence scheduling.
type CalendarEventSchedulerReader interface {
	ListForSchedulerByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
		providers []string,
	) ([]entity.CalendarEventWithProvider, error)
	// ListCalendarOccurrences is the Sprint 1B shared calendar Occurrence read
	// (same projection as ListForSchedulerByUserInRange).
	ListCalendarOccurrences(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
		providers []string,
	) ([]entity.CalendarEventWithProvider, error)
}

// DonnaEventReader is the wide Donna-event read surface (Timeline).
type DonnaEventReader interface {
	ListByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaEvent, error)
}

// DonnaEventSchedulerReader is the narrow Donna-event read surface for Occurrence scheduling.
type DonnaEventSchedulerReader interface {
	ListForSchedulerByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaEvent, error)
}

// DonnaReminderReader is the wide Donna-reminder read surface (Timeline).
type DonnaReminderReader interface {
	ListByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaReminder, error)
}

// DonnaReminderSchedulerReader is the narrow Donna-reminder read surface for Occurrence scheduling.
type DonnaReminderSchedulerReader interface {
	ListForSchedulerByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaReminder, error)
}
