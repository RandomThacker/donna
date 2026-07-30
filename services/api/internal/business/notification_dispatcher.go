package business

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
)

// NotificationDispatcher promotes due PENDING notifications to SENT for
// in-app surfaces (Notification Center + Chat). Web Push is disabled.
type NotificationDispatcher struct {
	notifications repository.NotificationRepository
	log           *logger.Logger
	interval      time.Duration
	now           func() time.Time
	batchLimit    int
}

// NewNotificationDispatcher constructs a minute-tick in-app publisher.
func NewNotificationDispatcher(
	notifications repository.NotificationRepository,
	log *logger.Logger,
) *NotificationDispatcher {
	return &NotificationDispatcher{
		notifications: notifications,
		log:           log,
		interval:      constant.NotificationDispatcherInterval,
		now:           time.Now,
		batchLimit:    constant.NotificationDispatcherBatchLimit,
	}
}

// Run blocks until ctx is canceled, ticking every minute.
func (d *NotificationDispatcher) Run(ctx context.Context) {
	if d.notifications == nil {
		return
	}
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	d.Tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.Tick(ctx)
		}
	}
}

// Tick publishes all due PENDING notifications once (exported for tests).
func (d *NotificationDispatcher) Tick(ctx context.Context) {
	now := d.now().UTC()
	due, err := d.notifications.ListDuePending(ctx, now, d.batchLimit)
	if err != nil {
		if d.log != nil {
			d.log.Error(ctx, "notification dispatcher list due failed", constant.LogAttrError, err)
		}
		return
	}
	var sent, failed int
	for _, n := range due {
		if err := d.publishOne(ctx, n, now); err != nil {
			if d.log != nil {
				d.log.Warn(ctx, "notification publish failed",
					"notification_id", n.ID.String(),
					constant.LogAttrError, err,
				)
			}
			failed++
			continue
		}
		sent++
	}
	if d.log != nil && (sent > 0 || failed > 0) {
		d.log.Info(ctx, "notification dispatcher tick",
			"due", len(due),
			"sent", sent,
			"failed", failed,
		)
	}
}

func (d *NotificationDispatcher) publishOne(ctx context.Context, n entity.Notification, now time.Time) error {
	channelStatus := parseChannelStatus(n.ChannelDeliveryStatus)
	marked := false
	for _, ch := range n.DeliveryChannels {
		if ch == constant.DeliveryChannelWebPush {
			// Web Push is intentionally disabled — skip without failing the row.
			continue
		}
		channelStatus[ch] = constant.ChannelDeliverySent
		marked = true
	}
	if !marked {
		channelStatus[constant.DeliveryChannelChat] = constant.ChannelDeliverySent
	}
	statusJSON, err := json.Marshal(channelStatus)
	if err != nil {
		return err
	}
	sentAt := now
	_, err = d.notifications.UpdateDelivery(
		ctx,
		n.ID,
		constant.NotificationStatusSent,
		statusJSON,
		&sentAt,
		now,
	)
	return err
}

func parseChannelStatus(raw []byte) map[string]string {
	out := map[string]string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = map[string]string{}
	}
	return out
}
