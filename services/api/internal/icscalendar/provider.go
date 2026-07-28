package icscalendar

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Provider implements calendarprovider.Provider for ICS feeds.
// accessToken is the normalized feed URL (opaque credential from sealed secrets).
type Provider struct {
	client *Client
	name   string
}

// NewProvider constructs an ICS calendar provider.
func NewProvider(client *Client) *Provider {
	if client == nil {
		client = NewClient(Config{})
	}
	return &Provider{client: client, name: constant.AuthProviderICS}
}

// Name returns the provider key stored on connected_accounts.provider.
func (p *Provider) Name() string { return p.name }

// ListCalendars returns one synthetic calendar for the ICS feed.
func (p *Provider) ListCalendars(
	ctx context.Context,
	accessToken string,
	opts calendarprovider.ListCalendarsOptions,
) (calendarprovider.ListCalendarsResult, error) {
	feedURL := strings.TrimSpace(accessToken)
	if feedURL == "" {
		return calendarprovider.ListCalendarsResult{}, fmt.Errorf("%w: missing ics feed url", errEmptyURL)
	}
	etag, lastMod := DecodeSyncCursor(opts.SyncToken)
	fetched, err := p.client.Fetch(ctx, feedURL, etag, lastMod)
	if err != nil {
		var auth *AuthError
		if errors.As(err, &auth) {
			return calendarprovider.ListCalendarsResult{}, &calendarprovider.AuthError{Status: auth.Status, Body: auth.Body}
		}
		return calendarprovider.ListCalendarsResult{}, err
	}

	calID := FeedCalendarID(feedURL)
	next := EncodeSyncCursor(firstNonEmpty(fetched.ETag, etag), firstNonEmpty(fetched.LastModified, lastMod))

	if fetched.NotModified {
		return calendarprovider.ListCalendarsResult{
			Calendars: []calendarprovider.RemoteCalendar{{
				ID:         calID,
				Name:       "ICS Calendar",
				Primary:    true,
				Writable:   false,
				AccessRole: "reader",
				ETag:       firstNonEmpty(fetched.ETag, etag),
			}},
			NextSyncToken: next,
			Incremental:   true,
		}, nil
	}

	calName, _, parseErr := ParseCalendar(fetched.Body, "")
	if parseErr != nil {
		return calendarprovider.ListCalendarsResult{}, fmt.Errorf("parse ics calendar: %w", parseErr)
	}

	return calendarprovider.ListCalendarsResult{
		Calendars: []calendarprovider.RemoteCalendar{{
			ID:          calID,
			Name:        calName,
			Primary:     true,
			Writable:    false,
			AccessRole:  "reader",
			ETag:        fetched.ETag,
			Description: "ICS feed",
			Selected:    true,
			Raw: map[string]any{
				"last_modified": fetched.LastModified,
			},
		}},
		NextSyncToken: next,
		Incremental:   false,
	}, nil
}

// ListEvents fetches (or reuses conditional GET) and maps VEVENTs.
func (p *Provider) ListEvents(
	ctx context.Context,
	accessToken string,
	calendarID string,
	opts calendarprovider.ListEventsOptions,
) (calendarprovider.ListEventsResult, error) {
	feedURL := strings.TrimSpace(accessToken)
	if feedURL == "" {
		return calendarprovider.ListEventsResult{}, fmt.Errorf("%w: missing ics feed url", errEmptyURL)
	}
	expectedID := FeedCalendarID(feedURL)
	if calendarID != "" && calendarID != expectedID {
		return calendarprovider.ListEventsResult{}, fmt.Errorf("ics calendar id mismatch")
	}

	etag, lastMod := DecodeSyncCursor(opts.SyncToken)
	fetched, err := p.client.Fetch(ctx, feedURL, etag, lastMod)
	if err != nil {
		var auth *AuthError
		if errors.As(err, &auth) {
			return calendarprovider.ListEventsResult{}, &calendarprovider.AuthError{Status: auth.Status, Body: auth.Body}
		}
		return calendarprovider.ListEventsResult{}, err
	}

	next := EncodeSyncCursor(firstNonEmpty(fetched.ETag, etag), firstNonEmpty(fetched.LastModified, lastMod))
	if fetched.NotModified {
		return calendarprovider.ListEventsResult{
			Events:        nil,
			NextSyncToken: next,
			Incremental:   true,
			ReplaceAll:    false,
		}, nil
	}

	_, events, parseErr := ParseCalendar(fetched.Body, "")
	if parseErr != nil {
		return calendarprovider.ListEventsResult{}, fmt.Errorf("parse ics events: %w", parseErr)
	}

	from, to := opts.TimeMin, opts.TimeMax
	if from.IsZero() || to.IsZero() {
		now := time.Now().UTC()
		from = now.Add(-constant.CalendarEventSyncLookback)
		to = now.Add(constant.CalendarEventSyncLookahead)
	}
	events = ExpandRecurring(events, from, to)

	return calendarprovider.ListEventsResult{
		Events:        events,
		NextSyncToken: next,
		Incremental:   false,
		ReplaceAll:    true,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
