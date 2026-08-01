package business

import (
	"context"
	"runtime"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/google/uuid"
)

// ActiveUserIDLister lists live user ids for background scans.
type ActiveUserIDLister interface {
	ListActiveIDs(ctx context.Context) ([]uuid.UUID, error)
}

// NotificationScheduler periodically enqueues PENDING notifications from occurrences.
type NotificationScheduler struct {
	notifications *NotificationService
	users         ActiveUserIDLister
	log           *logger.Logger
	interval      time.Duration
	now           func() time.Time
}

// NewNotificationScheduler constructs a minute ticker scheduler.
func NewNotificationScheduler(
	notifications *NotificationService,
	users ActiveUserIDLister,
	log *logger.Logger,
) *NotificationScheduler {
	return &NotificationScheduler{
		notifications: notifications,
		users:         users,
		log:           log,
		interval:      constant.NotificationSchedulerInterval,
		now:           time.Now,
	}
}

// Run blocks until ctx is canceled, ticking every minute.
func (s *NotificationScheduler) Run(ctx context.Context) {
	if s.notifications == nil || s.users == nil {
		return
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.Tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Tick(ctx)
		}
	}
}

// Tick scans all active users once (exported for tests) and logs tick metrics.
func (s *NotificationScheduler) Tick(ctx context.Context) {
	wallStart := time.Now()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	now := s.now().UTC()
	metrics := SchedulerTickMetrics{
		FeedSource: FeedSourceOccurrence,
		WindowFrom: now,
		WindowTo:   now.Add(constant.NotificationLookaheadWindow),
	}

	ids, err := s.users.ListActiveIDs(ctx)
	metrics.DatabaseQueries++ // ListActiveIDs
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "notification scheduler list users failed", constant.LogAttrError, err)
		}
		metrics.Duration = time.Since(wallStart)
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		if memAfter.TotalAlloc >= memBefore.TotalAlloc {
			metrics.AllocBytes = memAfter.TotalAlloc - memBefore.TotalAlloc
		}
		metrics.Log(ctx, s.log)
		return
	}
	metrics.UsersScanned = len(ids)
	if s.notifications != nil {
		metrics.ProvidersQueried = occurrence.ProviderCount(s.notifications.occurrences)
	}

	for _, userID := range ids {
		_, userStats, err := s.notifications.enqueueForUser(ctx, userID)
		if err != nil {
			if s.log != nil {
				s.log.Warn(ctx, "notification enqueue failed",
					"user_id", userID.String(),
					constant.LogAttrError, err,
				)
			}
			continue
		}
		metrics.MergeUser(userStats)
	}

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	if memAfter.TotalAlloc >= memBefore.TotalAlloc {
		metrics.AllocBytes = memAfter.TotalAlloc - memBefore.TotalAlloc
	}
	metrics.Duration = time.Since(wallStart)
	metrics.Log(ctx, s.log)
}
