package business

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
	"github.com/google/uuid"
)

func TestDispatcherPublishesDueToChat(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000203")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000204")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)
	occ := "occ-success"
	channelStatus, _ := json.Marshal(map[string]string{
		constant.DeliveryChannelChat: constant.ChannelDeliveryPending,
	})

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID:                    notifID,
		UserID:                userID,
		Title:                 "Guitar",
		Body:                  "Starts in 15 minutes",
		Status:                constant.NotificationStatusPending,
		ScheduledFor:          &scheduled,
		OccurrenceID:          &occ,
		DeliveryChannels:      []string{constant.DeliveryChannelChat},
		ChannelDeliveryStatus: channelStatus,
	}

	d := NewNotificationDispatcher(notifRepo, nil, nil, nil, nil, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	n := notifRepo.byID[notifID]
	if n.Status != constant.NotificationStatusSent {
		t.Fatalf("status = %s", n.Status)
	}
	if n.SentAt == nil {
		t.Fatal("expected sent_at")
	}
	var status map[string]string
	_ = json.Unmarshal(n.ChannelDeliveryStatus, &status)
	if status[constant.DeliveryChannelChat] != constant.ChannelDeliverySent {
		t.Fatalf("channel status = %v", status)
	}
}

func TestDispatcherLeavesWebPushPendingWhenNotConfigured(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000206")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000207")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "t", Body: "b",
		Status: constant.NotificationStatusPending, ScheduledFor: &scheduled,
		DeliveryChannels: []string{constant.DeliveryChannelWebPush, constant.DeliveryChannelChat},
		ChannelDeliveryStatus: json.RawMessage(`{"WEB_PUSH":"PENDING","CHAT":"PENDING"}`),
	}

	d := NewNotificationDispatcher(notifRepo, nil, nil, nil, nil, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	n := notifRepo.byID[notifID]
	if n.Status != constant.NotificationStatusSent {
		t.Fatalf("status = %s", n.Status)
	}
	var status map[string]string
	_ = json.Unmarshal(n.ChannelDeliveryStatus, &status)
	if status[constant.DeliveryChannelChat] != constant.ChannelDeliverySent {
		t.Fatalf("CHAT should be SENT, got %v", status)
	}
	if status[constant.DeliveryChannelWebPush] != constant.ChannelDeliveryPending {
		t.Fatalf("WEB_PUSH should stay PENDING when push not configured, got %v", status)
	}
}

type stubPushSender struct {
	calls int
}

func (s *stubPushSender) Configured() bool { return true }

func (s *stubPushSender) Send(_ context.Context, _ entity.PushSubscription, _ webpush.Payload) (webpush.Result, error) {
	s.calls++
	return webpush.Result{StatusCode: 201}, nil
}

func TestDispatcherSendsWebPushWhenConfigured(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000226")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000227")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "Stretch", Body: "Reminder is due",
		Status: constant.NotificationStatusPending, ScheduledFor: &scheduled,
		DeliveryChannels: []string{constant.DeliveryChannelWebPush, constant.DeliveryChannelChat},
		ChannelDeliveryStatus: json.RawMessage(`{"WEB_PUSH":"PENDING","CHAT":"PENDING"}`),
	}

	pushRepo := newMemPushSubRepo()
	_, err := pushRepo.Upsert(context.Background(), entity.PushSubscription{
		ID:       uuid.MustParse("018f0000-0000-7000-8000-000000000228"),
		PublicID: "psub_test",
		UserID:   userID,
		Endpoint: "https://push.example/sub",
		P256dh:   "p256",
		Auth:     "auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	pushSvc := NewPushSubscriptionService(pushRepo)
	sender := &stubPushSender{}

	d := NewNotificationDispatcher(notifRepo, nil, pushSvc, sender, nil, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if sender.calls != 1 {
		t.Fatalf("expected 1 push send, got %d", sender.calls)
	}
	n := notifRepo.byID[notifID]
	var status map[string]string
	_ = json.Unmarshal(n.ChannelDeliveryStatus, &status)
	if status[constant.DeliveryChannelWebPush] != constant.ChannelDeliverySent {
		t.Fatalf("WEB_PUSH should be SENT, got %v", status)
	}
}

func TestDispatcherIgnoresFuturePending(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000215")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000216")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	future := now.Add(10 * time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "t", Body: "b",
		Status: constant.NotificationStatusPending, ScheduledFor: &future,
		DeliveryChannels: []string{constant.DeliveryChannelChat},
	}
	d := NewNotificationDispatcher(notifRepo, nil, nil, nil, nil, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if notifRepo.byID[notifID].Status != constant.NotificationStatusPending {
		t.Fatalf("future should stay PENDING, got %s", notifRepo.byID[notifID].Status)
	}
}
