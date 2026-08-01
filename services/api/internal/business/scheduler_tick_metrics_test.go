package business

import (
	"strings"
	"testing"
	"time"
)

func TestSchedulerTickMetricsSummary(t *testing.T) {
	t.Parallel()
	m := SchedulerTickMetrics{
		FeedSource:          FeedSourceTimeline,
		WindowFrom:          time.Date(2026, 8, 1, 12, 40, 0, 0, time.UTC),
		WindowTo:            time.Date(2026, 8, 1, 13, 0, 0, 0, time.UTC),
		UsersScanned:        2,
		ProvidersQueried:    4,
		OccurrencesReturned: 86,
		AfterExpansion:      112,
		AfterDedup:          108,
		Notifications:       3,
		DatabaseQueries:     11,
		AllocBytes:          4096,
		Duration:            48 * time.Millisecond,
	}
	got := m.Summary()
	for _, want := range []string{
		"Scheduler Tick",
		"Feed: timeline",
		"Window: 12:40 → 13:00",
		"Providers: 4",
		"Occurrences: 86",
		"After Expansion: 112",
		"After Dedup: 108",
		"Notifications: 3",
		"DB Queries: 11",
		"Duration: 48 ms",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary missing %q\n%s", want, got)
		}
	}
}

func TestSchedulerTickMetricsMergeUser(t *testing.T) {
	t.Parallel()
	m := SchedulerTickMetrics{ProvidersQueried: 4}
	m.MergeUser(FeedBuildStats{
		FeedSource:          FeedSourceTimeline,
		ProvidersQueried:    4,
		OccurrencesReturned: 10,
		AfterExpansion:      12,
		AfterDedup:          11,
		DatabaseQueries:     5,
		Notifications:       2,
	})
	m.MergeUser(FeedBuildStats{
		ProvidersQueried:    4,
		OccurrencesReturned: 3,
		AfterExpansion:      3,
		AfterDedup:          3,
		DatabaseQueries:     4,
		Notifications:       1,
	})
	if m.ProvidersQueried != 4 {
		t.Fatalf("providers = %d", m.ProvidersQueried)
	}
	if m.OccurrencesReturned != 13 || m.AfterExpansion != 15 || m.AfterDedup != 14 {
		t.Fatalf("pipeline totals = %+v", m)
	}
	if m.Notifications != 3 || m.DatabaseQueries != 9 {
		t.Fatalf("notif/db = %+v", m)
	}
}
