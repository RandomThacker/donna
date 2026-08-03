package business_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockEventRepo struct {
	byKey       map[string]entity.CalendarEvent
	updateCalls int
}

func (m *mockEventRepo) key(sourceID uuid.UUID, providerID string) string {
	return sourceID.String() + "|" + providerID
}

func (m *mockEventRepo) Create(_ context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error) {
	if m.byKey == nil {
		m.byKey = map[string]entity.CalendarEvent{}
	}
	pid := ""
	if event.ProviderEventID != nil {
		pid = *event.ProviderEventID
	}
	m.byKey[m.key(event.CalendarSourceID, pid)] = event
	return event, nil
}

func (m *mockEventRepo) GetBySourceAndProviderEvent(_ context.Context, sourceID uuid.UUID, providerEventID string) (entity.CalendarEvent, error) {
	event, ok := m.byKey[m.key(sourceID, providerEventID)]
	if !ok {
		return entity.CalendarEvent{}, apperr.ErrNotFound
	}
	return event, nil
}

func (m *mockEventRepo) ListByUserInRange(_ context.Context, userID uuid.UUID, from, to time.Time) ([]entity.CalendarEvent, error) {
	out := make([]entity.CalendarEvent, 0)
	for _, event := range m.byKey {
		if event.UserID != userID || event.DeletedAt != nil {
			continue
		}
		if event.StartsAt.Before(to) && event.EndsAt.After(from) {
			out = append(out, event)
		}
	}
	return out, nil
}

func (m *mockEventRepo) ListByUserInRangeWithProvider(
	_ context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.CalendarEventWithProvider, error) {
	events, err := m.ListByUserInRange(context.Background(), userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]entity.CalendarEventWithProvider, 0, len(events))
	for _, event := range events {
		out = append(out, entity.CalendarEventWithProvider{
			Event:    event,
			Provider: constant.AuthProviderGoogle,
		})
	}
	return out, nil
}

func (m *mockEventRepo) ListForSchedulerByUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	rows, err := m.ListByUserInRangeWithProvider(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return rows, nil
	}
	allow := map[string]struct{}{}
	for _, p := range providers {
		allow[p] = struct{}{}
	}
	out := make([]entity.CalendarEventWithProvider, 0, len(rows))
	for _, row := range rows {
		if _, ok := allow[row.Provider]; ok {
			out = append(out, row)
		}
	}
	return out, nil
}

func (m *mockEventRepo) ListCalendarOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
	providers []string,
) ([]entity.CalendarEventWithProvider, error) {
	return m.ListForSchedulerByUserInRange(ctx, userID, from, to, providers)
}

func (m *mockEventRepo) UpdateFromSync(_ context.Context, event entity.CalendarEvent) (entity.CalendarEvent, error) {
	m.updateCalls++
	pid := ""
	if event.ProviderEventID != nil {
		pid = *event.ProviderEventID
	}
	event.DeletedAt = nil
	m.byKey[m.key(event.CalendarSourceID, pid)] = event
	return event, nil
}

func (m *mockEventRepo) SoftDeleteByProviderEventID(_ context.Context, sourceID uuid.UUID, providerEventID string, deletedAt time.Time) (entity.CalendarEvent, error) {
	key := m.key(sourceID, providerEventID)
	event, ok := m.byKey[key]
	if !ok || event.DeletedAt != nil {
		return entity.CalendarEvent{}, apperr.ErrNotFound
	}
	event.Status = constant.CalendarEventStatusCancelled
	event.DeletedAt = &deletedAt
	event.UpdatedAt = deletedAt
	m.byKey[key] = event
	return event, nil
}

func (m *mockEventRepo) SoftDeleteMissing(_ context.Context, sourceID uuid.UUID, keepProviderIDs []string, deletedAt time.Time) (int64, error) {
	keep := map[string]struct{}{}
	for _, id := range keepProviderIDs {
		keep[id] = struct{}{}
	}
	var n int64
	for key, event := range m.byKey {
		if event.CalendarSourceID != sourceID || event.DeletedAt != nil || event.ProviderEventID == nil {
			continue
		}
		if _, ok := keep[*event.ProviderEventID]; ok {
			continue
		}
		event.DeletedAt = &deletedAt
		event.Status = constant.CalendarEventStatusCancelled
		m.byKey[key] = event
		n++
	}
	return n, nil
}

func (m *mockEventRepo) DeleteByConnectedAccountID(_ context.Context, _ uuid.UUID) (int64, error) {
	n := int64(len(m.byKey))
	m.byKey = map[string]entity.CalendarEvent{}
	return n, nil
}

