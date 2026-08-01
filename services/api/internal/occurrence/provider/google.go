package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

// GoogleOccurrenceProvider maps Google Calendar events into Occurrences.
type GoogleOccurrenceProvider struct {
	events CalendarEventReader
}

// NewGoogleOccurrenceProvider constructs a GoogleOccurrenceProvider.
func NewGoogleOccurrenceProvider(events CalendarEventReader) *GoogleOccurrenceProvider {
	return &GoogleOccurrenceProvider{events: events}
}

// ListOccurrences implements OccurrenceProvider.
func (p *GoogleOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	return listCalendarOccurrences(ctx, p.events, userID, from, to, func(provider string) bool {
		return provider == constant.AuthProviderGoogle
	})
}

var _ OccurrenceProvider = (*GoogleOccurrenceProvider)(nil)
