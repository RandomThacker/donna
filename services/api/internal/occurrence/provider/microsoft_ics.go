package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

// MicrosoftICSOccurrenceProvider maps Microsoft Graph + ICS events into Occurrences.
type MicrosoftICSOccurrenceProvider struct {
	events CalendarEventReader
}

// NewMicrosoftICSOccurrenceProvider constructs a MicrosoftICSOccurrenceProvider.
func NewMicrosoftICSOccurrenceProvider(events CalendarEventReader) *MicrosoftICSOccurrenceProvider {
	return &MicrosoftICSOccurrenceProvider{events: events}
}

// ListOccurrences implements OccurrenceProvider.
func (p *MicrosoftICSOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	return listCalendarOccurrences(ctx, p.events, userID, from, to, func(provider string) bool {
		return provider == constant.AuthProviderMicrosoft || provider == constant.AuthProviderICS
	})
}

func listCalendarOccurrences(
	ctx context.Context,
	events CalendarEventReader,
	userID uuid.UUID,
	from, to time.Time,
	match func(provider string) bool,
) ([]occurrence.Occurrence, error) {
	if events == nil {
		return nil, nil
	}
	rows, err := events.ListByUserInRangeWithProvider(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]occurrence.Occurrence, 0, len(rows))
	for _, row := range rows {
		if !match(row.Provider) {
			continue
		}
		out = append(out, mapCalendarEvent(row))
	}
	return out, nil
}

var _ OccurrenceProvider = (*MicrosoftICSOccurrenceProvider)(nil)
