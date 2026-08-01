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
