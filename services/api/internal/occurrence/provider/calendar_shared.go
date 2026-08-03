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

// ActiveCalendarOccurrenceProviders are calendar account providers queried for
// Occurrence scheduling / notification enqueue. Still one SQL query via
// SharedCalendarOccurrenceProvider — providers are filtered with ANY($4).
var ActiveCalendarOccurrenceProviders = []string{
	constant.AuthProviderGoogle,
	constant.AuthProviderICS,
	constant.AuthProviderMicrosoft,
}

// SharedCalendarOccurrenceProvider loads all requested calendar providers in
// ONE narrow calendar_events query (Sprint 1B), then maps rows to Occurrences.
type SharedCalendarOccurrenceProvider struct {
	events    CalendarEventSchedulerReader
	providers []string
	log       *logger.Logger
}

// NewSharedCalendarOccurrenceProvider constructs a SharedCalendarOccurrenceProvider.
// Empty providers defaults to ActiveCalendarOccurrenceProviders.
func NewSharedCalendarOccurrenceProvider(
	events CalendarEventSchedulerReader,
	providers []string,
	log *logger.Logger,
) *SharedCalendarOccurrenceProvider {
	if len(providers) == 0 {
		providers = append([]string(nil), ActiveCalendarOccurrenceProviders...)
	}
	return &SharedCalendarOccurrenceProvider{
		events:    events,
		providers: providers,
		log:       log,
	}
}

// ListOccurrences implements OccurrenceProvider.
func (p *SharedCalendarOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	if p.events == nil {
		return nil, nil
	}
	start := time.Now()
	rows, err := p.events.ListCalendarOccurrences(ctx, userID, from, to, p.providers)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	logSchedulerQuery(
		ctx, p.log, "shared_calendar",
		repository.CalendarEventSchedulerColumnNames,
		EstBytesPerRowCalendarScheduler,
		len(rows),
		duration,
	)
	if p.log != nil {
		p.log.Info(ctx, "occurrence calendar query consolidated",
			"queries_executed", 1,
			"provider_filters", p.providers,
			"rows_returned", len(rows),
			"est_bytes_total", EstBytesPerRowCalendarScheduler*len(rows),
			"duration_ms", duration.Milliseconds(),
		)
	}
	out := make([]occurrence.Occurrence, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCalendarEvent(row))
	}
	return out, nil
}

var _ OccurrenceProvider = (*SharedCalendarOccurrenceProvider)(nil)
