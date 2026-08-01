package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

// DonnaReminderOccurrenceProvider maps Donna-owned reminders into Occurrences.
type DonnaReminderOccurrenceProvider struct {
	reminders DonnaReminderReader
}

// NewDonnaReminderOccurrenceProvider constructs a DonnaReminderOccurrenceProvider.
func NewDonnaReminderOccurrenceProvider(reminders DonnaReminderReader) *DonnaReminderOccurrenceProvider {
	return &DonnaReminderOccurrenceProvider{reminders: reminders}
}

// ListOccurrences implements OccurrenceProvider.
func (p *DonnaReminderOccurrenceProvider) ListOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]occurrence.Occurrence, error) {
	if p.reminders == nil {
		return nil, nil
	}
	reminders, err := p.reminders.ListByUserInRange(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]occurrence.Occurrence, 0, len(reminders))
	for _, reminder := range reminders {
		items, err := expandDonnaReminder(reminder, from, to)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

var _ OccurrenceProvider = (*DonnaReminderOccurrenceProvider)(nil)
