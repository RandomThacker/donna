package googlecalendar

import (
	"context"
	"errors"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Ensure Client implements calendarprovider.Provider.
var _ calendarprovider.Provider = (*Provider)(nil)

// Provider adapts the Google Calendar HTTP client to calendarprovider.Provider.
type Provider struct {
	client *Client
}

// NewProvider wraps a Google Calendar API client.
func NewProvider(client *Client) *Provider {
	if client == nil {
		client = NewClient(Config{})
	}
	return &Provider{client: client}
}

func (p *Provider) Name() string { return constant.AuthProviderGoogle }

func (p *Provider) ListCalendars(
	ctx context.Context,
	accessToken string,
	opts calendarprovider.ListCalendarsOptions,
) (calendarprovider.ListCalendarsResult, error) {
	listed, err := p.client.ListCalendars(ctx, accessToken, ListOptions{SyncToken: opts.SyncToken})
	if err != nil {
		return calendarprovider.ListCalendarsResult{}, mapProviderError(err)
	}
	out := calendarprovider.ListCalendarsResult{
		Calendars:     make([]calendarprovider.RemoteCalendar, 0, len(listed.Calendars)),
		NextSyncToken: listed.NextSyncToken,
		Incremental:   listed.Incremental,
	}
	for _, c := range listed.Calendars {
		out.Calendars = append(out.Calendars, calendarprovider.RemoteCalendar{
			ID:          c.ID,
			Name:        c.Name,
			Primary:     c.Primary,
			Writable:    c.Writable,
			AccessRole:  c.AccessRole,
			Color:       c.Color,
			TimeZone:    c.TimeZone,
			ETag:        c.ETag,
			Description: c.Description,
			Hidden:      c.Hidden,
			Selected:    c.Selected,
			Deleted:     c.Deleted,
			Raw:         c.Raw,
		})
	}
	return out, nil
}

func (p *Provider) ListEvents(
	ctx context.Context,
	accessToken, calendarID string,
	opts calendarprovider.ListEventsOptions,
) (calendarprovider.ListEventsResult, error) {
	listed, err := p.client.ListEvents(ctx, accessToken, calendarID, EventListOptions{
		SyncToken: opts.SyncToken,
		TimeMin:   opts.TimeMin,
		TimeMax:   opts.TimeMax,
	})
	if err != nil {
		return calendarprovider.ListEventsResult{}, mapProviderError(err)
	}
	out := calendarprovider.ListEventsResult{
		Events:        make([]calendarprovider.RemoteEvent, 0, len(listed.Events)),
		NextSyncToken: listed.NextSyncToken,
		Incremental:   listed.Incremental,
	}
	for _, e := range listed.Events {
		attendees := make([]calendarprovider.RemoteAttendee, 0, len(e.Attendees))
		for _, a := range e.Attendees {
			attendees = append(attendees, calendarprovider.RemoteAttendee{
				Email:          a.Email,
				DisplayName:    a.DisplayName,
				ResponseStatus: a.ResponseStatus,
				Organizer:      a.Organizer,
				Self:           a.Self,
			})
		}
		out.Events = append(out.Events, calendarprovider.RemoteEvent{
			ID:                   e.ID,
			ETag:                 e.ETag,
			Status:               e.Status,
			Title:                e.Title,
			Description:          e.Description,
			Location:             e.Location,
			StartsAt:             e.StartsAt,
			EndsAt:               e.EndsAt,
			IsAllDay:             e.IsAllDay,
			Timezone:             e.Timezone,
			Visibility:           e.Visibility,
			OrganizerEmail:       e.OrganizerEmail,
			OrganizerDisplayName: e.OrganizerDisplayName,
			OrganizerSelf:        e.OrganizerSelf,
			Attendees:            attendees,
			Recurrence:           e.Recurrence,
			RecurringEventID:     e.RecurringEventID,
			UpdatedAt:            e.UpdatedAt,
			Deleted:              e.Deleted,
			Raw:                  e.Raw,
		})
	}
	return out, nil
}

func mapProviderError(err error) error {
	var gone *GoneError
	if errors.As(err, &gone) {
		return &calendarprovider.SyncCursorInvalidError{Body: gone.Body}
	}
	var auth *AuthError
	if errors.As(err, &auth) {
		return &calendarprovider.AuthError{Status: auth.Status, Body: auth.Body}
	}
	return err
}
