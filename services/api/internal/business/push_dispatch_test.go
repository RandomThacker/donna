package business

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type memPushSubRepo struct {
	byID       map[uuid.UUID]entity.PushSubscription
	byEndpoint map[string]uuid.UUID // userID|endpoint -> id
}

func newMemPushSubRepo() *memPushSubRepo {
	return &memPushSubRepo{
		byID:       map[uuid.UUID]entity.PushSubscription{},
		byEndpoint: map[string]uuid.UUID{},
	}
}

func (m *memPushSubRepo) epKey(userID uuid.UUID, endpoint string) string {
	return userID.String() + "|" + endpoint
}

func (m *memPushSubRepo) Upsert(_ context.Context, sub entity.PushSubscription) (entity.PushSubscription, error) {
	k := m.epKey(sub.UserID, sub.Endpoint)
	if id, ok := m.byEndpoint[k]; ok {
		existing := m.byID[id]
		existing.P256dh = sub.P256dh
		existing.Auth = sub.Auth
		existing.UserAgent = sub.UserAgent
		existing.UpdatedAt = sub.UpdatedAt
		m.byID[id] = existing
		return existing, nil
	}
	m.byID[sub.ID] = sub
	m.byEndpoint[k] = sub.ID
	return sub, nil
}

func (m *memPushSubRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]entity.PushSubscription, error) {
	out := make([]entity.PushSubscription, 0)
	for _, s := range m.byID {
		if s.UserID == userID && s.DeletedAt == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *memPushSubRepo) SoftDeleteByEndpoint(_ context.Context, userID uuid.UUID, endpoint string, deletedAt time.Time) error {
	k := m.epKey(userID, endpoint)
	id, ok := m.byEndpoint[k]
	if !ok {
		return apperr.ErrNotFound
	}
	s := m.byID[id]
	if s.DeletedAt != nil {
		return apperr.ErrNotFound
	}
	s.DeletedAt = &deletedAt
	s.UpdatedAt = deletedAt
	m.byID[id] = s
	delete(m.byEndpoint, k)
	return nil
}

func (m *memPushSubRepo) WithTx(pgx.Tx) repository.PushSubscriptionRepository { return m }

var _ repository.PushSubscriptionRepository = (*memPushSubRepo)(nil)

type stubPushSender struct {
	configured bool
	fail       map[string]error // endpoint -> error
	gone       map[string]bool
	sent       []string
}

func (s *stubPushSender) Configured() bool { return s.configured }

func (s *stubPushSender) Send(_ context.Context, sub entity.PushSubscription, _ webpush.Payload) (webpush.Result, error) {
	s.sent = append(s.sent, sub.Endpoint)
	if err, ok := s.fail[sub.Endpoint]; ok {
		return webpush.Result{Gone: s.gone[sub.Endpoint]}, err
	}
	return webpush.Result{StatusCode: 201}, nil
}

func TestPushSubscribeUpsertAndUnsubscribe(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000201")
	repo := newMemPushSubRepo()
	svc := NewPushSubscriptionService(repo)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	ua := "DonnaTest/1.0"
	first, err := svc.Subscribe(context.Background(), userID, SubscribePushInput{
		Endpoint:  "https://push.example/a",
		P256dh:    "p256-1",
		Auth:      "auth-1",
		UserAgent: &ua,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Endpoint != "https://push.example/a" || first.P256dh != "p256-1" {
		t.Fatalf("first = %+v", first)
	}

	second, err := svc.Subscribe(context.Background(), userID, SubscribePushInput{
		Endpoint: "https://push.example/a",
		P256dh:   "p256-2",
		Auth:     "auth-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("upsert should keep id: %s vs %s", second.ID, first.ID)
	}
	if second.P256dh != "p256-2" {
		t.Fatalf("keys not updated: %s", second.P256dh)
	}

	other, err := svc.Subscribe(context.Background(), userID, SubscribePushInput{
		Endpoint: "https://push.example/b",
		P256dh:   "p256-b",
		Auth:     "auth-b",
	})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(listed))
	}

	if err := svc.Unsubscribe(context.Background(), userID, other.Endpoint); err != nil {
		t.Fatal(err)
	}
	listed, err = svc.List(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 after unsubscribe, got %d", len(listed))
	}

	err = svc.Unsubscribe(context.Background(), userID, "https://push.example/missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPushSubscribeValidation(t *testing.T) {
	t.Parallel()
	svc := NewPushSubscriptionService(newMemPushSubRepo())
	_, err := svc.Subscribe(context.Background(), uuid.Nil, SubscribePushInput{
		Endpoint: "https://x", P256dh: "a", Auth: "b",
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("nil user: %v", err)
	}
	_, err = svc.Subscribe(context.Background(), uuid.MustParse("018f0000-0000-7000-8000-000000000202"), SubscribePushInput{})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("empty keys: %v", err)
	}
}

func TestDispatcherSuccessfulSend(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000203")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000204")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)
	occ := "occ-success"
	channelStatus, _ := json.Marshal(map[string]string{
		constant.DeliveryChannelWebPush: constant.ChannelDeliveryPending,
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
		DeliveryChannels:      []string{constant.DeliveryChannelWebPush},
		ChannelDeliveryStatus: channelStatus,
		Payload:               json.RawMessage(`{"source":"DONNA","type":"EVENT","startAt":"2026-07-30T12:15:00Z","occurrenceId":"occ-success"}`),
	}

	subRepo := newMemPushSubRepo()
	_, _ = subRepo.Upsert(context.Background(), entity.PushSubscription{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000205"),
		UserID: userID, Endpoint: "https://push.example/ok", P256dh: "k", Auth: "a",
		CreatedAt: now, UpdatedAt: now,
	})

	sender := &stubPushSender{configured: true}
	d := NewNotificationDispatcher(notifRepo, subRepo, sender, nil)
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
	if status[constant.DeliveryChannelWebPush] != constant.ChannelDeliverySent {
		t.Fatalf("channel status = %v", status)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("sent count = %d", len(sender.sent))
	}
}

func TestDispatcherFailedSend(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000206")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000207")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "t", Body: "b",
		Status: constant.NotificationStatusPending, ScheduledFor: &scheduled,
		DeliveryChannels: []string{constant.DeliveryChannelWebPush},
		ChannelDeliveryStatus: json.RawMessage(`{"WEB_PUSH":"PENDING"}`),
	}

	subRepo := newMemPushSubRepo()
	_, _ = subRepo.Upsert(context.Background(), entity.PushSubscription{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000208"),
		UserID: userID, Endpoint: "https://push.example/bad", P256dh: "k", Auth: "a",
		CreatedAt: now, UpdatedAt: now,
	})

	sender := &stubPushSender{
		configured: true,
		fail:       map[string]error{"https://push.example/bad": errors.New("boom")},
	}
	d := NewNotificationDispatcher(notifRepo, subRepo, sender, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	n := notifRepo.byID[notifID]
	if n.Status != constant.NotificationStatusFailed {
		t.Fatalf("status = %s", n.Status)
	}
	var status map[string]string
	_ = json.Unmarshal(n.ChannelDeliveryStatus, &status)
	if status[constant.DeliveryChannelWebPush] != constant.ChannelDeliveryFailed {
		t.Fatalf("channel = %v", status)
	}

	// No retry: still FAILED after another tick.
	d.Tick(context.Background())
	if notifRepo.byID[notifID].Status != constant.NotificationStatusFailed {
		t.Fatal("should remain FAILED without retry")
	}
}

func TestDispatcherMultipleDevices(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000209")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000210")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "t", Body: "b",
		Status: constant.NotificationStatusPending, ScheduledFor: &scheduled,
		DeliveryChannels: []string{constant.DeliveryChannelWebPush},
		ChannelDeliveryStatus: json.RawMessage(`{"WEB_PUSH":"PENDING"}`),
	}

	subRepo := newMemPushSubRepo()
	_, _ = subRepo.Upsert(context.Background(), entity.PushSubscription{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000211"),
		UserID: userID, Endpoint: "https://push.example/phone", P256dh: "k1", Auth: "a1",
		CreatedAt: now, UpdatedAt: now,
	})
	_, _ = subRepo.Upsert(context.Background(), entity.PushSubscription{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000212"),
		UserID: userID, Endpoint: "https://push.example/laptop", P256dh: "k2", Auth: "a2",
		CreatedAt: now, UpdatedAt: now,
	})

	sender := &stubPushSender{
		configured: true,
		fail:       map[string]error{"https://push.example/phone": errors.New("dead")},
		gone:       map[string]bool{"https://push.example/phone": true},
	}
	d := NewNotificationDispatcher(notifRepo, subRepo, sender, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if notifRepo.byID[notifID].Status != constant.NotificationStatusSent {
		t.Fatalf("one device success should SENT, got %s", notifRepo.byID[notifID].Status)
	}
	if len(sender.sent) != 2 {
		t.Fatalf("expected both devices attempted, got %d", len(sender.sent))
	}
	live, _ := subRepo.ListByUser(context.Background(), userID)
	if len(live) != 1 || live[0].Endpoint != "https://push.example/laptop" {
		t.Fatalf("gone endpoint should be removed: %+v", live)
	}
}

