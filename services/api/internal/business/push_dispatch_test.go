package business

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
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

func TestPermissionDeniedMeansNoSubscribeCall(t *testing.T) {
	t.Parallel()
	// Client contract: Notification.permission === "denied" → never POST /push/subscribe.
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
