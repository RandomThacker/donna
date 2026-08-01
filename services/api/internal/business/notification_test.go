package business

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestNotificationPolicyOffsets(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	item := occurrence.Occurrence{StartAt: start, Type: occurrence.TypeEvent}

	if got := (googleEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-10 * time.Minute)) {
		t.Fatalf("google = %v", got)
	}
	if got := (microsoftICSEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-10 * time.Minute)) {
		t.Fatalf("ics = %v", got)
	}
	if got := (donnaEventNotificationPolicy{}).ReminderTime(item); !got.Equal(start.Add(-15 * time.Minute)) {
		t.Fatalf("donna event = %v", got)
	}
	reminder := occurrence.Occurrence{StartAt: start, Type: occurrence.TypeReminder}
	if got := (donnaReminderNotificationPolicy{}).ReminderTime(reminder); !got.Equal(start) {
		t.Fatalf("donna reminder = %v", got)
	}
}

func TestNotificationPolicyResolver(t *testing.T) {
	t.Parallel()
	r := NewNotificationPolicyResolver()
	cases := []struct {
		source occurrence.OccurrenceSource
		typ    occurrence.OccurrenceType
		wantNil bool
	}{
		{occurrence.SourceGoogle, occurrence.TypeEvent, false},
		{occurrence.SourceMicrosoftICS, occurrence.TypeEvent, false},
		{occurrence.SourceDonna, occurrence.TypeEvent, false},
		{occurrence.SourceDonna, occurrence.TypeReminder, false},
		{occurrence.SourceGoogle, occurrence.TypeReminder, true},
	}
	for _, tc := range cases {
		p := r.Resolve(occurrence.Occurrence{Source: tc.source, Type: tc.typ})
		if tc.wantNil && p != nil {
			t.Fatalf("%s/%s: expected nil", tc.source, tc.typ)
		}
		if !tc.wantNil && p == nil {
			t.Fatalf("%s/%s: expected policy", tc.source, tc.typ)
		}
	}
}

type stubOccurrenceProvider struct {
	items []occurrence.Occurrence
}

func (s stubOccurrenceProvider) ListOccurrences(
	context.Context,
	uuid.UUID,
	time.Time,
	time.Time,
) ([]occurrence.Occurrence, error) {
	out := make([]occurrence.Occurrence, len(s.items))
	copy(out, s.items)
	return out, nil
}

func newTestOccurrenceService(items ...occurrence.Occurrence) occurrence.Service {
	return occurrence.NewService(occurrence.ServiceDeps{
		Providers: []occurrence.Provider{
			stubOccurrenceProvider{items: items},
		},
	})
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
	allow := map[string]struct{}{}
	for _, s := range statuses {
		allow[s] = struct{}{}
	}
	out := make([]entity.Notification, 0)
	for _, n := range m.byID {
		if n.UserID != userID {
			continue
		}
		if len(allow) > 0 {
			if _, ok := allow[n.Status]; !ok {
				continue
			}
		}
		out = append(out, n)
	}
	return out, nil
}

