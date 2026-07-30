package business

import (
	"context"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// googleTimelineProvider maps Google Calendar events into TimelineItems.
type googleTimelineProvider struct {
	events repository.CalendarEventRepository
}

// NewGoogleTimelineProvider constructs a Google calendar TimelineProvider.
func NewGoogleTimelineProvider(events repository.CalendarEventRepository) TimelineProvider {
	return &googleTimelineProvider{events: events}
}

func (p *googleTimelineProvider) Fetch(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	return fetchCalendarTimeline(ctx, p.events, userID, from, to, func(provider string) bool {
		return provider == constant.AuthProviderGoogle
	})
}

// microsoftICSTimelineProvider maps Microsoft Graph + ICS feed events into TimelineItems.
type microsoftICSTimelineProvider struct {
	events repository.CalendarEventRepository
}

// NewMicrosoftICSTimelineProvider constructs a Microsoft/ICS TimelineProvider.
func NewMicrosoftICSTimelineProvider(events repository.CalendarEventRepository) TimelineProvider {
	return &microsoftICSTimelineProvider{events: events}
}

func (p *microsoftICSTimelineProvider) Fetch(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	return fetchCalendarTimeline(ctx, p.events, userID, from, to, func(provider string) bool {
		return provider == constant.AuthProviderMicrosoft || provider == constant.AuthProviderICS
	})
}

func fetchCalendarTimeline(
	ctx context.Context,
	events repository.CalendarEventRepository,
	userID uuid.UUID,
	from, to time.Time,
	match func(provider string) bool,
) ([]entity.TimelineItem, error) {
	if events == nil {
		return nil, nil
	}
	rows, err := events.ListByUserInRangeWithProvider(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.TimelineItem, 0, len(rows))
	for _, row := range rows {
		if !match(row.Provider) {
			continue
		}
		out = append(out, mapCalendarEventToTimeline(row))
	}
	return out, nil
}

// donnaEventTimelineProvider maps Donna-owned events into TimelineItems.
type donnaEventTimelineProvider struct {
	events repository.DonnaEventRepository
}

// NewDonnaEventTimelineProvider constructs a Donna-event TimelineProvider.
func NewDonnaEventTimelineProvider(events repository.DonnaEventRepository) TimelineProvider {
	return &donnaEventTimelineProvider{events: events}
}

func (p *donnaEventTimelineProvider) Fetch(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	if p.events == nil {
		return nil, nil
	}
	events, err := p.events.ListByUserInRange(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.TimelineItem, 0, len(events))
	for _, event := range events {
		items, err := expandDonnaEvent(event, from, to)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

// donnaReminderTimelineProvider maps Donna-owned reminders into TimelineItems.
type donnaReminderTimelineProvider struct {
	reminders repository.DonnaReminderRepository
}

// NewDonnaReminderTimelineProvider constructs a Donna-reminder TimelineProvider.
func NewDonnaReminderTimelineProvider(reminders repository.DonnaReminderRepository) TimelineProvider {
	return &donnaReminderTimelineProvider{reminders: reminders}
}

func (p *donnaReminderTimelineProvider) Fetch(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	if p.reminders == nil {
		return nil, nil
	}
	reminders, err := p.reminders.ListByUserInRange(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.TimelineItem, 0, len(reminders))
	for _, reminder := range reminders {
		items, err := expandDonnaReminder(reminder, from, to)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func mapCalendarEventToTimeline(row entity.CalendarEventWithProvider) entity.TimelineItem {
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
	return entity.TimelineItem{
		ID:           e.ID.String(),
		Source:       timelineSourceFromProvider(row.Provider),
		Type:         constant.TimelineTypeEvent,
		Status:       timelineStatusFromCalendar(e.Status),
		Title:        e.Title,
		Description:  e.Description,
		StartAt:      e.StartsAt,
		EndAt:        e.EndsAt,
		Timezone:     tz,
		AllDay:       e.IsAllDay,
		Color:        row.SourceColor,
		ReadOnly:     true,
		Metadata:     meta,
		IsRecurring:  false,
		OccurrenceID: e.ID.String(),
	}
}

func mapDonnaEventToTimeline(e entity.DonnaEvent) entity.TimelineItem {
	return mapDonnaEventOccurrence(e, e.StartAt, e.EndAt, false, nil)
}

func expandDonnaEvent(e entity.DonnaEvent, from, to time.Time) ([]entity.TimelineItem, error) {
	rule, ok := NormalizeRecurrenceRule(ptrString(e.RecurrenceRule))
	if !ok {
		item := mapDonnaEventOccurrence(e, e.StartAt, e.EndAt, false, nil)
		if e.StartAt.Before(to) && e.EndAt.After(from) {
			return []entity.TimelineItem{item}, nil
		}
		return nil, nil
	}
	occs, err := ExpandRecurrence(rule, e.StartAt, e.EndAt, e.Timezone, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.TimelineItem, 0, len(occs))
	for _, occ := range occs {
		out = append(out, mapDonnaEventOccurrence(e, occ.Start, occ.End, true, &rule))
	}
	return out, nil
}

func mapDonnaEventOccurrence(
	e entity.DonnaEvent,
	start, end time.Time,
	isRecurring bool,
	rule *string,
) entity.TimelineItem {
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
	var occStart, occEnd *time.Time
	if isRecurring {
		occID = OccurrenceID(parentID, start)
		parentPtr = &parentID
		s, en := start, end
		occStart, occEnd = &s, &en
	} else {
		occID = parentID
	}

	return entity.TimelineItem{
		ID:              occID,
		Source:          constant.TimelineSourceDonna,
		Type:            constant.TimelineTypeEvent,
		Status:          timelineStatusFromDonnaEvent(e.Status),
		Title:           e.Title,
		Description:     e.Description,
		StartAt:         start,
		EndAt:           end,
		Timezone:        e.Timezone,
		AllDay:          e.AllDay,
		Color:           e.Color,
		ReadOnly:        false,
		Metadata:        meta,
		IsRecurring:     isRecurring,
		RecurrenceRule:  rule,
		ParentID:        parentPtr,
		OccurrenceID:    occID,
		OccurrenceStart: occStart,
		OccurrenceEnd:   occEnd,
	}
}

func mapDonnaReminderToTimeline(r entity.DonnaReminder) entity.TimelineItem {
	return mapDonnaReminderOccurrence(r, r.TriggerAt, false, nil)
}

func expandDonnaReminder(r entity.DonnaReminder, from, to time.Time) ([]entity.TimelineItem, error) {
	rule, ok := NormalizeRecurrenceRule(ptrString(r.RecurrenceRule))
	if !ok {
		if !r.TriggerAt.Before(to) || r.TriggerAt.Before(from) {
			return nil, nil
		}
		return []entity.TimelineItem{mapDonnaReminderOccurrence(r, r.TriggerAt, false, nil)}, nil
	}
	occs, err := ExpandRecurrence(rule, r.TriggerAt, r.TriggerAt, r.Timezone, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.TimelineItem, 0, len(occs))
	for _, occ := range occs {
		out = append(out, mapDonnaReminderOccurrence(r, occ.Start, true, &rule))
	}
	return out, nil
}

func mapDonnaReminderOccurrence(
	r entity.DonnaReminder,
	trigger time.Time,
	isRecurring bool,
	rule *string,
) entity.TimelineItem {
	meta := map[string]any{
		"public_id": r.PublicID,
	}
	if rule != nil {
		meta["recurrence_rule"] = *rule
		meta["repeat"] = true
	}

	parentID := r.ID.String()
	occID := parentID
	var parentPtr *string
	var occStart, occEnd *time.Time
	if isRecurring {
		occID = OccurrenceID(parentID, trigger)
		parentPtr = &parentID
		s := trigger
		occStart, occEnd = &s, &s
	}

	return entity.TimelineItem{
		ID:              occID,
		Source:          constant.TimelineSourceDonna,
		Type:            constant.TimelineTypeReminder,
		Status:          timelineStatusFromDonnaReminder(r.Status),
		Title:           r.Title,
		Description:     r.Description,
		StartAt:         trigger,
		EndAt:           trigger,
		Timezone:        r.Timezone,
		AllDay:          false,
		Color:           r.Color,
		ReadOnly:        false,
		Metadata:        meta,
		IsRecurring:     isRecurring,
		RecurrenceRule:  rule,
		ParentID:        parentPtr,
		OccurrenceID:    occID,
		OccurrenceStart: occStart,
		OccurrenceEnd:   occEnd,
	}
}

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func timelineSourceFromProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case constant.AuthProviderGoogle:
		return constant.TimelineSourceGoogle
	case constant.AuthProviderMicrosoft, constant.AuthProviderICS:
		return constant.TimelineSourceMicrosoftICS
	default:
		return constant.TimelineSourceMicrosoftICS
	}
}

func timelineStatusFromCalendar(status string) string {
	switch status {
	case constant.CalendarEventStatusCancelled:
		return constant.TimelineStatusCancelled
	default:
		return constant.TimelineStatusActive
	}
}

func timelineStatusFromDonnaEvent(status string) string {
	switch status {
	case constant.DonnaEventStatusCancelled:
		return constant.TimelineStatusCancelled
	default:
		return constant.TimelineStatusActive
	}
}

func timelineStatusFromDonnaReminder(status string) string {
	switch status {
	case constant.DonnaReminderStatusCancelled:
		return constant.TimelineStatusCancelled
	default:
		return constant.TimelineStatusActive
	}
}
