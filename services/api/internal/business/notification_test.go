package business

import (
	"context"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestNotificationPolicyOffsets(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	item := entity.TimelineItem{StartAt: start, Type: constant.TimelineTypeEvent}

	if got := (googleEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-10 * time.Minute)) {
		t.Fatalf("google = %v", got)
	}
	if got := (microsoftICSEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-10 * time.Minute)) {
		t.Fatalf("ics = %v", got)
	}
	if got := (donnaEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-15 * time.Minute)) {
		t.Fatalf("donna event = %v", got)
	}
	reminder := entity.TimelineItem{StartAt: start, Type: constant.TimelineTypeReminder}
	if got := (donnaReminderNotificationPolicy{}).ReminderTime(reminder); !got.Equal(start) {
		t.Fatalf("donna reminder = %v", got)
	}
}

func TestNotificationPolicyResolver(t *testing.T) {
	t.Parallel()
	r := NewNotificationPolicyResolver()
	cases := []struct {
		source, typ string
		wantNil     bool
	}{
		{constant.TimelineSourceGoogle, constant.TimelineTypeEvent, false},
		{constant.TimelineSourceMicrosoftICS, constant.TimelineTypeEvent, false},
		{constant.TimelineSourceDonna, constant.TimelineTypeEvent, false},
		{constant.TimelineSourceDonna, constant.TimelineTypeReminder, false},
		{constant.TimelineSourceGoogle, constant.TimelineTypeReminder, true},
	}
	for _, tc := range cases {
		p := r.Resolve(entity.TimelineItem{Source: tc.source, Type: tc.typ})
		if tc.wantNil && p != nil {
			t.Fatalf("%s/%s: expected nil", tc.source, tc.typ)
		}
		if !tc.wantNil && p == nil {
			t.Fatalf("%s/%s: expected policy", tc.source, tc.typ)
		}
	}
}

type stubTimelineProvider struct {
	items []entity.TimelineItem
}

func (s stubTimelineProvider) Fetch(context.Context, uuid.UUID, time.Time, time.Time) ([]entity.TimelineItem, error) {
	return s.items, nil
}

type memNotificationRepo struct {
	byKey map[string]entity.Notification
	byID  map[uuid.UUID]entity.Notification
}

func newMemNotificationRepo() *memNotificationRepo {
	return &memNotificationRepo{
		byKey: map[string]entity.Notification{},
		byID:  map[uuid.UUID]entity.Notification{},
	}
}

func (m *memNotificationRepo) key(occ, typ string) string { return occ + "|" + typ }

func (m *memNotificationRepo) CreateIdempotent(_ context.Context, n entity.Notification) (bool, entity.Notification, error) {
	if n.OccurrenceID != nil && n.NotificationType != nil {
		k := m.key(*n.OccurrenceID, *n.NotificationType)
		if _, ok := m.byKey[k]; ok {
			return false, entity.Notification{}, nil
		}
		m.byKey[k] = n
	}
	m.byID[n.ID] = n
	return true, n, nil
}

func (m *memNotificationRepo) GetByID(_ context.Context, id uuid.UUID) (entity.Notification, error) {
	n, ok := m.byID[id]
	if !ok {
		return entity.Notification{}, apperr.ErrNotFound
	}
	return n, nil
}

