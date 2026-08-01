package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// DonnaEventOccurrenceProvider maps Donna-owned events into Occurrences
// using the narrow scheduler projection.
type DonnaEventOccurrenceProvider struct {
	events DonnaEventSchedulerReader
	log    *logger.Logger
}

// NewDonnaEventOccurrenceProvider constructs a DonnaEventOccurrenceProvider.
func NewDonnaEventOccurrenceProvider(events DonnaEventSchedulerReader, log *logger.Logger) *DonnaEventOccurrenceProvider {
	return &DonnaEventOccurrenceProvider{events: events, log: log}
}

// ListOccurrences implements OccurrenceProvider.
func (p *DonnaEventOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	if p.events == nil {
		return nil, nil
	}
	start := time.Now()
	events, err := p.events.ListForSchedulerByUserInRange(ctx, userID, from, to)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	logSchedulerQuery(
		ctx, p.log, "donna_event",
		repository.DonnaEventSchedulerColumnNames,
		EstBytesPerRowDonnaEvent,
		len(events),
		duration,
	)
	out := make([]occurrence.Occurrence, 0, len(events))
	for _, event := range events {
		items, err := expandDonnaEvent(event, from, to)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

var _ OccurrenceProvider = (*DonnaEventOccurrenceProvider)(nil)
