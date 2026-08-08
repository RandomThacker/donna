package repository

import (
	"strings"
	"testing"
)

func TestSchedulerSQLOmitsHeavyColumns(t *testing.T) {
	t.Parallel()
	if strings.Contains(sqlSelectCalendarEventsForScheduler, "provider_payload") {
		t.Fatal("scheduler calendar SQL must not select provider_payload")
	}
	if strings.Contains(sqlSelectCalendarEventsForScheduler, "attendees_summary") {
		t.Fatal("scheduler calendar SQL must not select attendees_summary")
	}
	if strings.Contains(sqlSelectCalendarEventsForScheduler, "organizer_summary") {
		t.Fatal("scheduler calendar SQL must not select organizer_summary")
	}
	if strings.Contains(sqlSelectCalendarEventsForScheduler, "s.color") {
		t.Fatal("scheduler calendar SQL must not select source color")
	}
	// Wide Timeline query still has payload (unchanged).
	if !strings.Contains(sqlSelectCalendarEventsByUserRangeWithProvider, "provider_payload") {
		t.Fatal("timeline calendar SQL should still select provider_payload")
	}
	if strings.Contains(sqlListDonnaEventsForScheduler, "color") {
		t.Fatal("donna event scheduler SQL must not select color")
	}
	if strings.Contains(sqlListDonnaRemindersForScheduler, "color") {
		t.Fatal("donna reminder scheduler SQL must not select color")
	}
}

func TestSyncDecisionSQLOmitsProviderPayload(t *testing.T) {
	t.Parallel()
	if strings.Contains(sqlSelectCalendarEventSyncDecisionBySourceProvider, "provider_payload") {
		t.Fatal("sync decision SQL must not select provider_payload")
	}
	if strings.Contains(sqlSelectCalendarEventSyncDecisionByProviderEventIDs, "provider_payload") {
		t.Fatal("batch sync decision SQL must not select provider_payload")
	}
	if !strings.Contains(sqlSelectCalendarEventBySourceProvider, "provider_payload") {
		t.Fatal("full-row get SQL should still select provider_payload")
	}
	if !strings.Contains(sqlSelectCalendarEventSyncDecisionByProviderEventIDs, "DISTINCT ON (provider_event_id)") {
		t.Fatal("batch sync decision SQL must DISTINCT ON provider_event_id")
	}
	if !strings.Contains(sqlSelectCalendarEventSyncDecisionByProviderEventIDs, "deleted_at NULLS FIRST") {
		t.Fatal("batch sync decision SQL must prefer live rows")
	}
	required := []string{
		"provider_etag",
		"provider_updated_at",
		"deleted_at",
		"title",
		"starts_at",
		"ends_at",
		"organizer_summary",
		"attendees_summary",
		"recurrence_rule",
	}
	for _, col := range required {
		if !strings.Contains(sqlSelectCalendarEventSyncDecisionByProviderEventIDs, col) {
			t.Fatalf("batch sync decision SQL missing %s", col)
		}
	}
}
