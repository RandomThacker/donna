package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// EventListOptions controls events.list pagination / incremental sync.
// When SyncToken is set, Google forbids most other filters (timeMin/timeMax/etc).
type EventListOptions struct {
	SyncToken string
	TimeMin   time.Time // full sync window start (ignored with SyncToken)
	TimeMax   time.Time // full sync window end (ignored with SyncToken)
}

// EventListResult is a pageable events.list response.
type EventListResult struct {
	Events        []RemoteEvent
	NextSyncToken string
	Incremental   bool
}

// RemoteEvent is a Google Calendar event Donna persists.
type RemoteEvent struct {
	ID                   string
	ETag                 string
	Status               string
	Title                string
	Description          string
	Location             string
	StartsAt             time.Time
	EndsAt               time.Time
	IsAllDay             bool
	Timezone             string
	Visibility           string
	OrganizerEmail       string
	OrganizerDisplayName string
	OrganizerSelf        bool
	Attendees            []RemoteAttendee
	Recurrence           []string
	RecurringEventID     string
	UpdatedAt            time.Time
	Deleted              bool
	Raw                  map[string]any
}

// RemoteAttendee is a thin attendee projection.
type RemoteAttendee struct {
	Email          string `json:"email,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	ResponseStatus string `json:"response_status,omitempty"`
	Organizer      bool   `json:"organizer,omitempty"`
	Self           bool   `json:"self,omitempty"`
}

// ListEvents fetches events for one calendar (paginated).
// HTTP 410 means the syncToken is invalid — callers should clear it and full-sync.
func (c *Client) ListEvents(ctx context.Context, accessToken, calendarID string, opts EventListOptions) (EventListResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	calendarID = strings.TrimSpace(calendarID)
	if accessToken == "" {
		return EventListResult{}, fmt.Errorf("access token is required")
	}
	if calendarID == "" {
		return EventListResult{}, fmt.Errorf("calendar id is required")
	}

	syncToken := strings.TrimSpace(opts.SyncToken)
	var (
		out       EventListResult
		pageToken string
	)
	out.Incremental = syncToken != ""

	encodedCal := url.PathEscape(calendarID)

	for {
		endpoint := c.cfg.BaseURL + "/calendars/" + encodedCal + "/events"
		q := url.Values{}
		q.Set("maxResults", "250")
		if syncToken != "" {
			q.Set("syncToken", syncToken)
		} else {
			q.Set("singleEvents", "false")
			q.Set("showDeleted", "true")
			if !opts.TimeMin.IsZero() {
				q.Set("timeMin", opts.TimeMin.UTC().Format(time.RFC3339))
			}
			if !opts.TimeMax.IsZero() {
				q.Set("timeMax", opts.TimeMax.UTC().Format(time.RFC3339))
			}
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
		if err != nil {
			return EventListResult{}, fmt.Errorf("events.list request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			return EventListResult{}, fmt.Errorf("events.list: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return EventListResult{}, fmt.Errorf("read events.list: %w", readErr)
		}
		if resp.StatusCode == http.StatusGone || (syncToken != "" && isInvalidSyncTokenResponse(resp.StatusCode, body)) {
			return EventListResult{}, &GoneError{Body: truncate(string(body), 500)}
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return EventListResult{}, &AuthError{Status: resp.StatusCode, Body: truncate(string(body), 500)}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return EventListResult{}, fmt.Errorf("events.list status %d: %s", resp.StatusCode, truncate(string(body), 500))
		}

		var raw struct {
			Items         []map[string]any `json:"items"`
			NextPageToken string           `json:"nextPageToken"`
			NextSyncToken string           `json:"nextSyncToken"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return EventListResult{}, fmt.Errorf("decode events.list: %w", err)
		}
		for _, item := range raw.Items {
			ev, ok := mapRemoteEvent(item)
			if !ok {
				continue
			}
			out.Events = append(out.Events, ev)
		}
		if raw.NextSyncToken != "" {
			out.NextSyncToken = raw.NextSyncToken
		}
		if raw.NextPageToken == "" {
			break
		}
		pageToken = raw.NextPageToken
	}

	return out, nil
}

