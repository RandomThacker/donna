package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// DonnaReminderOccurrenceProvider maps Donna-owned reminders into Occurrences
// using the narrow scheduler projection.
type DonnaReminderOccurrenceProvider struct {
	reminders DonnaReminderSchedulerReader
	log       *logger.Logger
}

// NewDonnaReminderOccurrenceProvider constructs a DonnaReminderOccurrenceProvider.
func NewDonnaReminderOccurrenceProvider(reminders DonnaReminderSchedulerReader, log *logger.Logger) *DonnaReminderOccurrenceProvider {
	return &DonnaReminderOccurrenceProvider{reminders: reminders, log: log}
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
	start := time.Now()
	reminders, err := p.reminders.ListForSchedulerByUserInRange(ctx, userID, from, to)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	logSchedulerQuery(
		ctx, p.log, "donna_reminder",
		repository.DonnaReminderSchedulerColumnNames,
		EstBytesPerRowDonnaReminder,
		len(reminders),
		duration,
	)
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
