package business

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/google/uuid"
)

// ActiveUserIDLister lists live user ids for background scans.
type ActiveUserIDLister interface {
	ListActiveIDs(ctx context.Context) ([]uuid.UUID, error)
}

// NotificationScheduler periodically enqueues PENDING notifications from the timeline.
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

// Tick scans all active users once (exported for tests).
func (s *NotificationScheduler) Tick(ctx context.Context) {
	ids, err := s.users.ListActiveIDs(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "notification scheduler list users failed", constant.LogAttrError, err)
		}
		return
	}
	var total int
	for _, userID := range ids {
		n, err := s.notifications.EnqueueForUser(ctx, userID)
		if err != nil {
			if s.log != nil {
				s.log.Warn(ctx, "notification enqueue failed",
					"user_id", userID.String(),
					constant.LogAttrError, err,
				)
			}
			continue
		}
		total += n
	}
	if s.log != nil && total > 0 {
		s.log.Info(ctx, "notification scheduler enqueued", "created", total, "users", len(ids))
	}
}
