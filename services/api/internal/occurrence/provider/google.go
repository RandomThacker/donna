package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// GoogleOccurrenceProvider maps Google Calendar events into Occurrences
// using the narrow scheduler projection (no provider_payload).
type GoogleOccurrenceProvider struct {
	events CalendarEventSchedulerReader
	log    *logger.Logger
}

// NewGoogleOccurrenceProvider constructs a GoogleOccurrenceProvider.
func NewGoogleOccurrenceProvider(events CalendarEventSchedulerReader, log *logger.Logger) *GoogleOccurrenceProvider {
	return &GoogleOccurrenceProvider{events: events, log: log}
}

// ListOccurrences implements OccurrenceProvider.
func (p *GoogleOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	if p.events == nil {
		return nil, nil
	}
	start := time.Now()
	rows, err := p.events.ListForSchedulerByUserInRange(
		ctx, userID, from, to, []string{constant.AuthProviderGoogle},
	)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	logSchedulerQuery(
		ctx, p.log, "google",
		repository.CalendarEventSchedulerColumnNames,
		EstBytesPerRowCalendarScheduler,
		len(rows),
		duration,
	)
	out := make([]occurrence.Occurrence, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCalendarEvent(row))
	}
	return out, nil
}

var _ OccurrenceProvider = (*GoogleOccurrenceProvider)(nil)