func (m *mockEventRepo) DeleteOrphansForUser(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockEventRepo) CountLiveByConnectedAccountID(context.Context, uuid.UUID) (int64, error) {
	var n int64
	for _, event := range m.byKey {
		if event.DeletedAt == nil {
			n++
		}
	}
	return n, nil
}

func (m *mockEventRepo) WithTx(pgx.Tx) repository.CalendarEventRepository { return m }

type mockEventSourceRepo struct {
	sources []entity.CalendarSource
}

func (m *mockEventSourceRepo) Create(context.Context, entity.CalendarSource) (entity.CalendarSource, error) {
	return entity.CalendarSource{}, errors.New("unused")
}
func (m *mockEventSourceRepo) GetByAccountAndProviderCalendar(context.Context, uuid.UUID, string) (entity.CalendarSource, error) {
	return entity.CalendarSource{}, apperr.ErrNotFound
}
func (m *mockEventSourceRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.CalendarSource, error) {
	return m.sources, nil
}
func (m *mockEventSourceRepo) ListByConnectedAccountID(_ context.Context, accountID uuid.UUID) ([]entity.CalendarSource, error) {
	out := make([]entity.CalendarSource, 0)
	for _, source := range m.sources {
		if source.ConnectedAccountID == accountID {
			out = append(out, source)
		}
	}
	return out, nil
}
func (m *mockEventSourceRepo) UpdateFromSync(context.Context, entity.CalendarSource) (entity.CalendarSource, error) {
	return entity.CalendarSource{}, errors.New("unused")
}
func (m *mockEventSourceRepo) SoftDeleteMissing(context.Context, uuid.UUID, []string, time.Time) (int64, error) {
	return 0, nil
}
func (m *mockEventSourceRepo) SoftDeleteByProviderIDs(context.Context, uuid.UUID, []string, time.Time) (int64, error) {
	return 0, nil
}
func (m *mockEventSourceRepo) DeleteByConnectedAccountID(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockEventSourceRepo) DeleteOrphansForUser(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockEventSourceRepo) UpdateEventSyncState(_ context.Context, id uuid.UUID, syncCursor *string, lastSyncedAt, updatedAt time.Time) (entity.CalendarSource, error) {
	for i := range m.sources {
		if m.sources[i].ID == id {
			m.sources[i].SyncCursor = syncCursor
			m.sources[i].LastSyncedAt = &lastSyncedAt
			m.sources[i].UpdatedAt = updatedAt
			return m.sources[i], nil
		}
	}
	return entity.CalendarSource{}, apperr.ErrNotFound
}
func (m *mockEventSourceRepo) ClearEventSyncCursor(_ context.Context, id uuid.UUID, updatedAt time.Time) (entity.CalendarSource, error) {
	for i := range m.sources {
		if m.sources[i].ID == id {
			m.sources[i].SyncCursor = nil
			m.sources[i].UpdatedAt = updatedAt
			return m.sources[i], nil
		}
	}
	return entity.CalendarSource{}, apperr.ErrNotFound
}
func (m *mockEventSourceRepo) UpdateSyncEnabledByAccount(_ context.Context, accountID uuid.UUID, syncEnabled bool, updatedAt time.Time) (int64, error) {
	var n int64
	for i := range m.sources {
		if m.sources[i].ConnectedAccountID != accountID || m.sources[i].DeletedAt != nil {
			continue
		}
		m.sources[i].SyncEnabled = syncEnabled
		m.sources[i].UpdatedAt = updatedAt
		n++
	}
	return n, nil
}
func (m *mockEventSourceRepo) WithTx(pgx.Tx) repository.CalendarSourceRepository { return m }

type mockEventAccountRepo struct {
	account  entity.ConnectedAccount
	accounts []entity.ConnectedAccount
}

