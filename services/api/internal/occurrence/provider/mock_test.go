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

func (m *memCalendarEvents) ListForSchedulerByUserInRange(
	_ context.Context,
	_ uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	if m.err != nil {
		return nil, m.err
	}
	allow := map[string]struct{}{}
	for _, p := range providers {
		allow[p] = struct{}{}
	}
	out := make([]entity.CalendarEventWithProvider, 0)
	for _, row := range m.rows {
		if len(allow) > 0 {
			if _, ok := allow[row.Provider]; !ok {
				continue
			}
		}
		e := row.Event
		if e.StartsAt.Before(to) && e.EndsAt.After(from) {
			// Mimic narrow projection: strip unused wide fields.
			narrow := row
			narrow.Event.ProviderPayload = nil
			narrow.Event.AttendeesSummary = nil
			narrow.Event.OrganizerSummary = nil
			narrow.Event.ProviderETag = nil
			narrow.SourceColor = nil
			out = append(out, narrow)
		}
	}
	return out, nil
}

func (m *memCalendarEvents) ListCalendarOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	return m.ListForSchedulerByUserInRange(ctx, userID, from, to, providers)
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

func (m *memDonnaEvents) ListForSchedulerByUserInRange(
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
			narrow := e
			narrow.Color = nil
			narrow.AllDay = false
			narrow.CreatedAt = time.Time{}
			narrow.UpdatedAt = time.Time{}
			narrow.DeletedAt = nil
			out = append(out, narrow)
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

func (m *memDonnaReminders) ListForSchedulerByUserInRange(
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
			narrow := r
			narrow.Color = nil
			narrow.CreatedAt = time.Time{}
			narrow.UpdatedAt = time.Time{}
			narrow.DeletedAt = nil
			out = append(out, narrow)
		}
	}
	return out, nil
}
