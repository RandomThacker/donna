package provider

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

type memCalendarEvents struct {
	rows []entity.CalendarEventWithProvider
	err  error
}

func (m *memCalendarEvents) ListByUserInRangeWithProvider(
	_ context.Context,
	_ uuid.UUID,
	from, to time.Time,
) ([]entity.CalendarEventWithProvider, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]entity.CalendarEventWithProvider, 0)
	for _, row := range m.rows {
		e := row.Event
		if e.StartsAt.Before(to) && e.EndsAt.After(from) {
			out = append(out, row)
		}
	}
	return out, nil
}

type memDonnaEvents struct {
	rows []entity.DonnaEvent
	err  error
}

func (m *memDonnaEvents) ListByUserInRange(
	_ context.Context,
	_ uuid.UUID,
	_, to time.Time,
) ([]entity.DonnaEvent, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]entity.DonnaEvent, 0)
	for _, e := range m.rows {
		if e.StartAt.Before(to) {
			out = append(out, e)
		}
	}
	return out, nil
}

type memDonnaReminders struct {
	rows []entity.DonnaReminder
	err  error
}

func (m *memDonnaReminders) ListByUserInRange(
	_ context.Context,
	_ uuid.UUID,
	_, to time.Time,
) ([]entity.DonnaReminder, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]entity.DonnaReminder, 0)
	for _, r := range m.rows {
		if r.TriggerAt.Before(to) {
			out = append(out, r)
		}
	}
	return out, nil
}
