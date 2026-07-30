package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const notificationColumns = `
	id, public_id, user_id, channel, title, body, priority,
	timeline_item_parent_id, occurrence_id, notification_type, scheduled_for,
	status, delivery_channels, channel_delivery_status, payload,
	sent_at, read_at, dismissed_at, created_at, updated_at, deleted_at`

const (
	sqlInsertNotification = `
INSERT INTO notifications (
	id, public_id, user_id, channel, title, body, priority,
	timeline_item_parent_id, occurrence_id, notification_type, scheduled_for,
	status, delivery_channels, channel_delivery_status, payload,
	created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,$15::jsonb,$16,$17
)
RETURNING` + notificationColumns

	sqlSelectNotificationByID = `
SELECT` + notificationColumns + `
FROM notifications
WHERE id = $1 AND deleted_at IS NULL`

	sqlListNotificationsByUser = `
SELECT` + notificationColumns + `
FROM notifications
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status = ANY($2::text[])
ORDER BY COALESCE(scheduled_for, created_at) DESC`

	sqlMarkNotificationRead = `
UPDATE notifications SET
	status = 'READ',
	read_at = $3,
	updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
  AND status IN ('PENDING', 'SENT', 'READ')
RETURNING` + notificationColumns

	sqlMarkNotificationDismissed = `
UPDATE notifications SET
	status = 'DISMISSED',
	dismissed_at = $3,
	updated_at = $3
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
  AND status <> 'DISMISSED'
RETURNING` + notificationColumns

	sqlExistsNotificationOccurrence = `
SELECT EXISTS(
	SELECT 1 FROM notifications
	WHERE occurrence_id = $1
	  AND notification_type = $2
	  AND deleted_at IS NULL
)`

	sqlListDuePendingNotifications = `
SELECT` + notificationColumns + `
FROM notifications
WHERE deleted_at IS NULL
  AND status = 'PENDING'
  AND scheduled_for IS NOT NULL
  AND scheduled_for <= $1
ORDER BY scheduled_for ASC
LIMIT $2`

	sqlUpdateNotificationDelivery = `
UPDATE notifications SET
	status = $2,
	channel_delivery_status = $3::jsonb,
	sent_at = $4,
	updated_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING` + notificationColumns
)

// NotificationRepository persists timeline-derived notifications.
type NotificationRepository interface {
	CreateIdempotent(ctx context.Context, n entity.Notification) (created bool, row entity.Notification, err error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, statuses []string) ([]entity.Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error)
	MarkDismissed(ctx context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error)
	ExistsByOccurrence(ctx context.Context, occurrenceID, notificationType string) (bool, error)
	ListDuePending(ctx context.Context, asOf time.Time, limit int) ([]entity.Notification, error)
	UpdateDelivery(ctx context.Context, id uuid.UUID, status string, channelStatus []byte, sentAt *time.Time, updatedAt time.Time) (entity.Notification, error)
	WithTx(tx pgx.Tx) NotificationRepository
}

type notificationRepository struct {
	q Querier
}

// NewNotificationRepository constructs a NotificationRepository.
func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{q: pool}
}

func (r *notificationRepository) WithTx(tx pgx.Tx) NotificationRepository {
	return &notificationRepository{q: tx}
}

func (r *notificationRepository) CreateIdempotent(ctx context.Context, n entity.Notification) (bool, entity.Notification, error) {
	if n.OccurrenceID != nil && n.NotificationType != nil {
		exists, err := r.ExistsByOccurrence(ctx, *n.OccurrenceID, *n.NotificationType)
		if err != nil {
			return false, entity.Notification{}, err
		}
		if exists {
			return false, entity.Notification{}, nil
		}
	}

	channels := n.DeliveryChannels
	if channels == nil {
		channels = []string{}
	}
	payload := n.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	channelStatus := n.ChannelDeliveryStatus
	if len(channelStatus) == 0 {
		channelStatus = json.RawMessage(`{}`)
	}
	channel := n.Channel
	if channel == "" {
		channel = "browser_push"
	}
	priority := n.Priority
	if priority == "" {
		priority = "normal"
	}

	row := r.q.QueryRow(ctx, sqlInsertNotification,
		n.ID, n.PublicID, n.UserID, channel, n.Title, n.Body, priority,
		n.TimelineItemParentID, n.OccurrenceID, n.NotificationType, n.ScheduledFor,
		n.Status, channels, channelStatus, payload,
		n.CreatedAt, n.UpdatedAt,
	)
	created, err := scanNotification(row)
	if err != nil {
		if isUniqueViolation(err) {
			return false, entity.Notification{}, nil
		}
		return false, entity.Notification{}, err
	}
	return true, created, nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Notification, error) {
	return scanNotification(r.q.QueryRow(ctx, sqlSelectNotificationByID, id))
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, statuses []string) ([]entity.Notification, error) {
	if len(statuses) == 0 {
		statuses = []string{"PENDING", "SENT", "READ", "DISMISSED", "FAILED"}
	}
	rows, err := r.q.Query(ctx, sqlListNotificationsByUser, userID, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *notificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error) {
	return scanNotification(r.q.QueryRow(ctx, sqlMarkNotificationRead, id, userID, at))
}

func (r *notificationRepository) MarkDismissed(ctx context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error) {
	return scanNotification(r.q.QueryRow(ctx, sqlMarkNotificationDismissed, id, userID, at))
}

func (r *notificationRepository) ExistsByOccurrence(ctx context.Context, occurrenceID, notificationType string) (bool, error) {
	var exists bool
	err := r.q.QueryRow(ctx, sqlExistsNotificationOccurrence, occurrenceID, notificationType).Scan(&exists)
	return exists, err
}

func (r *notificationRepository) ListDuePending(ctx context.Context, asOf time.Time, limit int) ([]entity.Notification, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.q.Query(ctx, sqlListDuePendingNotifications, asOf, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *notificationRepository) UpdateDelivery(
	ctx context.Context,
	id uuid.UUID,
	status string,
	channelStatus []byte,
	sentAt *time.Time,
	updatedAt time.Time,
) (entity.Notification, error) {
	return scanNotification(r.q.QueryRow(ctx, sqlUpdateNotificationDelivery, id, status, channelStatus, sentAt, updatedAt))
}

func scanNotification(row pgx.Row) (entity.Notification, error) {
	var n entity.Notification
	err := row.Scan(
		&n.ID, &n.PublicID, &n.UserID, &n.Channel, &n.Title, &n.Body, &n.Priority,
		&n.TimelineItemParentID, &n.OccurrenceID, &n.NotificationType, &n.ScheduledFor,
		&n.Status, &n.DeliveryChannels, &n.ChannelDeliveryStatus, &n.Payload,
		&n.SentAt, &n.ReadAt, &n.DismissedAt, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Notification{}, apperr.ErrNotFound
		}
		return entity.Notification{}, fmt.Errorf("scan notification: %w", err)
	}
	return n, nil
}