func (m *mockEventAccountRepo) Create(context.Context, entity.ConnectedAccount) (entity.ConnectedAccount, error) {
	return entity.ConnectedAccount{}, errors.New("unused")
}
func (m *mockEventAccountRepo) GetByID(context.Context, uuid.UUID) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) GetByProviderAccount(context.Context, string, string) (entity.ConnectedAccount, error) {
	return entity.ConnectedAccount{}, errors.New("unused")
}
func (m *mockEventAccountRepo) GetByUserAndProvider(context.Context, uuid.UUID, string) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.ConnectedAccount, error) {
	if len(m.accounts) > 0 {
		return m.accounts, nil
	}
	if m.account.ID == uuid.Nil {
		return []entity.ConnectedAccount{}, nil
	}
	return []entity.ConnectedAccount{m.account}, nil
}
func (m *mockEventAccountRepo) UpdateCredentials(context.Context, uuid.UUID, string, *time.Time, []string, string, time.Time) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) UpdateProfile(context.Context, uuid.UUID, *string, []byte, time.Time) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) MarkCalendarSyncRunning(context.Context, uuid.UUID, string, time.Time) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) RecordCalendarSync(context.Context, uuid.UUID, repository.CalendarSyncRecord) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) ClearCalendarListSyncToken(context.Context, uuid.UUID, time.Time) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) SoftDelete(context.Context, uuid.UUID, time.Time) (entity.ConnectedAccount, error) {
	return m.account, nil
}
func (m *mockEventAccountRepo) WithTx(pgx.Tx) repository.ConnectedAccountRepository { return m }

func eventTestLog() *logger.Logger {
	return logger.NewFactory(logger.Options{
		Service: "test", Environment: "test", Level: "error", Output: io.Discard,
	}).Module("calendar")
}

func newEventSyncService(
	t *testing.T,
	accounts *mockEventAccountRepo,
	secrets *mockCalendarSecretRepo,
	sources *mockEventSourceRepo,
	events *mockEventRepo,
	provider *mockCalendarProvider,
	key []byte,
) *business.CalendarService {
	t.Helper()
	return business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts,
		Secrets:  secrets,
		Sources:  sources,
		Events:   events,
		Tx:       fakeCalendarTx{},
		Providers: map[string]calendarprovider.Provider{
			constant.AuthProviderGoogle: provider,
		},
		Tokens: map[string]calendarprovider.TokenRefresher{
			constant.AuthProviderGoogle: mockTokenRefresher{},
		},
		SealKey: key,
		Log:     eventTestLog(),
	})
}

func TestSyncEventsFullIncrementalUpdateDelete(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-events-test-secret-key!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000801")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000802")
	sourceID := uuid.MustParse("01900000-0000-7000-8000-000000000803")
	nowUnix := time.Now().Add(time.Hour).Unix()

	accounts := &mockEventAccountRepo{account: entity.ConnectedAccount{
		ID:             accountID,
		UserID:         userID,
		Provider:       constant.AuthProviderGoogle,
		Status:         constant.ConnectedAccountStatusActive,
		Scopes:         []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_events",
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref:        "cred_events",
		Ciphertext: sealedCred(t, key, "access", "refresh", nowUnix),
	}}
	sources := &mockEventSourceRepo{sources: []entity.CalendarSource{{
		ID:                 sourceID,
		UserID:             userID,
		ConnectedAccountID: accountID,
		ProviderCalendarID: "primary",
		Name:               "Personal",
		SyncEnabled:        true,
	}}}
	events := &mockEventRepo{byKey: map[string]entity.CalendarEvent{}}
	start := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	api := &mockCalendarProvider{eventsByCal: map[string]calendarprovider.ListEventsResult{
		"primary": {
			NextSyncToken: "evt-sync-1",
			Events: []calendarprovider.RemoteEvent{
				{
					ID: "evt-1", Title: "Standup", StartsAt: start, EndsAt: end,
					Status: "confirmed", OrganizerEmail: "a@example.com",
					Attendees: []calendarprovider.RemoteAttendee{{Email: "b@example.com"}},
					UpdatedAt: start, Raw: map[string]any{"id": "evt-1"},
				},
				{
					ID: "evt-2", Title: "Planning", StartsAt: start.Add(2 * time.Hour), EndsAt: end.Add(2 * time.Hour),
					Status: "confirmed", Raw: map[string]any{"id": "evt-2"},
				},
			},
		},
	}}

	svc := newEventSyncService(t, accounts, secrets, sources, events, api, key)

	first, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatalf("full sync: %v", err)
	}
	if first.CreatedCount != 2 || first.UpdatedCount != 0 || first.RemovedCount != 0 {
		t.Fatalf("first counts = %+v", first)
	}
	if sources.sources[0].SyncCursor == nil || *sources.sources[0].SyncCursor != "evt-sync-1" {
		t.Fatalf("sync cursor = %#v", sources.sources[0].SyncCursor)
	}

	api.eventsByCal["primary"] = calendarprovider.ListEventsResult{
		NextSyncToken: "evt-sync-2",
		Incremental:   true,
		Events: []calendarprovider.RemoteEvent{
			{
				ID: "evt-1", Title: "Standup Renamed", StartsAt: start, EndsAt: end,
				Status: "confirmed", Raw: map[string]any{"id": "evt-1"},
			},
			{ID: "evt-2", Status: "cancelled", Deleted: true},
		},
	}
	second, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatalf("incremental sync: %v", err)
	}
	if second.CreatedCount != 0 || second.UpdatedCount != 1 || second.RemovedCount != 1 {
		t.Fatalf("second counts = %+v", second)
	}
	if len(api.eventCalls) < 2 || api.eventCalls[1] != "primary|evt-sync-1" {
		t.Fatalf("expected incremental token call, calls=%v", api.eventCalls)
	}
	got, err := events.GetBySourceAndProviderEvent(context.Background(), sourceID, "evt-1")
	if err != nil || got.Title != "Standup Renamed" {
		t.Fatalf("updated event = %#v err=%v", got, err)
	}
	deleted, err := events.GetBySourceAndProviderEvent(context.Background(), sourceID, "evt-2")
	if err != nil || deleted.DeletedAt == nil {
		t.Fatalf("deleted event = %#v err=%v", deleted, err)
	}
}

