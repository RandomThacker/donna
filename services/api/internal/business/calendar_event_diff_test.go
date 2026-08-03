package business

import (
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
)

func TestShouldUpdateWhenETagChanged(t *testing.T) {
	etag1, etag2 := "etag-1", "etag-2"
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "A", StartsAt: start, EndsAt: start.Add(time.Hour),
		Status: constant.CalendarEventStatusConfirmed, ProviderETag: &etag1,
		AttendeesSummary: []byte(`[]`),
	}
	mapped := existing
	mapped.ProviderETag = &etag2
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if skip || reason != eventReasonETagChanged {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldNotTrustETagAloneWhenContentDiffers(t *testing.T) {
	etag := "etag-1"
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "A", StartsAt: start, EndsAt: start.Add(time.Hour),
		Status: constant.CalendarEventStatusConfirmed, ProviderETag: &etag,
		AttendeesSummary: []byte(`[]`),
	}
	mapped := existing
	mapped.Title = "B" // same etag, no updated_at → hash must catch it
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if skip || reason != eventReasonContentHash {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldSkipWhenETagAndUpdatedAtAgree(t *testing.T) {
	etag := "etag-1"
	ts := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "A", StartsAt: start, EndsAt: start.Add(time.Hour),
		Status: constant.CalendarEventStatusConfirmed,
		ProviderETag: &etag, ProviderUpdatedAt: &ts,
		AttendeesSummary: []byte(`[]`),
	}
	mapped := existing
	mapped.Title = "B" // content differs; etag+updated_at agree → skip (reason=etag)
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if !skip || reason != eventReasonETag {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldHashWhenETagSameButUpdatedAtDiffers(t *testing.T) {
	etag := "etag-1"
	ts1 := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	ts2 := ts1.Add(time.Minute)
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "Standup", StartsAt: start, EndsAt: start.Add(time.Hour),
		Status: constant.CalendarEventStatusConfirmed,
		ProviderETag: &etag, ProviderUpdatedAt: &ts1,
		AttendeesSummary: []byte(`[]`),
	}

	sameContent := existing
	sameContent.ProviderUpdatedAt = &ts2
	skip, reason := shouldSkipEventUpdate(existing, sameContent)
	if !skip || reason != eventReasonContentHash {
		t.Fatalf("unchanged content: skip=%v reason=%q", skip, reason)
	}

	changed := existing
	changed.ProviderUpdatedAt = &ts2
	changed.Title = "Standup Renamed"
	skip, reason = shouldSkipEventUpdate(existing, changed)
	if skip || reason != eventReasonContentHash {
		t.Fatalf("changed content: skip=%v reason=%q", skip, reason)
	}
}

func TestShouldSkipEventUpdateByUpdatedAtWhenETagMissing(t *testing.T) {
	ts := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "A", Status: constant.CalendarEventStatusConfirmed, ProviderUpdatedAt: &ts,
	}
	mapped := entity.CalendarEvent{
		Title: "B", Status: constant.CalendarEventStatusConfirmed, ProviderUpdatedAt: &ts,
	}
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if !skip || reason != eventReasonProviderUpdatedAt {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldHashWhenUpdatedAtDiffersAndETagMissing(t *testing.T) {
	ts1 := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	ts2 := ts1.Add(time.Hour)
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	existing := entity.CalendarEvent{
		Title: "Standup", StartsAt: start, EndsAt: start.Add(time.Hour),
		Status: constant.CalendarEventStatusConfirmed, ProviderUpdatedAt: &ts1,
		AttendeesSummary: []byte(`[]`),
	}
	changed := existing
	changed.ProviderUpdatedAt = &ts2
	changed.Title = "Renamed"
	skip, reason := shouldSkipEventUpdate(existing, changed)
	if skip || reason != eventReasonContentHash {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldSkipEventUpdateByHash(t *testing.T) {
	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	existing := entity.CalendarEvent{
		Title: "Standup", StartsAt: start, EndsAt: end,
		Status: constant.CalendarEventStatusConfirmed, AttendeesSummary: []byte(`[]`),
	}
	mapped := entity.CalendarEvent{
		Title: "Standup", StartsAt: start, EndsAt: end,
		Status: constant.CalendarEventStatusConfirmed, AttendeesSummary: []byte(`[]`),
		ProviderPayload: []byte(`{"noise":true}`),
	}
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if !skip || reason != eventReasonContentHash {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestShouldNotSkipWhenSoftDeleted(t *testing.T) {
	etag := "etag-1"
	deleted := time.Now().UTC()
	existing := entity.CalendarEvent{
		Title: "A", Status: constant.CalendarEventStatusCancelled,
		ProviderETag: &etag, DeletedAt: &deleted,
	}
	mapped := entity.CalendarEvent{
		Title: "A", Status: constant.CalendarEventStatusConfirmed, ProviderETag: &etag,
	}
	skip, reason := shouldSkipEventUpdate(existing, mapped)
	if skip || reason != eventReasonResurrect {
		t.Fatalf("skip=%v reason=%q", skip, reason)
	}
}

func TestProviderIdentityChangedPreservesPayloadGate(t *testing.T) {
	etag := "e1"
	existing := entity.CalendarEvent{ProviderETag: &etag, ProviderPayload: []byte(`{"a":1}`)}
	mappedSame := entity.CalendarEvent{ProviderETag: &etag, ProviderPayload: []byte(`{"a":2}`)}
	if providerIdentityChanged(existing, mappedSame) {
		t.Fatal("expected identity unchanged for same etag")
	}
	etag2 := "e2"
	mappedDiff := entity.CalendarEvent{ProviderETag: &etag2}
	if !providerIdentityChanged(existing, mappedDiff) {
		t.Fatal("expected identity changed for different etag")
	}
}
