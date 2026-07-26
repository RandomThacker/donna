package icscalendar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/icscalendar"
)

const sampleICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Donna//Test//EN
X-WR-CALNAME:Work Feed
BEGIN:VEVENT
UID:evt-1@donna.test
DTSTART;TZID=America/New_York:20260701T100000
DTEND;TZID=America/New_York:20260701T110000
SUMMARY:Standup
DESCRIPTION:Daily sync
LOCATION:Zoom
STATUS:CONFIRMED
ORGANIZER;CN=Alex:mailto:alex@example.com
ATTENDEE;CN=Sam;PARTSTAT=ACCEPTED:mailto:sam@example.com
RRULE:FREQ=WEEKLY;BYDAY=MO
EXDATE;TZID=America/New_York:20260708T100000
CATEGORIES:Work,Team
URL:https://example.com/meet
X-CUSTOM-PROP:keep-me
END:VEVENT
BEGIN:VEVENT
UID:evt-2@donna.test
DTSTART;VALUE=DATE:20260704
DTEND;VALUE=DATE:20260705
SUMMARY:Holiday
STATUS:CANCELLED
END:VEVENT
END:VCALENDAR
`

func TestNormalizeFeedURL(t *testing.T) {
	t.Parallel()
	got, err := icscalendar.NormalizeFeedURL("webcal://calendar.example.com/feed.ics")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://calendar.example.com/feed.ics" {
		t.Fatalf("got %q", got)
	}
	id1 := icscalendar.FeedCalendarID(got)
	id2 := icscalendar.FeedCalendarID(got)
	if id1 != id2 || !strings.HasPrefix(id1, "ics_") {
		t.Fatalf("deterministic id failed: %s %s", id1, id2)
	}
}

func TestParseCalendarRFC5545(t *testing.T) {
	t.Parallel()
	name, events, err := icscalendar.ParseCalendar([]byte(sampleICS), "Fallback")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Work Feed" {
		t.Fatalf("calendar name = %q", name)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 events, got %d", len(events))
	}
	evt := events[0]
	if evt.ID != "evt-1@donna.test" {
		t.Fatalf("uid = %q", evt.ID)
	}
	if evt.Title != "Standup" || evt.Location != "Zoom" {
		t.Fatalf("mapped fields: %+v", evt)
	}
	if evt.OrganizerEmail != "alex@example.com" {
		t.Fatalf("organizer = %q", evt.OrganizerEmail)
	}
	if len(evt.Attendees) != 1 || evt.Attendees[0].Email != "sam@example.com" {
		t.Fatalf("attendees = %+v", evt.Attendees)
	}
	if len(evt.Recurrence) != 1 || !strings.Contains(evt.Recurrence[0], "FREQ=WEEKLY") {
		t.Fatalf("recurrence = %+v", evt.Recurrence)
	}
	if evt.Raw["exdate"] == nil {
		t.Fatal("expected exdate in raw")
	}
	extras, ok := evt.Raw["unsupported_properties"].(map[string]string)
	if !ok || extras["X-CUSTOM-PROP"] != "keep-me" {
		t.Fatalf("unsupported props = %#v", evt.Raw["unsupported_properties"])
	}
	if events[1].Status != constant.CalendarEventStatusCancelled || !events[1].Deleted {
		t.Fatalf("cancelled mapping failed: %+v", events[1])
	}
	if !events[1].IsAllDay {
		t.Fatal("all-day expected")
	}
}

func TestParseMalformedFeed(t *testing.T) {
	t.Parallel()
	_, _, err := icscalendar.ParseCalendar([]byte("not a calendar"), "")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestEncodeDecodeSyncCursor(t *testing.T) {
	t.Parallel()
	token := icscalendar.EncodeSyncCursor(`"abc"`, "Wed, 01 Jul 2026 00:00:00 GMT")
	etag, lm := icscalendar.DecodeSyncCursor(token)
	if etag != `"abc"` || !strings.Contains(lm, "2026") {
		t.Fatalf("etag=%q lm=%q", etag, lm)
	}
}

func TestClientETagNotModified(t *testing.T) {
	t.Parallel()
	var sawIfNoneMatch string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawIfNoneMatch = r.Header.Get("If-None-Match")
		if sawIfNoneMatch != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		w.Header().Set("Last-Modified", "Wed, 01 Jul 2026 00:00:00 GMT")
		_, _ = w.Write([]byte(sampleICS))
	}))
	defer srv.Close()

	client := icscalendar.NewClient(icscalendar.Config{})
	first, err := client.Fetch(context.Background(), srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.ETag != `"v1"` || first.NotModified {
		t.Fatalf("first fetch: %+v", first)
	}
	second, err := client.Fetch(context.Background(), srv.URL, first.ETag, first.LastModified)
	if err != nil {
		t.Fatal(err)
	}
	if !second.NotModified {
		t.Fatal("expected 304 not modified")
	}
	if sawIfNoneMatch != `"v1"` {
		t.Fatalf("If-None-Match = %q", sawIfNoneMatch)
	}
}

func TestClientLastModified(t *testing.T) {
	t.Parallel()
	var sawIMS string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawIMS = r.Header.Get("If-Modified-Since")
		if sawIMS != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Last-Modified", "Wed, 01 Jul 2026 12:00:00 GMT")
		_, _ = w.Write([]byte(sampleICS))
	}))
	defer srv.Close()

	client := icscalendar.NewClient(icscalendar.Config{})
	first, err := client.Fetch(context.Background(), srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.Fetch(context.Background(), srv.URL, "", first.LastModified)
	if err != nil {
		t.Fatal(err)
	}
	if !second.NotModified || sawIMS == "" {
		t.Fatalf("last-modified handling failed: notModified=%v ims=%q", second.NotModified, sawIMS)
	}
}

func TestProviderListEventsReplaceAllAndIdempotent(t *testing.T) {
	t.Parallel()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("ETag", `"feed-1"`)
		_, _ = w.Write([]byte(sampleICS))
	}))
	defer srv.Close()

	p := icscalendar.NewProvider(icscalendar.NewClient(icscalendar.Config{}))
	window := calendarprovider.ListEventsOptions{
		TimeMin: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		TimeMax: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}
	first, err := p.ListEvents(context.Background(), srv.URL, icscalendar.FeedCalendarID(srv.URL), window)
	if err != nil {
		t.Fatal(err)
	}
	if !first.ReplaceAll || len(first.Events) < 2 {
		t.Fatalf("first result len=%d replaceAll=%v", len(first.Events), first.ReplaceAll)
	}
	// Weekly standup should expand into multiple July Mondays.
	standup := 0
	for _, event := range first.Events {
		if strings.HasPrefix(event.ID, "evt-1@donna.test_") {
			standup++
		}
	}
	if standup < 3 {
		t.Fatalf("expected expanded standups, got %d", standup)
	}
	second, err := p.ListEvents(context.Background(), srv.URL, icscalendar.FeedCalendarID(srv.URL), calendarprovider.ListEventsOptions{
		SyncToken: first.NextSyncToken,
		TimeMin:   window.TimeMin,
		TimeMax:   window.TimeMax,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !second.ReplaceAll || len(second.Events) != len(first.Events) {
		t.Fatalf("second result: len=%d want=%d", len(second.Events), len(first.Events))
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestExpandRecurringWeekly(t *testing.T) {
	t.Parallel()
	_, events, err := icscalendar.ParseCalendar([]byte(sampleICS), "")
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expanded := icscalendar.ExpandRecurring(events, from, to)
	var mondays []time.Time
	for _, event := range expanded {
		if !strings.HasPrefix(event.ID, "evt-1@donna.test_") {
			continue
		}
		if event.Recurrence != nil {
			t.Fatal("expanded instances must not carry RRULE")
		}
		mondays = append(mondays, event.StartsAt)
	}
	if len(mondays) < 3 {
		t.Fatalf("expected multiple mondays, got %v", mondays)
	}
}

func TestWindowsTimezoneIndia(t *testing.T) {
	t.Parallel()
	icsBody := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:team@donna.test
DTSTART;TZID=India Standard Time:20260727T180000
DTEND;TZID=India Standard Time:20260727T183000
SUMMARY:Team Connect
END:VEVENT
END:VCALENDAR`
	_, events, err := icscalendar.ParseCalendar([]byte(icsBody), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events=%d", len(events))
	}
	// 18:00 IST = 12:30 UTC
	if events[0].StartsAt.UTC().Hour() != 12 || events[0].StartsAt.UTC().Minute() != 30 {
		t.Fatalf("starts_at=%s want 12:30 UTC", events[0].StartsAt.UTC())
	}
}


func TestProviderNotModifiedSkipsReplaceAll(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(sampleICS))
	}))
	defer srv.Close()

	p := icscalendar.NewProvider(icscalendar.NewClient(icscalendar.Config{}))
	first, err := p.ListEvents(context.Background(), srv.URL, "", calendarprovider.ListEventsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.ListEvents(context.Background(), srv.URL, "", calendarprovider.ListEventsOptions{
		SyncToken: first.NextSyncToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ReplaceAll || len(second.Events) != 0 || !second.Incremental {
		t.Fatalf("304 handling: %+v", second)
	}
}

func TestTimezoneParsing(t *testing.T) {
	t.Parallel()
	_, events, err := icscalendar.ParseCalendar([]byte(sampleICS), "")
	if err != nil {
		t.Fatal(err)
	}
	start := events[0].StartsAt
	if start.Location() != time.UTC {
		t.Fatalf("stored as UTC expected, got %v", start.Location())
	}
	// 10:00 America/New_York in July is UTC-4 → 14:00 UTC
	if start.Hour() != 14 {
		t.Fatalf("expected 14:00 UTC, got %s", start)
	}
}
