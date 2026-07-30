package business

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
)

// NotificationDispatcher delivers due PENDING notifications via Web Push.
type NotificationDispatcher struct {
	notifications repository.NotificationRepository
	subs          repository.PushSubscriptionRepository
	sender        webpush.Sender
	log           *logger.Logger
	interval      time.Duration
	now           func() time.Time
	batchLimit    int
}

// NewNotificationDispatcher constructs a minute-tick Web Push dispatcher.
func NewNotificationDispatcher(
	notifications repository.NotificationRepository,
	subs repository.PushSubscriptionRepository,
	sender webpush.Sender,
	log *logger.Logger,
) *NotificationDispatcher {
	return &NotificationDispatcher{
		notifications: notifications,
		subs:          subs,
		sender:        sender,
		log:           log,
		interval:      constant.NotificationDispatcherInterval,
		now:           time.Now,
		batchLimit:    constant.NotificationDispatcherBatchLimit,
	}
}

// Run blocks until ctx is canceled, ticking every minute.
func (d *NotificationDispatcher) Run(ctx context.Context) {
	if d.notifications == nil || d.subs == nil {
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

// Tick delivers all due PENDING notifications once (exported for tests).
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
		outcome, err := d.deliverOne(ctx, n, now)
		if err != nil {
			if d.log != nil {
				d.log.Warn(ctx, "notification dispatch failed",
					"notification_id", n.ID.String(),
					constant.LogAttrError, err,
				)
			}
			failed++
			continue
		}
		switch outcome {
		case deliverOutcomeSent:
			sent++
		case deliverOutcomeFailed:
			failed++
		}
	}
	if d.log != nil && (sent > 0 || failed > 0) {
		d.log.Info(ctx, "notification dispatcher tick",
			"due", len(due),
			"sent", sent,
			"failed", failed,
		)
	}
}

func (d *NotificationDispatcher) deliverOne(ctx context.Context, n entity.Notification, now time.Time) (deliverOutcome, error) {
	if !wantsWebPush(n) {
		// Leave PENDING until a supported channel exists; do not mark FAILED.
		return deliverOutcomeSkipped, nil
	}

	channelStatus := parseChannelStatus(n.ChannelDeliveryStatus)
	payload := webpush.PayloadFromNotification(n)

	subs, err := d.subs.ListByUser(ctx, n.UserID)
	if err != nil {
		return deliverOutcomeFailed, err
	}

	success := false
	if d.sender != nil && d.sender.Configured() && len(subs) > 0 {
		for _, sub := range subs {
			res, sendErr := d.sender.Send(ctx, sub, payload)
			if sendErr != nil {
				if res.Gone {
					_ = d.subs.SoftDeleteByEndpoint(ctx, n.UserID, sub.Endpoint, now)
				}
				continue
			}
			success = true
		}
	}

	if success {
		channelStatus[constant.DeliveryChannelWebPush] = constant.ChannelDeliverySent
		statusJSON, err := json.Marshal(channelStatus)
		if err != nil {
			return deliverOutcomeFailed, err
		}
		sentAt := now
		_, err = d.notifications.UpdateDelivery(ctx, n.ID, constant.NotificationStatusSent, statusJSON, &sentAt, now)
		if err != nil {
			return deliverOutcomeFailed, err
		}
		return deliverOutcomeSent, nil
	}

	channelStatus[constant.DeliveryChannelWebPush] = constant.ChannelDeliveryFailed
	statusJSON, err := json.Marshal(channelStatus)
	if err != nil {
		return deliverOutcomeFailed, err
	}
	_, err = d.notifications.UpdateDelivery(ctx, n.ID, constant.NotificationStatusFailed, statusJSON, nil, now)
	if err != nil {
		return deliverOutcomeFailed, err
	}
	return deliverOutcomeFailed, nil
}

type deliverOutcome int

const (
	deliverOutcomeSkipped deliverOutcome = iota
	deliverOutcomeSent
	deliverOutcomeFailed
)

func wantsWebPush(n entity.Notification) bool {
	for _, ch := range n.DeliveryChannels {
		if ch == constant.DeliveryChannelWebPush {
			return true
		}
	}
	return false
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
