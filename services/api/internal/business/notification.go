package business

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// NotificationService manages timeline-derived notification records (no delivery).
type NotificationService struct {
	notifications repository.NotificationRepository
	timeline      *TimelineService
	policies      *NotificationPolicyResolver
	now           func() time.Time
}

// NewNotificationService constructs a NotificationService.
func NewNotificationService(
	notifications repository.NotificationRepository,
	timeline *TimelineService,
	policies *NotificationPolicyResolver,
) *NotificationService {
	if policies == nil {
		policies = NewNotificationPolicyResolver()
	}
	return &NotificationService{
		notifications: notifications,
		timeline:      timeline,
		policies:      policies,
		now:           time.Now,
	}
}

// List returns notifications for a user filtered by status (empty = all lifecycle states).
func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, statuses []string) ([]entity.Notification, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.notifications.ListByUser(ctx, userID, statuses)
}

// MarkRead marks a notification as READ.
func (s *NotificationService) MarkRead(ctx context.Context, userID, id uuid.UUID) (entity.Notification, error) {
	if userID == uuid.Nil || id == uuid.Nil {
		return entity.Notification{}, fmt.Errorf("%w: user and notification id are required", apperr.ErrValidation)
	}
	existing, err := s.notifications.GetByID(ctx, id)
	if err != nil {
		return entity.Notification{}, err
	}
	if existing.UserID != userID {
		return entity.Notification{}, apperr.ErrForbidden
	}
	return s.notifications.MarkRead(ctx, id, userID, s.now().UTC())
}

// MarkDismissed marks a notification as DISMISSED.
func (s *NotificationService) MarkDismissed(ctx context.Context, userID, id uuid.UUID) (entity.Notification, error) {
	if userID == uuid.Nil || id == uuid.Nil {
		return entity.Notification{}, fmt.Errorf("%w: user and notification id are required", apperr.ErrValidation)
	}
	existing, err := s.notifications.GetByID(ctx, id)
	if err != nil {
		return entity.Notification{}, err
	}
	if existing.UserID != userID {
		return entity.Notification{}, apperr.ErrForbidden
	}
	return s.notifications.MarkDismissed(ctx, id, userID, s.now().UTC())
}

// EnqueueForUser inspects the user's near-term timeline and creates PENDING notifications.
// Idempotent on (occurrence_id, notification_type).
func (s *NotificationService) EnqueueForUser(ctx context.Context, userID uuid.UUID) (created int, err error) {
	if userID == uuid.Nil {
		return 0, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if s.timeline == nil {
		return 0, fmt.Errorf("%w: timeline service is not configured", apperr.ErrInvalid)
	}

	now := s.now().UTC()
	windowEnd := now.Add(constant.NotificationLookaheadWindow)
	timelineTo := windowEnd.Add(constant.NotificationMaxPolicyLead)

	items, err := s.timeline.List(ctx, userID, now, timelineTo)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if item.Status == constant.TimelineStatusCancelled {
			continue
		}
		policy := s.policies.Resolve(item)
		if policy == nil {
			continue
		}
		reminderAt := policy.ReminderTime(item).UTC()
		if reminderAt.Before(now) || !reminderAt.Before(windowEnd) {
			// ReminderTime must fall in [now, now+20m).
			continue
		}
		ok, err := s.enqueueItem(ctx, userID, item, reminderAt)
		if err != nil {
			return created, err
		}
		if ok {
			created++
		}
	}
	return created, nil
}

func (s *NotificationService) enqueueItem(
	ctx context.Context,
	userID uuid.UUID,
	item entity.TimelineItem,
	scheduledFor time.Time,
) (bool, error) {
	occID := item.OccurrenceID
	if occID == "" {
		occID = item.ID
	}
	notifType := notificationTypeFromTimeline(item.Type)
	if notifType == "" {
		return false, nil
	}

	parentID := occID
	if item.ParentID != nil && *item.ParentID != "" {
		parentID = *item.ParentID
	}

	payload, err := buildNotificationPayload(item, occID)
	if err != nil {
		return false, err
	}
	channelStatus, err := json.Marshal(map[string]string{
		constant.DeliveryChannelWebPush: constant.ChannelDeliveryPending,
		constant.DeliveryChannelChat:    constant.ChannelDeliveryPending,
	})
	if err != nil {
		return false, err
	}

	id, err := idgen.NewUUIDv7()
	if err != nil {
		return false, err
	}
	now := s.now().UTC()
	title := item.Title
	body := notificationBody(item, scheduledFor)

	created, _, err := s.notifications.CreateIdempotent(ctx, entity.Notification{
		ID:                    id,
		PublicID:              idgen.PublicID(constant.PublicIDPrefixNotification, id),
		UserID:                userID,
		TimelineItemParentID:  &parentID,
		OccurrenceID:          &occID,
		Title:                 title,
		Body:                  body,
		NotificationType:      &notifType,
		ScheduledFor:          &scheduledFor,
		Status:                constant.NotificationStatusPending,
		DeliveryChannels:      append([]string(nil), constant.DefaultDeliveryChannels...),
		ChannelDeliveryStatus: channelStatus,
		Payload:               payload,
		Channel:               "browser_push",
		Priority:              "normal",
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	return created, err
}

func notificationTypeFromTimeline(timelineType string) string {
	switch timelineType {
	case constant.TimelineTypeEvent:
		return constant.NotificationTypeEvent
	case constant.TimelineTypeReminder:
		return constant.NotificationTypeReminder
	default:
		return ""
	}
}

func notificationBody(item entity.TimelineItem, scheduledFor time.Time) string {
	switch item.Type {
	case constant.TimelineTypeReminder:
		return "Reminder is due"
	default:
		mins := int(item.StartAt.Sub(scheduledFor).Minutes())
		if mins < 1 {
			mins = 1
		}
		return fmt.Sprintf("Starts in %d minutes", mins)
	}
}

func buildNotificationPayload(item entity.TimelineItem, occurrenceID string) (json.RawMessage, error) {
	payload := map[string]any{
		"timelineItemId": item.ID,
		"occurrenceId":   occurrenceID,
		"source":         item.Source,
		"type":           item.Type,
		"startAt":        item.StartAt.UTC().Format(time.RFC3339Nano),
		"endAt":          item.EndAt.UTC().Format(time.RFC3339Nano),
		"timezone":       item.Timezone,
		"deepLink":       constant.NotificationDeepLinkPath + occurrenceID,
	}
	if item.ParentID != nil {
		payload["parentId"] = *item.ParentID
	}
	if item.RecurrenceRule != nil {
		payload["recurrenceRule"] = *item.RecurrenceRule
	}
	return json.Marshal(payload)
}
