package business

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// TimelineProvider fetches TimelineItems from one source for a time window.
// New sources plug in here — TimelineService never switches on Source.
type TimelineProvider interface {
	Fetch(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error)
}

// NotificationPolicy decides when a timeline item should notify.
// Implementations live in notification_policy.go; delivery is a later phase.
type NotificationPolicy interface {
	ReminderTime(item entity.TimelineItem) time.Time
}