func (m *memNotificationRepo) MarkRead(_ context.Context, id, userID uuid.UUID, readAt time.Time) (entity.Notification, error) {
	n, ok := m.byID[id]
	if !ok || n.UserID != userID {
		return entity.Notification{}, apperr.ErrNotFound
	}
	n.Status = constant.NotificationStatusRead
	n.ReadAt = &readAt
	n.UpdatedAt = readAt
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

func (m *memNotificationRepo) ListRecentChatDelivered(_ context.Context, since time.Time, limit int) ([]entity.Notification, error) {
	out := make([]entity.Notification, 0)
	for _, n := range m.byID {
		if n.DeletedAt != nil {
			continue
		}
		if n.Status != constant.NotificationStatusSent && n.Status != constant.NotificationStatusRead {
			continue
		}
		if n.SentAt == nil || n.SentAt.Before(since) {
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
	svc := NewNotificationService(repo, newTestOccurrenceService(occurrence.Occurrence{
		ID:           itemID,
		OccurrenceID: itemID,
		UserID:       userID,
		Source:       occurrence.SourceDonna,
		Type:         occurrence.TypeEvent,
		Status:       occurrence.StatusActive,
		Title:        "Guitar",
		StartAt:      start,
		EndAt:        start.Add(time.Hour),
		Timezone:     "UTC",
	}), NewNotificationPolicyResolver())
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
	svc := NewNotificationService(repo, newTestOccurrenceService(occurrence.Occurrence{
		ID:           occ,
		OccurrenceID: occ,
		ParentID:     &parent,
		UserID:       userID,
		Source:       occurrence.SourceDonna,
		Type:         occurrence.TypeReminder,
		Status:       occurrence.StatusActive,
		Title:        "Drink water",
		StartAt:      trigger,
		EndAt:        trigger,
		Timezone:     "UTC",
	}), NewNotificationPolicyResolver())
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
	svc := NewNotificationService(repo, newTestOccurrenceService(occurrence.Occurrence{
		ID:           occ,
		OccurrenceID: occ,
		ParentID:     &parent,
		UserID:       userID,
		Source:       occurrence.SourceDonna,
		Type:         occurrence.TypeEvent,
		Status:       occurrence.StatusActive,
		Title:        "Class",
		StartAt:      start,
		EndAt:        start.Add(time.Hour),
		Timezone:     "UTC",
	}), NewNotificationPolicyResolver())
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
	svc := NewNotificationService(repo, newTestOccurrenceService(occurrence.Occurrence{
		ID: "g1", OccurrenceID: "g1", UserID: userID,
		Source: occurrence.SourceGoogle, Type: occurrence.TypeEvent,
		Status: occurrence.StatusActive, Title: "Meet",
		StartAt: start, EndAt: start.Add(30 * time.Minute), Timezone: "UTC",
	}), NewNotificationPolicyResolver())
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

func TestSchedulerTickEmitsFeedStats(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000107")
	now := time.Date(2026, 7, 30, 12, 40, 0, 0, time.UTC)
	start := now.Add(20 * time.Minute)
	repo := newMemNotificationRepo()
	occSvc := occurrence.NewService(occurrence.ServiceDeps{
		Providers: []occurrence.Provider{
			stubOccurrenceProvider{items: []occurrence.Occurrence{{
				ID: "a", OccurrenceID: "a", UserID: userID,
				Source: occurrence.SourceDonna, Type: occurrence.TypeEvent,
				Status: occurrence.StatusActive, Title: "A",
				StartAt: start, EndAt: start.Add(time.Hour), Timezone: "UTC",
			}}},
			stubOccurrenceProvider{items: nil},
			nil,
		},
	})
	svc := NewNotificationService(repo, occSvc, NewNotificationPolicyResolver())
	svc.now = func() time.Time { return now }

	created, stats, err := svc.enqueueForUser(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("created = %d", created)
	}
	if stats.FeedSource != FeedSourceOccurrence {
		t.Fatalf("feed = %s", stats.FeedSource)
	}
	if stats.ProvidersQueried != 2 {
		t.Fatalf("providers = %d", stats.ProvidersQueried)
	}
	if stats.OccurrencesReturned != 1 || stats.AfterExpansion != 1 || stats.AfterDedup != 1 {
		t.Fatalf("pipeline = %+v", stats)
	}
	if stats.Notifications != 1 {
		t.Fatalf("notifications = %d", stats.Notifications)
	}
	if stats.DatabaseQueries < 3 { // 2 provider fetches + 1 CreateIdempotent
		t.Fatalf("db_queries = %d", stats.DatabaseQueries)
	}
	if !stats.WindowFrom.Equal(now) || !stats.WindowTo.Equal(now.Add(constant.NotificationLookaheadWindow)) {
		t.Fatalf("window = %v → %v", stats.WindowFrom, stats.WindowTo)
	}
}

func TestNotificationServiceDoesNotUseTimeline(t *testing.T) {
	t.Parallel()
	svc := NewNotificationService(newMemNotificationRepo(), newTestOccurrenceService(), NewNotificationPolicyResolver())
	if svc.occurrences == nil {
		t.Fatal("expected occurrence service")
	}
}

func TestTimelineOccurrenceSchedulingEquivalence(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	windowEnd := now.Add(constant.NotificationLookaheadWindow)
	policies := NewNotificationPolicyResolver()

	type pair struct {
		name string
		tl   entity.TimelineItem
		occ  occurrence.Occurrence
	}
	parent := "series-parent"
	cases := []pair{
		{
			name: "google_event",
			tl: entity.TimelineItem{
				ID: "g1", OccurrenceID: "g1",
				Source: constant.TimelineSourceGoogle, Type: constant.TimelineTypeEvent,
				Status: constant.TimelineStatusActive, Title: "Meet",
				StartAt: now.Add(20 * time.Minute), EndAt: now.Add(50 * time.Minute), Timezone: "UTC",
			},
			occ: occurrence.Occurrence{
				ID: "g1", OccurrenceID: "g1",
				Source: occurrence.SourceGoogle, Type: occurrence.TypeEvent,
				Status: occurrence.StatusActive, Title: "Meet",
				StartAt: now.Add(20 * time.Minute), EndAt: now.Add(50 * time.Minute), Timezone: "UTC",
			},
		},
		{
			name: "microsoft_event",
			tl: entity.TimelineItem{
				ID: "m1", OccurrenceID: "m1",
				Source: constant.TimelineSourceMicrosoftICS, Type: constant.TimelineTypeEvent,
				Status: constant.TimelineStatusActive, Title: "Sync",
				StartAt: now.Add(18 * time.Minute), EndAt: now.Add(48 * time.Minute), Timezone: "UTC",
			},
			occ: occurrence.Occurrence{
				ID: "m1", OccurrenceID: "m1",
				Source: occurrence.SourceMicrosoftICS, Type: occurrence.TypeEvent,
				Status: occurrence.StatusActive, Title: "Sync",
				StartAt: now.Add(18 * time.Minute), EndAt: now.Add(48 * time.Minute), Timezone: "UTC",
			},
		},
		{
			name: "donna_event",
			tl: entity.TimelineItem{
				ID: "d1", OccurrenceID: "d1",
				Source: constant.TimelineSourceDonna, Type: constant.TimelineTypeEvent,
				Status: constant.TimelineStatusActive, Title: "Guitar",
				StartAt: now.Add(25 * time.Minute), EndAt: now.Add(85 * time.Minute), Timezone: "UTC",
			},
			occ: occurrence.Occurrence{
				ID: "d1", OccurrenceID: "d1",
				Source: occurrence.SourceDonna, Type: occurrence.TypeEvent,
				Status: occurrence.StatusActive, Title: "Guitar",
				StartAt: now.Add(25 * time.Minute), EndAt: now.Add(85 * time.Minute), Timezone: "UTC",
			},
		},
		{
			name: "donna_reminder_recurring",
			tl: entity.TimelineItem{
				ID: parent + "_20260730T120500Z", OccurrenceID: parent + "_20260730T120500Z",
				ParentID: &parent,
				Source:   constant.TimelineSourceDonna, Type: constant.TimelineTypeReminder,
				Status: constant.TimelineStatusActive, Title: "Water",
				StartAt: now.Add(5 * time.Minute), EndAt: now.Add(5 * time.Minute), Timezone: "UTC",
			},
			occ: occurrence.Occurrence{
				ID: parent + "_20260730T120500Z", OccurrenceID: parent + "_20260730T120500Z",
				ParentID: &parent,
				Source:   occurrence.SourceDonna, Type: occurrence.TypeReminder,
				Status: occurrence.StatusActive, Title: "Water",
				StartAt: now.Add(5 * time.Minute), EndAt: now.Add(5 * time.Minute), Timezone: "UTC",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Legacy timeline policy selection (source/type string constants).
			var legacyReminder time.Time
			var legacyType string
			switch {
			case tc.tl.Source == constant.TimelineSourceGoogle && tc.tl.Type == constant.TimelineTypeEvent:
				legacyReminder = tc.tl.StartAt.Add(-10 * time.Minute)
				legacyType = constant.NotificationTypeEvent
			case tc.tl.Source == constant.TimelineSourceMicrosoftICS && tc.tl.Type == constant.TimelineTypeEvent:
				legacyReminder = tc.tl.StartAt.Add(-10 * time.Minute)
				legacyType = constant.NotificationTypeEvent
			case tc.tl.Source == constant.TimelineSourceDonna && tc.tl.Type == constant.TimelineTypeEvent:
				legacyReminder = tc.tl.StartAt.Add(-15 * time.Minute)
				legacyType = constant.NotificationTypeEvent
			case tc.tl.Source == constant.TimelineSourceDonna && tc.tl.Type == constant.TimelineTypeReminder:
				legacyReminder = tc.tl.StartAt
				legacyType = constant.NotificationTypeReminder
			default:
				t.Fatal("unexpected timeline case")
			}

			policy := policies.Resolve(tc.occ)
			if policy == nil {
				t.Fatal("expected occurrence policy")
			}
			gotReminder := policy.ReminderTime(tc.occ).UTC()
			if !gotReminder.Equal(legacyReminder.UTC()) {
				t.Fatalf("reminder timeline=%v occurrence=%v", legacyReminder, gotReminder)
			}
			if gotReminder.Before(now) || !gotReminder.Before(windowEnd) {
				t.Fatalf("reminder outside window: %v", gotReminder)
			}

			gotType := notificationTypeFromOccurrence(tc.occ.Type)
			if gotType != legacyType {
				t.Fatalf("type timeline=%s occurrence=%s", legacyType, gotType)
			}
			if tc.occ.OccurrenceID != tc.tl.OccurrenceID {
				t.Fatalf("occurrence id mismatch %s vs %s", tc.occ.OccurrenceID, tc.tl.OccurrenceID)
			}

			bodyOcc := notificationBody(tc.occ, gotReminder)
			var bodyTL string
			switch tc.tl.Type {
			case constant.TimelineTypeReminder:
				bodyTL = "Reminder is due"
			default:
				mins := int(tc.tl.StartAt.Sub(legacyReminder).Minutes())
				if mins < 1 {
					mins = 1
				}
				bodyTL = fmt.Sprintf("Starts in %d minutes", mins)
			}
			if bodyOcc != bodyTL {
				t.Fatalf("body timeline=%q occurrence=%q", bodyTL, bodyOcc)
			}

			payload, err := buildNotificationPayload(tc.occ, tc.occ.OccurrenceID)
			if err != nil {
				t.Fatal(err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded["timelineItemId"] != tc.tl.ID {
				t.Fatalf("payload id = %v", decoded["timelineItemId"])
			}
			if decoded["source"] != tc.tl.Source || decoded["type"] != tc.tl.Type {
				t.Fatalf("payload source/type = %v/%v", decoded["source"], decoded["type"])
			}
		})
	}
}

type stubUserLister struct {
	ids []uuid.UUID
}

func (s *stubUserLister) ListActiveIDs(context.Context) ([]uuid.UUID, error) {
	return s.ids, nil
}