func TestSyncEventsSkipsUnchangedWithoutUpdate(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-events-test-secret-key!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000831")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000832")
	sourceID := uuid.MustParse("01900000-0000-7000-8000-000000000833")
	nowUnix := time.Now().Add(time.Hour).Unix()
	accounts := &mockEventAccountRepo{account: entity.ConnectedAccount{
		ID: accountID, UserID: userID, Provider: constant.AuthProviderGoogle,
		Status: constant.ConnectedAccountStatusActive, Scopes: []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_events",
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref: "cred_events", Ciphertext: sealedCred(t, key, "access", "refresh", nowUnix),
	}}
	sources := &mockEventSourceRepo{sources: []entity.CalendarSource{{
		ID: sourceID, UserID: userID, ConnectedAccountID: accountID,
		ProviderCalendarID: "primary", Name: "Personal", SyncEnabled: true,
	}}}
	events := &mockEventRepo{byKey: map[string]entity.CalendarEvent{}}
	start := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	etag := "etag-stable"
	remote := calendarprovider.RemoteEvent{
		ID: "evt-1", Title: "Standup", StartsAt: start, EndsAt: end,
		Status: "confirmed", ETag: etag, UpdatedAt: start,
		Raw: map[string]any{"id": "evt-1"},
	}
	api := &mockCalendarProvider{eventsByCal: map[string]calendarprovider.ListEventsResult{
		"primary": {NextSyncToken: "tok-1", Events: []calendarprovider.RemoteEvent{remote}},
	}}
	svc := newEventSyncService(t, accounts, secrets, sources, events, api, key)

	first, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if first.CreatedCount != 1 || events.updateCalls != 0 {
		t.Fatalf("first = %+v updates=%d", first, events.updateCalls)
	}

	api.eventsByCal["primary"] = calendarprovider.ListEventsResult{
		NextSyncToken: "tok-2", Incremental: true,
		Events: []calendarprovider.RemoteEvent{remote},
	}
	second, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if second.SkippedCount != 1 || second.UpdatedCount != 0 || events.updateCalls != 0 {
		t.Fatalf("second = %+v updates=%d", second, events.updateCalls)
	}
}

func TestSyncEventsRecoversFromGoneToken(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-events-test-secret-key!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000811")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000812")
	sourceID := uuid.MustParse("01900000-0000-7000-8000-000000000813")
	tok := "stale"
	accounts := &mockEventAccountRepo{account: entity.ConnectedAccount{
		ID: accountID, UserID: userID, Provider: constant.AuthProviderGoogle,
		Status: constant.ConnectedAccountStatusActive, Scopes: []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_events",
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref: "cred_events", Ciphertext: sealedCred(t, key, "access", "refresh", time.Now().Add(time.Hour).Unix()),
	}}
	sources := &mockEventSourceRepo{sources: []entity.CalendarSource{{
		ID: sourceID, UserID: userID, ConnectedAccountID: accountID,
		ProviderCalendarID: "primary", SyncEnabled: true, SyncCursor: &tok,
	}}}
	start := time.Now().UTC().Truncate(time.Hour)
	api := &mockCalendarProvider{
		eventsGoneOnce: map[string]bool{"primary": true},
		eventsByCal: map[string]calendarprovider.ListEventsResult{
			"primary": {
				NextSyncToken: "fresh",
				Events: []calendarprovider.RemoteEvent{{
					ID: "evt-9", Title: "Recovered", StartsAt: start, EndsAt: start.Add(time.Hour),
					Status: "confirmed", Raw: map[string]any{"id": "evt-9"},
				}},
			},
		},
	}
	events := &mockEventRepo{byKey: map[string]entity.CalendarEvent{}}
	svc := newEventSyncService(t, accounts, secrets, sources, events, api, key)

	result, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedCount != 1 {
		t.Fatalf("result = %+v", result)
	}
	if sources.sources[0].SyncCursor == nil || *sources.sources[0].SyncCursor != "fresh" {
		t.Fatalf("cursor = %#v", sources.sources[0].SyncCursor)
	}
}