func mapRemoteEvent(item map[string]any) (RemoteEvent, bool) {
	id := stringField(item, "id")
	if id == "" {
		return RemoteEvent{}, false
	}
	status := strings.ToLower(stringField(item, "status"))
	if status == "" {
		status = "confirmed"
	}
	deleted := boolField(item, "deleted") || status == "cancelled"

	start, allDay, tz := mapEventInstant(mapField(item, "start"))
	end, _, endTZ := mapEventInstant(mapField(item, "end"))
	if tz == "" {
		tz = endTZ
	}
	if start.IsZero() {
		start = time.Unix(0, 0).UTC()
	}
	if end.IsZero() || end.Before(start) {
		end = start
	}

	title := stringField(item, "summary")
	if title == "" {
		title = "(untitled)"
	}

	org := mapField(item, "organizer")
	attendees := mapAttendees(item["attendees"])

	var recurrence []string
	if rawRec, ok := item["recurrence"].([]any); ok {
		for _, r := range rawRec {
			if s, ok := r.(string); ok && strings.TrimSpace(s) != "" {
				recurrence = append(recurrence, s)
			}
		}
	}

	updatedAt, _ := time.Parse(time.RFC3339, stringField(item, "updated"))

	return RemoteEvent{
		ID:                   id,
		ETag:                 stringField(item, "etag"),
		Status:               status,
		Title:                title,
		Description:          stringField(item, "description"),
		Location:             stringField(item, "location"),
		StartsAt:             start,
		EndsAt:               end,
		IsAllDay:             allDay,
		Timezone:             tz,
		Visibility:           stringField(item, "visibility"),
		OrganizerEmail:       stringField(org, "email"),
		OrganizerDisplayName: stringField(org, "displayName"),
		OrganizerSelf:        boolField(org, "self"),
		Attendees:            attendees,
		Recurrence:           recurrence,
		RecurringEventID:     stringField(item, "recurringEventId"),
		UpdatedAt:            updatedAt.UTC(),
		Deleted:              deleted,
		Raw:                  item,
	}, true
}

func mapEventInstant(m map[string]any) (time.Time, bool, string) {
	if m == nil {
		return time.Time{}, false, ""
	}
	tz := stringField(m, "timeZone")
	if dateOnly := stringField(m, "date"); dateOnly != "" {
		t, err := time.ParseInLocation("2006-01-02", dateOnly, time.UTC)
		if err != nil {
			return time.Time{}, true, tz
		}
		return t.UTC(), true, tz
	}
	if dt := stringField(m, "dateTime"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			return t.UTC(), false, tz
		}
		if t, err := time.Parse(time.RFC3339Nano, dt); err == nil {
			return t.UTC(), false, tz
		}
	}
	return time.Time{}, false, tz
}

func mapField(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	nested, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return nested
}

func mapAttendees(raw any) []RemoteAttendee {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	const maxAttendees = 50
	out := make([]RemoteAttendee, 0, len(arr))
	for _, item := range arr {
		if len(out) >= maxAttendees {
			break
		}
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		email := stringField(m, "email")
		if email == "" {
			continue
		}
		out = append(out, RemoteAttendee{
			Email:          email,
			DisplayName:    stringField(m, "displayName"),
			ResponseStatus: stringField(m, "responseStatus"),
			Organizer:      boolField(m, "organizer"),
			Self:           boolField(m, "self"),
		})
	}
	return out
}

// isInvalidSyncTokenResponse detects Google's 400 "Invalid sync token value" (and similar),
// which occurs when a stored cursor is stale or was never an events.list syncToken (e.g. etag).
func isInvalidSyncTokenResponse(status int, body []byte) bool {
	if status != http.StatusBadRequest {
		return false
	}
	lower := strings.ToLower(string(body))
	return strings.Contains(lower, "sync token") || strings.Contains(lower, "synctoken")
}
