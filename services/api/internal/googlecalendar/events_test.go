package googlecalendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
)

func TestListEventsMapsFieldsAndSyncToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/calendars/primary/events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("syncToken") != "" {
			http.Error(w, "unexpected syncToken", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"id":      "evt-1",
					"summary": "Lunch",
					"status":  "confirmed",
					"start":   map[string]any{"dateTime": "2026-07-25T12:00:00Z", "timeZone": "UTC"},
					"end":     map[string]any{"dateTime": "2026-07-25T13:00:00Z", "timeZone": "UTC"},
					"organizer": map[string]any{
						"email": "host@example.com", "displayName": "Host", "self": true,
					},
					"attendees": []map[string]any{
						{"email": "guest@example.com", "responseStatus": "accepted"},
					},
					"updated": "2026-07-24T10:00:00Z",
					"etag":    "etag-1",
				},
				{
					"id":      "evt-allday",
					"summary": "Holiday",
					"status":  "confirmed",
					"start":   map[string]any{"date": "2026-07-26"},
					"end":     map[string]any{"date": "2026-07-27"},
				},
			},
			"nextSyncToken": "events-token-1",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	result, err := client.ListEvents(context.Background(), "access", "primary", googlecalendar.EventListOptions{
		TimeMin: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TimeMax: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.NextSyncToken != "events-token-1" || len(result.Events) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if result.Events[0].Title != "Lunch" || result.Events[0].OrganizerEmail != "host@example.com" {
		t.Fatalf("event0 = %#v", result.Events[0])
	}
	if !result.Events[1].IsAllDay {
		t.Fatal("expected all-day event")
	}
}

func TestListEventsGoneError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Gone", http.StatusGone)
	}))
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	_, err := client.ListEvents(context.Background(), "access", "primary", googlecalendar.EventListOptions{SyncToken: "old"})
	var gone *googlecalendar.GoneError
	if !errors.As(err, &gone) {
		t.Fatalf("err = %v", err)
	}
}

func TestListEventsInvalidSyncTokenAsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid sync token value.","code":400}}`))
	}))
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{BaseURL: srv.URL, HTTPClient: srv.Client()})
	_, err := client.ListEvents(context.Background(), "access", "primary", googlecalendar.EventListOptions{SyncToken: "etag-not-token"})
	var gone *googlecalendar.GoneError
	if !errors.As(err, &gone) {
		t.Fatalf("err = %v", err)
	}
}