func TestSyncEventsReplaceAllSoftDeletesMissing(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-events-test-secret-key!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000901")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000902")
	sourceID := uuid.MustParse("01900000-0000-7000-8000-000000000903")
	accounts := &mockEventAccountRepo{account: entity.ConnectedAccount{
		ID: accountID, UserID: userID, Provider: constant.AuthProviderICS,
		Status: constant.ConnectedAccountStatusActive, Scopes: []string{constant.ICSScopeCalendar},
		CredentialsRef: "cred_ics",
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref: "cred_ics", Ciphertext: sealedCred(t, key, "https://example.com/feed.ics", "", time.Now().Add(time.Hour).Unix()),
	}}
	sources := &mockEventSourceRepo{sources: []entity.CalendarSource{{
		ID: sourceID, UserID: userID, ConnectedAccountID: accountID,
		ProviderCalendarID: "ics_feed", SyncEnabled: true,
	}}}
	start := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	api := &mockCalendarProvider{eventsByCal: map[string]calendarprovider.ListEventsResult{
		"ics_feed": {
			ReplaceAll: true,
			Events: []calendarprovider.RemoteEvent{
				{ID: "keep", Title: "Keep", StartsAt: start, EndsAt: end, Status: "confirmed"},
				{ID: "gone-later", Title: "Temp", StartsAt: start, EndsAt: end, Status: "confirmed"},
			},
		},
	}}
	events := &mockEventRepo{byKey: map[string]entity.CalendarEvent{}}
	svc := business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts, Secrets: secrets, Sources: sources, Events: events,
		Tx: fakeCalendarTx{},
		Providers: map[string]calendarprovider.Provider{constant.AuthProviderICS: api},
		SealKey:   key,
		Log:       eventTestLog(),
	})

	if _, err := svc.SyncEvents(context.Background(), userID); err != nil {
		t.Fatal(err)
	}
	api.eventsByCal["ics_feed"] = calendarprovider.ListEventsResult{
		ReplaceAll: true,
		Events: []calendarprovider.RemoteEvent{
			{ID: "keep", Title: "Keep", StartsAt: start, EndsAt: end, Status: "confirmed"},
		},
	}
	second, err := svc.SyncEvents(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if second.RemovedCount != 1 {
		t.Fatalf("want soft-delete of missing event, got %+v", second)
	}
	deleted, err := events.GetBySourceAndProviderEvent(context.Background(), sourceID, "gone-later")
	if err != nil || deleted.DeletedAt == nil {
		t.Fatalf("missing event not soft-deleted: %#v err=%v", deleted, err)
	}
}

func TestListEventsReadsLocalDB(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000821")
	sourceID := uuid.MustParse("01900000-0000-7000-8000-000000000822")
	start := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	pid := "evt-local"
	events := &mockEventRepo{byKey: map[string]entity.CalendarEvent{
		sourceID.String() + "|evt-local": {
			ID:     uuid.MustParse("01900000-0000-7000-8000-000000000823"),
			UserID: userID, CalendarSourceID: sourceID, Title: "Local",
			StartsAt: start, EndsAt: start.Add(time.Hour), ProviderEventID: &pid,
			AttendeesSummary: []byte(`[]`), Origin: constant.CalendarEventOriginProviderSync,
		},
	}}
	key, _ := seal.KeyFromSecret("calendar-events-test-secret-key!!")
	svc := newEventSyncService(t, &mockEventAccountRepo{}, &mockCalendarSecretRepo{}, &mockEventSourceRepo{}, events, &mockCalendarProvider{}, key)

	got, err := svc.ListEvents(context.Background(), userID, start.Add(-time.Hour), start.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "Local" {
		t.Fatalf("got = %#v", got)
	}
}
