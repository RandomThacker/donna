package googlecalendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
)

func TestListCalendarsPaginatesAndMapsWritable(t *testing.T) {
	mux := http.NewServeMux()
	page := 0
	mux.HandleFunc("/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		page++
		if page == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{
						"id":         "primary",
						"summary":    "Personal",
						"primary":    true,
						"accessRole": "owner",
						"etag":       "e1",
						"timeZone":   "UTC",
					},
				},
				"nextPageToken": "page-2",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"id":         "reader-cal",
					"summary":    "Shared",
					"accessRole": "reader",
					"etag":       "e2",
				},
			},
			"nextSyncToken": "sync-abc",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	result, err := client.ListCalendars(context.Background(), "access", googlecalendar.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Calendars) != 2 {
		t.Fatalf("calendars = %#v", result.Calendars)
	}
	if !result.Calendars[0].Writable || result.Calendars[1].Writable {
		t.Fatalf("writable flags = %#v", result.Calendars)
	}
	if result.NextSyncToken != "sync-abc" {
		t.Fatalf("nextSyncToken = %q", result.NextSyncToken)
	}
	if result.Incremental {
		t.Fatal("full sync should not be incremental")
	}
}

func TestListCalendarsIncrementalUsesSyncToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("syncToken") != "tok-1" {
			http.Error(w, "missing syncToken", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "primary", "summary": "Personal", "accessRole": "owner", "deleted": true},
			},
			"nextSyncToken": "tok-2",
		})
	}))
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	result, err := client.ListCalendars(context.Background(), "access", googlecalendar.ListOptions{SyncToken: "tok-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Incremental || !result.Calendars[0].Deleted || result.NextSyncToken != "tok-2" {
		t.Fatalf("result = %#v", result)
	}
}

func TestListCalendarsGoneError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Gone", http.StatusGone)
	}))
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	_, err := client.ListCalendars(context.Background(), "access", googlecalendar.ListOptions{SyncToken: "old"})
	var gone *googlecalendar.GoneError
	if !errors.As(err, &gone) {
		t.Fatalf("err = %v", err)
	}
}

func TestListCalendarsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer srv.Close()

	client := googlecalendar.NewClient(googlecalendar.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	_, err := client.ListCalendars(context.Background(), "access", googlecalendar.ListOptions{})
	var authErr *googlecalendar.AuthError
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.As(err, &authErr) {
		t.Fatalf("err = %v", err)
	}
}
