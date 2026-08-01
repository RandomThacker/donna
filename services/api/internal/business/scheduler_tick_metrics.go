package business

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/logger"
)

// FeedSourceTimeline marks metrics gathered via TimelineService (pre–Task 4).
const FeedSourceTimeline = "timeline"

// FeedSourceOccurrence marks metrics gathered via OccurrenceService (post–Task 4).
const FeedSourceOccurrence = "occurrence"

// FeedBuildStats captures one user's scheduling-feed build for a scheduler tick.
// Field names mirror the Occurrence pipeline so before/after Task 4 is comparable.
type FeedBuildStats struct {
	FeedSource string

	WindowFrom time.Time
	WindowTo   time.Time

	ProvidersQueried    int
	OccurrencesReturned int // after provider collect (providers may already expand)
	AfterExpansion      int // after service-level expansion stage
	AfterDedup          int // after service-level dedup stage
	DatabaseQueries     int
	Notifications       int
}

// SchedulerTickMetrics aggregates one NotificationScheduler.Tick across all users.
type SchedulerTickMetrics struct {
	FeedSource string

	WindowFrom time.Time
	WindowTo   time.Time

	UsersScanned int

	ProvidersQueried    int
	OccurrencesReturned int
	AfterExpansion      int
	AfterDedup          int
	Notifications       int
	DatabaseQueries     int

	AllocBytes uint64
	Duration   time.Duration
}

// MergeUser folds one user's enqueue stats into the tick totals.
// ProvidersQueried keeps the configured provider count (max), not users×providers.
func (m *SchedulerTickMetrics) MergeUser(u FeedBuildStats) {
	if m == nil {
		return
	}
	if m.FeedSource == "" && u.FeedSource != "" {
		m.FeedSource = u.FeedSource
	}
	if m.WindowFrom.IsZero() && !u.WindowFrom.IsZero() {
		m.WindowFrom = u.WindowFrom
		m.WindowTo = u.WindowTo
	}
	if u.ProvidersQueried > m.ProvidersQueried {
		m.ProvidersQueried = u.ProvidersQueried
	}
	m.OccurrencesReturned += u.OccurrencesReturned
	m.AfterExpansion += u.AfterExpansion
	m.AfterDedup += u.AfterDedup
	m.Notifications += u.Notifications
	m.DatabaseQueries += u.DatabaseQueries
}

// Summary formats a human-readable tick report for log scanners / before-after diffs.
func (m SchedulerTickMetrics) Summary() string {
	feed := m.FeedSource
	if feed == "" {
		feed = "unknown"
	}
	return fmt.Sprintf(
		"Scheduler Tick\n"+
			"Feed: %s\n"+
			"Window: %s → %s\n"+
			"Users: %d\n"+
			"Providers: %d\n"+
			"Occurrences: %d\n"+
			"After Expansion: %d\n"+
			"After Dedup: %d\n"+
			"Notifications: %d\n"+
			"DB Queries: %d\n"+
			"Alloc: %d B\n"+
			"Duration: %d ms",
		feed,
		formatTickTime(m.WindowFrom),
		formatTickTime(m.WindowTo),
		m.UsersScanned,
		m.ProvidersQueried,
		m.OccurrencesReturned,
		m.AfterExpansion,
		m.AfterDedup,
		m.Notifications,
		m.DatabaseQueries,
		m.AllocBytes,
		m.Duration.Milliseconds(),
	)
}

func formatTickTime(t time.Time) string {
	if t.IsZero() {
		return "?"
	}
	return t.UTC().Format("15:04")
}

// Log emits structured attrs plus the human-readable Summary on every tick.
func (m SchedulerTickMetrics) Log(ctx context.Context, log *logger.Logger) {
	if log == nil {
		return
	}
	log.Info(ctx, "scheduler tick",
		"feed_source", m.FeedSource,
		"window_from", m.WindowFrom.UTC().Format(time.RFC3339),
		"window_to", m.WindowTo.UTC().Format(time.RFC3339),
		"users", m.UsersScanned,
		"providers", m.ProvidersQueried,
		"occurrences", m.OccurrencesReturned,
		"after_expansion", m.AfterExpansion,
		"after_dedup", m.AfterDedup,
		"notifications", m.Notifications,
		"db_queries", m.DatabaseQueries,
		"alloc_bytes", m.AllocBytes,
		"duration_ms", m.Duration.Milliseconds(),
		"summary", m.Summary(),
	)
}