func TestDispatcherNoSubscriptionsFails(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000213")
	notifID := uuid.MustParse("018f0000-0000-7000-8000-000000000214")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(-time.Minute)

	notifRepo := newMemNotificationRepo()
	notifRepo.byID[notifID] = entity.Notification{
		ID: notifID, UserID: userID, Title: "t", Body: "b",
		Status: constant.NotificationStatusPending, ScheduledFor: &scheduled,
		DeliveryChannels: []string{constant.DeliveryChannelWebPush},
		ChannelDeliveryStatus: json.RawMessage(`{"WEB_PUSH":"PENDING"}`),
	}

	d := NewNotificationDispatcher(notifRepo, newMemPushSubRepo(), &stubPushSender{configured: true}, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if notifRepo.byID[notifID].Status != constant.NotificationStatusFailed {
		t.Fatalf("status = %s", notifRepo.byID[notifID].Status)
	}
}

func TestPermissionDeniedMeansNoSubscribeCall(t *testing.T) {
	t.Parallel()
	// Client contract: Notification.permission === "denied" → never POST /push/subscribe.
	// Server stays empty for that device until permission is granted later.
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000218")
	repo := newMemPushSubRepo()
	svc := NewPushSubscriptionService(repo)
	listed, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected no subscriptions when client withholds permission, got %d", len(listed))
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
		DeliveryChannels: []string{constant.DeliveryChannelWebPush},
	}
	subRepo := newMemPushSubRepo()
	_, _ = subRepo.Upsert(context.Background(), entity.PushSubscription{
		ID: uuid.MustParse("018f0000-0000-7000-8000-000000000217"),
		UserID: userID, Endpoint: "https://push.example/x", P256dh: "k", Auth: "a",
		CreatedAt: now, UpdatedAt: now,
	})
	sender := &stubPushSender{configured: true}
	d := NewNotificationDispatcher(notifRepo, subRepo, sender, nil)
	d.now = func() time.Time { return now }
	d.Tick(context.Background())

	if notifRepo.byID[notifID].Status != constant.NotificationStatusPending {
		t.Fatalf("future should stay PENDING, got %s", notifRepo.byID[notifID].Status)
	}
	if len(sender.sent) != 0 {
		t.Fatal("should not send future notifications")
	}
}
