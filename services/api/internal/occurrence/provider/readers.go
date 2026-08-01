package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// CalendarEventReader is the calendar-event read surface Occurrence providers need.
type CalendarEventReader interface {
	ListByUserInRangeWithProvider(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.CalendarEventWithProvider, error)
}

// DonnaEventReader is the Donna-event read surface Occurrence providers need.
type DonnaEventReader interface {
	ListByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaEvent, error)
}

// DonnaReminderReader is the Donna-reminder read surface Occurrence providers need.
type DonnaReminderReader interface {
	ListByUserInRange(
		ctx context.Context,
		userID uuid.UUID,
		from, to time.Time,
	) ([]entity.DonnaReminder, error)
}
