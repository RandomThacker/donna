package business

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
)

// NotificationDispatcher promotes due PENDING notifications to SENT for
// Notification Center, Chat, and (when configured) Web Push.
type NotificationDispatcher struct {
	notifications repository.NotificationRepository
	chat          *ConversationService
	pushSubs      *PushSubscriptionService
	pushSender    webpush.Sender
	log           *logger.Logger
	interval      time.Duration
	now           func() time.Time
	batchLimit    int
}

// NewNotificationDispatcher constructs a minute-tick publisher.
// pushSubs / pushSender may be nil when Web Push is not configured.
func NewNotificationDispatcher(
	notifications repository.NotificationRepository,
	chat *ConversationService,
	pushSubs *PushSubscriptionService,
	pushSender webpush.Sender,
	log *logger.Logger,
) *NotificationDispatcher {
	return &NotificationDispatcher{
		notifications: notifications,
		chat:          chat,
		pushSubs:      pushSubs,
		pushSender:    pushSender,
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

	// Backfill chat bubbles for already-SENT notifications that never got a message
	// (status-only CHAT delivery from before chat posting existed).
	backfilled := d.backfillChatNotices(ctx, now)

	if d.log != nil && (sent > 0 || failed > 0 || backfilled > 0) {
		d.log.Info(ctx, "notification dispatcher tick",
			"due", len(due),
			"sent", sent,
			"failed", failed,
			"chat_backfill", backfilled,
		)
	}
}

func (d *NotificationDispatcher) publishOne(ctx context.Context, n entity.Notification, now time.Time) error {
	if wantsChatChannel(n) {
		if _, err := d.postChatNotice(ctx, n); err != nil {
			return err
		}
	}

	channelStatus := parseChannelStatus(n.ChannelDeliveryStatus)
	marked := false
	for _, ch := range n.DeliveryChannels {
		switch ch {
		case constant.DeliveryChannelWebPush:
			if d.webPushConfigured() {
				if d.deliverWebPush(ctx, n) {
					channelStatus[ch] = constant.ChannelDeliverySent
				} else {
					channelStatus[ch] = constant.ChannelDeliveryFailed
				}
				marked = true
			}
			// Not configured: leave WEB_PUSH as-is (usually PENDING).
		default:
			channelStatus[ch] = constant.ChannelDeliverySent
			marked = true
		}
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

func (d *NotificationDispatcher) webPushConfigured() bool {
	return d.pushSender != nil && d.pushSender.Configured() && d.pushSubs != nil
}

// deliverWebPush fans out to all live device subscriptions.
// Returns true when delivery is considered successful (any device ok, or no devices).
func (d *NotificationDispatcher) deliverWebPush(ctx context.Context, n entity.Notification) bool {
	subs, err := d.pushSubs.List(ctx, n.UserID)
	if err != nil {
		if d.log != nil {
			d.log.Warn(ctx, "web push list subscriptions failed",
				"notification_id", n.ID.String(),
				constant.LogAttrError, err,
			)
		}
		return false
	}
	if len(subs) == 0 {
		return true
	}

	payload := webpush.PayloadFromNotification(n)
	var delivered int
	for _, sub := range subs {
		result, err := d.pushSender.Send(ctx, sub, payload)
		if err != nil {
			if d.log != nil {
				d.log.Warn(ctx, "web push send failed",
					"notification_id", n.ID.String(),
					"endpoint", sub.Endpoint,
					"status_code", result.StatusCode,
					constant.LogAttrError, err,
				)
			}
			if result.Gone {
				_ = d.pushSubs.Unsubscribe(ctx, n.UserID, sub.Endpoint)
			}
			continue
		}
		delivered++
	}
	return delivered > 0
}

func (d *NotificationDispatcher) backfillChatNotices(ctx context.Context, now time.Time) int {
	if d.chat == nil || d.notifications == nil {
		return 0
	}
	since := now.Add(-7 * 24 * time.Hour)
	rows, err := d.notifications.ListRecentChatDelivered(ctx, since, d.batchLimit)
	if err != nil {
		if d.log != nil {
			d.log.Warn(ctx, "notification chat backfill list failed", constant.LogAttrError, err)
		}
		return 0
	}
	var n int
	for _, row := range rows {
		created, err := d.postChatNotice(ctx, row)
		if err != nil {
			if d.log != nil {
				d.log.Warn(ctx, "notification chat backfill failed",
					"notification_id", row.ID.String(),
					constant.LogAttrError, err,
				)
			}
			continue
		}
		if created {
			n++
		}
	}
	return n
}

func (d *NotificationDispatcher) postChatNotice(ctx context.Context, n entity.Notification) (bool, error) {
	if d.chat == nil {
		return false, nil
	}
	text := notificationChatText(n)
	if text == "" {
		return false, nil
	}
	clientID := notificationChatClientMessageID(n)
	_, created, err := d.chat.PostAssistantNotice(ctx, n.UserID, text, clientID)
	return created, err
}

func wantsChatChannel(n entity.Notification) bool {
	for _, ch := range n.DeliveryChannels {
		if ch == constant.DeliveryChannelChat {
			return true
		}
	}
	status := parseChannelStatus(n.ChannelDeliveryStatus)
	if _, ok := status[constant.DeliveryChannelChat]; ok {
		return true
	}
	// Default when channels empty: in-app chat.
	return len(n.DeliveryChannels) == 0
}

func notificationChatText(n entity.Notification) string {
	title := strings.TrimSpace(n.Title)
	body := strings.TrimSpace(n.Body)
	var text string
	switch {
	case title != "" && body != "" && !strings.EqualFold(title, body):
		text = title + "\n" + body
	case title != "":
		text = title
	default:
		text = body
	}
	if n.OccurrenceID != nil {
		occ := strings.TrimSpace(*n.OccurrenceID)
		if occ != "" {
			href := constant.NotificationCalendarEventPath + url.QueryEscape(occ)
			if text != "" {
				text += "\n\n"
			}
			text += "[View Event](" + href + ")"
		}
	}
	return text
}

func notificationChatClientMessageID(n entity.Notification) string {
	if n.PublicID != "" {
		return "notif:" + n.PublicID
	}
	return "notif:" + n.ID.String()
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
