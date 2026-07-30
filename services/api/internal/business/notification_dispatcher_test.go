package business

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
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

	d := NewNotificationDispatcher(notifRepo, nil, nil)
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

func TestDispatcherSkipsWebPushWithoutFailing(t *testing.T) {
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

	d := NewNotificationDispatcher(notifRepo, nil, nil)
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
		t.Fatalf("WEB_PUSH should stay PENDING, got %v", status)
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
	d := NewNotificationDispatcher(notifRepo, nil, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if notifRepo.byID[notifID].Status != constant.NotificationStatusPending {
		t.Fatalf("future should stay PENDING, got %s", notifRepo.byID[notifID].Status)
	}
}
