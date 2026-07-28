package microsoftcalendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/microsoftcalendar"
)

func TestListCalendarsPaginatesAndMaps(t *testing.T) {
	var srv *httptest.Server
	page := 0
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/me/calendars" {
			http.NotFound(w, r)
			return
		}
		page++
		if page == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"value": []map[string]any{
					{
						"id":                "cal-1",
						"name":              "Personal",
						"isDefaultCalendar": true,
						"canEdit":           true,
						"hexColor":          "#ff0000",
						"changeKey":         "ck-1",
					},
				},
				"@odata.nextLink": srv.URL + "/me/calendars?$skiptoken=2",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{
					"id":                "cal-2",
					"name":              "Shared",
					"isDefaultCalendar": false,
					"canEdit":           false,
					"hexColor":          "#00ff00",
					"changeKey":         "ck-2",
				},
			},
		})
	}))
	defer srv.Close()

	client := microsoftcalendar.NewClient(microsoftcalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	result, err := client.ListCalendars(context.Background(), "access", microsoftcalendar.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Calendars) != 2 {
		t.Fatalf("calendars = %#v", result.Calendars)
	}
	if !result.Calendars[0].Primary || !result.Calendars[0].Writable || result.Calendars[0].Color != "#ff0000" {
		t.Fatalf("cal0 = %#v", result.Calendars[0])
	}
	if result.Calendars[0].ETag != "ck-1" {
		t.Fatalf("etag = %q", result.Calendars[0].ETag)
	}
	if result.Calendars[1].Writable || result.Calendars[1].Primary {
		t.Fatalf("cal1 = %#v", result.Calendars[1])
	}
	if result.Incremental || result.NextSyncToken != "" {
		t.Fatalf("expected full non-incremental list, got %#v", result)
	}
}

func TestListCalendarsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer srv.Close()

	client := microsoftcalendar.NewClient(microsoftcalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	_, err := client.ListCalendars(context.Background(), "access", microsoftcalendar.ListOptions{})
	var authErr *microsoftcalendar.AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("err = %v", err)
	}
}

func TestListEventsDeltaAndGone(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch {
		case r.URL.Path == "/me/calendars/cal-1/calendarView/delta":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"value": []map[string]any{
					{
						"id":      "evt-1",
						"subject": "Lunch",
						"body":    map[string]any{"content": "Soup"},
						"location": map[string]any{
							"displayName": "Cafe",
						},
						"isAllDay":    false,
						"isCancelled": false,
						"start": map[string]any{
							"dateTime": "2026-07-25T12:00:00.0000000",
							"timeZone": "UTC",
						},
						"end": map[string]any{
							"dateTime": "2026-07-25T13:00:00.0000000",
							"timeZone": "UTC",
						},
						"organizer": map[string]any{
							"emailAddress": map[string]any{
								"address": "host@example.com",
								"name":    "Host",
							},
						},
						"attendees": []map[string]any{
							{
								"type": "required",
								"status": map[string]any{
									"response": "accepted",
								},
								"emailAddress": map[string]any{
									"address": "guest@example.com",
									"name":    "Guest",
								},
							},
						},
						"onlineMeeting": map[string]any{
							"joinUrl": "https://teams.example/join/1",
						},
						"seriesMasterId":       "series-1",
						"lastModifiedDateTime": "2026-07-24T10:00:00Z",
						"changeKey":            "etag-1",
					},
					{
						"id": "evt-gone",
						"@removed": map[string]any{
							"reason": "deleted",
						},
					},
				},
				"@odata.deltaLink": srv.URL + "/delta?token=next",
			})
		case r.URL.Path == "/delta":
			http.Error(w, "Gone", http.StatusGone)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := microsoftcalendar.NewClient(microsoftcalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})

	result, err := client.ListEvents(context.Background(), "access", "cal-1", microsoftcalendar.EventListOptions{
		TimeMin: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TimeMax: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Incremental {
		t.Fatal("full delta should not be marked incremental")
	}
	if result.NextSyncToken != srv.URL+"/delta?token=next" {
		t.Fatalf("NextSyncToken = %q", result.NextSyncToken)
	}
	if len(result.Events) != 2 {
		t.Fatalf("events = %#v", result.Events)
	}
	ev := result.Events[0]
	if ev.Title != "Lunch" || ev.Description != "Soup" || ev.Location != "Cafe" {
		t.Fatalf("event fields = %#v", ev)
	}
	if ev.OnlineMeetingURL != "https://teams.example/join/1" {
		t.Fatalf("OnlineMeetingURL = %q", ev.OnlineMeetingURL)
	}
	if ev.OrganizerEmail != "host@example.com" || ev.RecurringEventID != "series-1" {
		t.Fatalf("organizer/series = %#v", ev)
	}
	if len(ev.Attendees) != 1 || ev.Attendees[0].Email != "guest@example.com" || ev.Attendees[0].ResponseStatus != "accepted" {
		t.Fatalf("attendees = %#v", ev.Attendees)
	}
	wantStart := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	if !ev.StartsAt.Equal(wantStart) || !ev.EndsAt.Equal(wantStart.Add(time.Hour)) {
		t.Fatalf("start/end = %v / %v", ev.StartsAt, ev.EndsAt)
	}
	if !result.Events[1].Deleted || result.Events[1].Status != "cancelled" {
		t.Fatalf("removed event = %#v", result.Events[1])
	}

	_, err = client.ListEvents(context.Background(), "access", "cal-1", microsoftcalendar.EventListOptions{
		SyncToken: result.NextSyncToken,
	})
	var gone *microsoftcalendar.GoneError
	if !errors.As(err, &gone) {
		t.Fatalf("err = %v", err)
	}
}

func TestProviderNameAndMapsSyncCursorInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Gone", http.StatusGone)
	}))
	defer srv.Close()

	provider := microsoftcalendar.NewProvider(microsoftcalendar.NewClient(microsoftcalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	}))
	if provider.Name() != constant.AuthProviderMicrosoft || provider.Name() != "microsoft" {
		t.Fatalf("Name = %q", provider.Name())
	}

	_, err := provider.ListEvents(context.Background(), "access", "cal-1", calendarprovider.ListEventsOptions{
		SyncToken: srv.URL + "/delta",
	})
	var invalid *calendarprovider.SyncCursorInvalidError
	if !errors.As(err, &invalid) {
		t.Fatalf("err = %v", err)
	}
}
