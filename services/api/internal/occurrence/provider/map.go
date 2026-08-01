package provider

import (
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/recurrence"
)

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func trimOptional(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func sourceFromProvider(provider string) occurrence.OccurrenceSource {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case constant.AuthProviderGoogle:
		return occurrence.SourceGoogle
	case constant.AuthProviderMicrosoft, constant.AuthProviderICS:
		return occurrence.SourceMicrosoftICS
	default:
		return occurrence.SourceMicrosoftICS
	}
}

func statusFromCalendar(status string) occurrence.OccurrenceStatus {
	switch status {
	case constant.CalendarEventStatusCancelled:
		return occurrence.StatusCancelled
	default:
		return occurrence.StatusActive
	}
}

func statusFromDonnaEvent(status string) occurrence.OccurrenceStatus {
	switch status {
	case constant.DonnaEventStatusCancelled:
		return occurrence.StatusCancelled
	default:
		return occurrence.StatusActive
	}
}

func statusFromDonnaReminder(status string) occurrence.OccurrenceStatus {
	switch status {
	case constant.DonnaReminderStatusCancelled:
		return occurrence.StatusCancelled
	default:
		return occurrence.StatusActive
	}
}

func mapCalendarEvent(row entity.CalendarEventWithProvider) occurrence.Occurrence {
	e := row.Event
	tz := "UTC"
	if e.Timezone != nil && strings.TrimSpace(*e.Timezone) != "" {
		tz = strings.TrimSpace(*e.Timezone)
	}
	meta := map[string]any{
		"public_id":          e.PublicID,
		"calendar_source_id": e.CalendarSourceID.String(),
		"origin":             e.Origin,
		"provider":           row.Provider,
		"provider_status":    e.Status,
	}
	if e.Location != nil {
		meta["location"] = *e.Location
	}
	if e.ProviderEventID != nil {
		meta["provider_event_id"] = *e.ProviderEventID
	}
	id := e.ID.String()
	return occurrence.Occurrence{
		ID:           id,
		OccurrenceID: id,
		UserID:       e.UserID,
		Source:       sourceFromProvider(row.Provider),
		Type:         occurrence.TypeEvent,
		Title:        e.Title,
		Description:  trimOptional(e.Description),
		StartAt:      e.StartsAt,
		EndAt:        e.EndsAt,
		Timezone:     tz,
		Status:       statusFromCalendar(e.Status),
		Metadata:     meta,
	}
}

func mapDonnaEventOccurrence(
	e entity.DonnaEvent,
	start, end time.Time,
	isRecurring bool,
	rule *string,
) occurrence.Occurrence {
	meta := map[string]any{
		"public_id": e.PublicID,
	}
	if e.Location != nil {
		meta["location"] = *e.Location
	}
	if e.ReminderOffsetMinutes != nil {
		meta["reminder_offset_minutes"] = *e.ReminderOffsetMinutes
	}
	if rule != nil {
		meta["recurrence_rule"] = *rule
	}

	parentID := e.ID.String()
	occID := parentID
	var parentPtr *string
	if isRecurring {
		occID = recurrence.ID(parentID, start)
		parentPtr = &parentID
	}

	return occurrence.Occurrence{
		ID:             occID,
		ParentID:       parentPtr,
		OccurrenceID:   occID,
		UserID:         e.UserID,
		Source:         occurrence.SourceDonna,
		Type:           occurrence.TypeEvent,
		Title:          e.Title,
		Description:    trimOptional(e.Description),
		StartAt:        start,
		EndAt:          end,
		Timezone:       e.Timezone,
		RecurrenceRule: rule,
		Status:         statusFromDonnaEvent(e.Status),
		Metadata:       meta,
	}
}

func mapDonnaReminderOccurrence(
	r entity.DonnaReminder,
	trigger time.Time,
	isRecurring bool,
	rule *string,
) occurrence.Occurrence {
	meta := map[string]any{
		"public_id": r.PublicID,
	}
	if rule != nil {
		meta["recurrence_rule"] = *rule
	}

	parentID := r.ID.String()
	occID := parentID
	var parentPtr *string
	if isRecurring {
		occID = recurrence.ID(parentID, trigger)
		parentPtr = &parentID
	}

	return occurrence.Occurrence{
		ID:             occID,
		ParentID:       parentPtr,
		OccurrenceID:   occID,
		UserID:         r.UserID,
		Source:         occurrence.SourceDonna,
		Type:           occurrence.TypeReminder,
		Title:          r.Title,
		Description:    trimOptional(r.Description),
		StartAt:        trigger,
		EndAt:          trigger,
		Timezone:       r.Timezone,
		RecurrenceRule: rule,
		Status:         statusFromDonnaReminder(r.Status),
		Metadata:       meta,
	}
}

func expandDonnaEvent(e entity.DonnaEvent, from, to time.Time) ([]occurrence.Occurrence, error) {
	rule, ok := recurrence.NormalizeRule(ptrString(e.RecurrenceRule))
	if !ok {
		item := mapDonnaEventOccurrence(e, e.StartAt, e.EndAt, false, nil)
		if e.StartAt.Before(to) && e.EndAt.After(from) {
			return []occurrence.Occurrence{item}, nil
		}
		return nil, nil
	}
	occs, err := recurrence.Expand(rule, e.StartAt, e.EndAt, e.Timezone, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]occurrence.Occurrence, 0, len(occs))
	for _, occ := range occs {
		out = append(out, mapDonnaEventOccurrence(e, occ.Start, occ.End, true, &rule))
	}
	return out, nil
}

func expandDonnaReminder(r entity.DonnaReminder, from, to time.Time) ([]occurrence.Occurrence, error) {
	rule, ok := recurrence.NormalizeRule(ptrString(r.RecurrenceRule))
	if !ok {
		if !r.TriggerAt.Before(to) || r.TriggerAt.Before(from) {
			return nil, nil
		}
		return []occurrence.Occurrence{mapDonnaReminderOccurrence(r, r.TriggerAt, false, nil)}, nil
	}
	occs, err := recurrence.Expand(rule, r.TriggerAt, r.TriggerAt, r.Timezone, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]occurrence.Occurrence, 0, len(occs))
	for _, occ := range occs {
		out = append(out, mapDonnaReminderOccurrence(r, occ.Start, true, &rule))
	}
	return out, nil
}
