package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

// DonnaEventOccurrenceProvider maps Donna-owned events into Occurrences.
type DonnaEventOccurrenceProvider struct {
	events DonnaEventReader
}

// NewDonnaEventOccurrenceProvider constructs a DonnaEventOccurrenceProvider.
func NewDonnaEventOccurrenceProvider(events DonnaEventReader) *DonnaEventOccurrenceProvider {
	return &DonnaEventOccurrenceProvider{events: events}
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
	events, err := p.events.ListByUserInRange(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
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
