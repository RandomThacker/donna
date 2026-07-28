package microsoftcalendar

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

// EventListOptions controls calendarView/delta pagination / incremental sync.
// When SyncToken is set it holds a prior @odata.deltaLink (absolute URL or path+query).
type EventListOptions struct {
	SyncToken string
	TimeMin   time.Time // full sync window start (ignored with SyncToken)
	TimeMax   time.Time // full sync window end (ignored with SyncToken)
}

// EventListResult is a pageable calendarView/delta response.
type EventListResult struct {
	Events        []RemoteEvent
	NextSyncToken string
	Incremental   bool
}

// RemoteEvent is a Microsoft Graph calendar event Donna persists.
type RemoteEvent struct {
	ID                   string
	ETag                 string
	Status               string
	Title                string
	Description          string
	Location             string
	OnlineMeetingURL     string
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

// ListEvents fetches events for one calendar via calendarView/delta (paginated).
// SyncToken stores the next delta URL (absolute or relative). HTTP 410 means the
// cursor is invalid — callers should clear it and full-sync.
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
	var out EventListResult
	out.Incremental = syncToken != ""

	nextURL, err := c.resolveEventsURL(calendarID, opts)
	if err != nil {
		return EventListResult{}, err
	}

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return EventListResult{}, fmt.Errorf("events.delta request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Prefer", `odata.maxpagesize=50`)

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			return EventListResult{}, fmt.Errorf("events.delta: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return EventListResult{}, fmt.Errorf("read events.delta: %w", readErr)
		}
		if resp.StatusCode == http.StatusGone {
			return EventListResult{}, &GoneError{Body: truncate(string(body), 500)}
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return EventListResult{}, &AuthError{Status: resp.StatusCode, Body: truncate(string(body), 500)}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return EventListResult{}, fmt.Errorf("events.delta status %d: %s", resp.StatusCode, truncate(string(body), 500))
		}

		var raw struct {
			Value     []map[string]any `json:"value"`
			NextLink  string           `json:"@odata.nextLink"`
			DeltaLink string           `json:"@odata.deltaLink"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return EventListResult{}, fmt.Errorf("decode events.delta: %w", err)
		}
		for _, item := range raw.Value {
			ev, ok := mapRemoteEvent(item)
			if !ok {
				continue
			}
			out.Events = append(out.Events, ev)
		}
		if raw.DeltaLink != "" {
			out.NextSyncToken = raw.DeltaLink
		}
		nextURL = strings.TrimSpace(raw.NextLink)
	}

	return out, nil
}

func (c *Client) resolveEventsURL(calendarID string, opts EventListOptions) (string, error) {
	syncToken := strings.TrimSpace(opts.SyncToken)
	if syncToken != "" {
		if strings.HasPrefix(syncToken, "http://") || strings.HasPrefix(syncToken, "https://") {
			return syncToken, nil
		}
		if strings.HasPrefix(syncToken, "/") {
			return c.cfg.BaseURL + syncToken, nil
		}
		return c.cfg.BaseURL + "/" + strings.TrimLeft(syncToken, "/"), nil
	}

	if opts.TimeMin.IsZero() || opts.TimeMax.IsZero() {
		return "", fmt.Errorf("timeMin and timeMax are required for full event sync")
	}
	encodedCal := url.PathEscape(calendarID)
	q := url.Values{}
	q.Set("startDateTime", opts.TimeMin.UTC().Format(time.RFC3339))
	q.Set("endDateTime", opts.TimeMax.UTC().Format(time.RFC3339))
	return c.cfg.BaseURL + "/me/calendars/" + encodedCal + "/calendarView/delta?" + q.Encode(), nil
}

func mapRemoteEvent(item map[string]any) (RemoteEvent, bool) {
	id := stringField(item, "id")
	if id == "" {
		return RemoteEvent{}, false
	}

	removed := mapField(item, "@removed") != nil || item["@removed"] != nil
	cancelled := boolField(item, "isCancelled")
	deleted := removed || cancelled

	status := "confirmed"
	if cancelled || removed {
		status = "cancelled"
	}

	start, allDay, tz := mapEventInstant(mapField(item, "start"), boolField(item, "isAllDay"))
	end, _, endTZ := mapEventInstant(mapField(item, "end"), boolField(item, "isAllDay"))
	if tz == "" {
		tz = endTZ
	}
	if start.IsZero() {
		start = time.Unix(0, 0).UTC()
	}
	if end.IsZero() || end.Before(start) {
		end = start
	}

	title := stringField(item, "subject")
	if title == "" {
		title = "(untitled)"
	}

	body := mapField(item, "body")
	description := ""
	if body != nil {
		description = stringField(body, "content")
	}

	loc := mapField(item, "location")
	location := ""
	if loc != nil {
		location = stringField(loc, "displayName")
	}

	orgWrap := mapField(item, "organizer")
	orgEmail := mapField(orgWrap, "emailAddress")
	attendees := mapAttendees(item["attendees"])

	updatedAt := parseGraphTime(stringField(item, "lastModifiedDateTime"))

	return RemoteEvent{
		ID:                   id,
		ETag:                 stringField(item, "changeKey"),
		Status:               status,
		Title:                title,
		Description:          description,
		Location:             location,
		OnlineMeetingURL:     firstNonEmpty(onlineMeetingJoinURL(item), stringField(item, "onlineMeetingUrl")),
		StartsAt:             start,
		EndsAt:               end,
		IsAllDay:             allDay,
		Timezone:             tz,
		Visibility:           stringField(item, "sensitivity"),
		OrganizerEmail:       stringField(orgEmail, "address"),
		OrganizerDisplayName: stringField(orgEmail, "name"),
		OrganizerSelf:        false,
		Attendees:            attendees,
		RecurringEventID:     stringField(item, "seriesMasterId"),
		UpdatedAt:            updatedAt,
		Deleted:              deleted,
		Raw:                  item,
	}, true
}

func onlineMeetingJoinURL(item map[string]any) string {
	om := mapField(item, "onlineMeeting")
	if om == nil {
		return ""
	}
	return stringField(om, "joinUrl")
}

func mapEventInstant(m map[string]any, isAllDay bool) (time.Time, bool, string) {
	if m == nil {
		return time.Time{}, isAllDay, ""
	}
	tz := stringField(m, "timeZone")
	dt := stringField(m, "dateTime")
	if dt == "" {
		return time.Time{}, isAllDay, tz
	}

	// All-day Graph values are often date-only or midnight local without offset.
	if isAllDay {
		datePart := dt
		if len(datePart) > 10 {
			datePart = datePart[:10]
		}
		if t, err := time.ParseInLocation("2006-01-02", datePart, time.UTC); err == nil {
			return t.UTC(), true, tz
		}
	}

	if t, err := time.Parse(time.RFC3339, dt); err == nil {
		return t.UTC(), isAllDay, tz
	}
	if t, err := time.Parse(time.RFC3339Nano, dt); err == nil {
		return t.UTC(), isAllDay, tz
	}
	// Graph frequently returns "2026-07-25T12:00:00.0000000" without zone.
	layouts := []string{
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, dt, time.UTC); err == nil {
			return t.UTC(), isAllDay, tz
		}
	}
	return time.Time{}, isAllDay, tz
}

func mapField(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
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
		emailAddr := mapField(m, "emailAddress")
		email := stringField(emailAddr, "address")
		if email == "" {
			continue
		}
		status := mapField(m, "status")
		response := stringField(status, "response")
		attendeeType := strings.ToLower(stringField(m, "type"))
		out = append(out, RemoteAttendee{
			Email:          email,
			DisplayName:    stringField(emailAddr, "name"),
			ResponseStatus: mapGraphResponseStatus(response),
			Organizer:      attendeeType == "required" && strings.EqualFold(response, "organizer"),
			Self:           false,
		})
	}
	return out
}

func mapGraphResponseStatus(response string) string {
	switch strings.ToLower(strings.TrimSpace(response)) {
	case "accepted":
		return "accepted"
	case "tentativelyaccepted", "tentatively_accepted":
		return "tentative"
	case "declined":
		return "declined"
	case "organizer":
		return "accepted"
	case "notresponded", "none", "":
		return "needsAction"
	default:
		return strings.ToLower(response)
	}
}

func parseGraphTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC()
	}
	layouts := []string{
		"2006-01-02T15:04:05.9999999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