func (m *memNotificationRepo) ListByUser(_ context.Context, userID uuid.UUID, statuses []string) ([]entity.Notification, error) {
	allow := map[string]bool{}
	for _, s := range statuses {
		allow[s] = true
	}
	out := make([]entity.Notification, 0)
	for _, n := range m.byID {
		if n.UserID != userID {
			continue
		}
		if len(allow) > 0 && !allow[n.Status] {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

func (m *memNotificationRepo) MarkRead(_ context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error) {
	n, ok := m.byID[id]
	if !ok || n.UserID != userID {
		return entity.Notification{}, apperr.ErrNotFound
	}
	n.Status = constant.NotificationStatusRead
	n.ReadAt = &at
	n.UpdatedAt = at
	m.byID[id] = n
	return n, nil
}

func (m *memNotificationRepo) MarkDismissed(_ context.Context, id, userID uuid.UUID, at time.Time) (entity.Notification, error) {
	n, ok := m.byID[id]
	if !ok || n.UserID != userID {
		return entity.Notification{}, apperr.ErrNotFound
	}
	n.Status = constant.NotificationStatusDismissed
	n.DismissedAt = &at
	n.UpdatedAt = at
	m.byID[id] = n
	return n, nil
}

func (m *memNotificationRepo) ExistsByOccurrence(_ context.Context, occurrenceID, notificationType string) (bool, error) {
	_, ok := m.byKey[m.key(occurrenceID, notificationType)]
	return ok, nil
}

func (m *memNotificationRepo) ListDuePending(_ context.Context, asOf time.Time, limit int) ([]entity.Notification, error) {
	out := make([]entity.Notification, 0)
	for _, n := range m.byID {
		if n.DeletedAt != nil || n.Status != constant.NotificationStatusPending {
			continue
		}
		if n.ScheduledFor == nil || n.ScheduledFor.After(asOf) {
			continue
		}
		out = append(out, n)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *memNotificationRepo) UpdateDelivery(
	_ context.Context,
	id uuid.UUID,
	status string,
	channelStatus []byte,
	sentAt *time.Time,
	updatedAt time.Time,
) (entity.Notification, error) {
	n, ok := m.byID[id]
	if !ok {
		return entity.Notification{}, apperr.ErrNotFound
	}
	n.Status = status
	n.ChannelDeliveryStatus = channelStatus
	n.SentAt = sentAt
	n.UpdatedAt = updatedAt
	m.byID[id] = n
	return n, nil
}

func (m *memNotificationRepo) WithTx(pgx.Tx) repository.NotificationRepository {
	return m
}

var _ repository.NotificationRepository = (*memNotificationRepo)(nil)

func TestEnqueueOneTimeDonnaEvent(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000101")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	start := now.Add(20 * time.Minute) // ReminderTime = now+5m (15 min before)
	itemID := "evt-1"
	repo := newMemNotificationRepo()
	timeline := NewTimelineService(TimelineServiceDeps{
		Providers: []TimelineProvider{
			stubTimelineProvider{items: []entity.TimelineItem{{
				ID:           itemID,
				OccurrenceID: itemID,
				Source:       constant.TimelineSourceDonna,
				Type:         constant.TimelineTypeEvent,
				Status:       constant.TimelineStatusActive,
				Title:        "Guitar",
				StartAt:      start,
				EndAt:        start.Add(time.Hour),
				Timezone:     "UTC",
			}}},
		},
	})
	svc := NewNotificationService(repo, timeline, NewNotificationPolicyResolver())
	svc.now = func() time.Time { return now }

	created, err := svc.EnqueueForUser(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("created = %d", created)
	}
	created2, err := svc.EnqueueForUser(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if created2 != 0 {
		t.Fatalf("idempotent created = %d", created2)
	}
	if len(repo.byID) != 1 {
		t.Fatalf("rows = %d", len(repo.byID))
	}
}

func TestEnqueueRecurringReminder(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000102")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	trigger := now.Add(5 * time.Minute)
	parent := "rem-parent"
	occ := parent + "_20260730T120500Z"
	repo := newMemNotificationRepo()
	timeline := NewTimelineService(TimelineServiceDeps{
		Providers: []TimelineProvider{
			stubTimelineProvider{items: []entity.TimelineItem{{
				ID:           occ,
				OccurrenceID: occ,
				ParentID:     &parent,
				Source:       constant.TimelineSourceDonna,
				Type:         constant.TimelineTypeReminder,
				Status:       constant.TimelineStatusActive,
				Title:        "Drink water",
				StartAt:      trigger,
				EndAt:        trigger,
				Timezone:     "UTC",
				IsRecurring:  true,
			}}},
		},
	})
	svc := NewNotificationService(repo, timeline, NewNotificationPolicyResolver())
	svc.now = func() time.Time { return now }

	created, err := svc.EnqueueForUser(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("created = %d", created)
	}
	var n entity.Notification
	for _, row := range repo.byID {
		n = row
	}
	if n.NotificationType == nil || *n.NotificationType != constant.NotificationTypeReminder {
		t.Fatalf("type = %v", n.NotificationType)
	}
	if n.OccurrenceID == nil || *n.OccurrenceID != occ {
		t.Fatalf("occurrence = %v", n.OccurrenceID)
	}
}

func TestEnqueueRecurringEvent(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000103")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	start := now.Add(25 * time.Minute) // 15m policy → reminder at now+10m
	parent := "dev-parent"
	occ := "dev-parent_occ1"
	repo := newMemNotificationRepo()
	timeline := NewTimelineService(TimelineServiceDeps{
		Providers: []TimelineProvider{
			stubTimelineProvider{items: []entity.TimelineItem{{
				ID:           occ,
				OccurrenceID: occ,
				ParentID:     &parent,
				Source:       constant.TimelineSourceDonna,
				Type:         constant.TimelineTypeEvent,
				Status:       constant.TimelineStatusActive,
				Title:        "Class",
				StartAt:      start,
				EndAt:        start.Add(time.Hour),
				Timezone:     "UTC",
				IsRecurring:  true,
			}}},
		},
	})
	svc := NewNotificationService(repo, timeline, NewNotificationPolicyResolver())
	svc.now = func() time.Time { return now }

	created, err := svc.EnqueueForUser(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("created = %d", created)
	}
}

func TestNotificationReadAndDismiss(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000104")
	id := uuid.MustParse("018f0000-0000-7000-8000-000000000105")
	repo := newMemNotificationRepo()
	repo.byID[id] = entity.Notification{
		ID:     id,
		UserID: userID,
		Title:  "t",
		Body:   "b",
		Status: constant.NotificationStatusPending,
	}
	svc := NewNotificationService(repo, nil, nil)

	read, err := svc.MarkRead(context.Background(), userID, id)
	if err != nil {
		t.Fatal(err)
	}
	if read.Status != constant.NotificationStatusRead || read.ReadAt == nil {
		t.Fatalf("read = %+v", read)
	}

	dismissed, err := svc.MarkDismissed(context.Background(), userID, id)
	if err != nil {
		t.Fatal(err)
	}
	if dismissed.Status != constant.NotificationStatusDismissed || dismissed.DismissedAt == nil {
		t.Fatalf("dismissed = %+v", dismissed)
	}

	listed, err := svc.List(context.Background(), userID, []string{constant.NotificationStatusDismissed})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("list = %d", len(listed))
	}
}

func TestSchedulerIdempotency(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000106")
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	start := now.Add(20 * time.Minute)
	repo := newMemNotificationRepo()
	timeline := NewTimelineService(TimelineServiceDeps{
		Providers: []TimelineProvider{
			stubTimelineProvider{items: []entity.TimelineItem{{
				ID: "g1", OccurrenceID: "g1",
				Source: constant.TimelineSourceGoogle, Type: constant.TimelineTypeEvent,
				Status: constant.TimelineStatusActive, Title: "Meet",
				StartAt: start, EndAt: start.Add(30 * time.Minute), Timezone: "UTC",
			}}},
		},
	})
	svc := NewNotificationService(repo, timeline, NewNotificationPolicyResolver())
	svc.now = func() time.Time { return now }

	users := &stubUserLister{ids: []uuid.UUID{userID}}
	sched := NewNotificationScheduler(svc, users, nil)
	sched.now = func() time.Time { return now }

	sched.Tick(context.Background())
	sched.Tick(context.Background())
	if len(repo.byID) != 1 {
		t.Fatalf("rows after two ticks = %d", len(repo.byID))
	}
}

type stubUserLister struct {
	ids []uuid.UUID
}

func (s *stubUserLister) ListActiveIDs(context.Context) ([]uuid.UUID, error) {
	return s.ids, nil
}
