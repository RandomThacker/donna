package business_test

import (
	"context"
	"encoding/json"
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

type mockCalendarProvider struct {
	name           string
	listCalls      []calendarprovider.ListCalendarsOptions
	listResult     calendarprovider.ListCalendarsResult
	listErr        error
	gone           bool
	listAfterGone  calendarprovider.ListCalendarsResult
	eventsByCal    map[string]calendarprovider.ListEventsResult
	eventsGoneOnce map[string]bool
	eventCalls     []string
}

func (m *mockCalendarProvider) Name() string {
	if m.name != "" {
		return m.name
	}
	return constant.AuthProviderGoogle
}

func (m *mockCalendarProvider) ListCalendars(_ context.Context, _ string, opts calendarprovider.ListCalendarsOptions) (calendarprovider.ListCalendarsResult, error) {
	m.listCalls = append(m.listCalls, opts)
	if m.gone && opts.SyncToken != "" {
		return calendarprovider.ListCalendarsResult{}, &calendarprovider.SyncCursorInvalidError{Body: "sync token expired"}
	}
	if m.listErr != nil {
		return calendarprovider.ListCalendarsResult{}, m.listErr
	}
	if m.gone && opts.SyncToken == "" && len(m.listAfterGone.Calendars) > 0 {
		return m.listAfterGone, nil
	}
	return m.listResult, nil
}

func (m *mockCalendarProvider) ListEvents(_ context.Context, _ string, calendarID string, opts calendarprovider.ListEventsOptions) (calendarprovider.ListEventsResult, error) {
	m.eventCalls = append(m.eventCalls, calendarID+"|"+opts.SyncToken)
	if m.eventsGoneOnce != nil && m.eventsGoneOnce[calendarID] && opts.SyncToken != "" {
		m.eventsGoneOnce[calendarID] = false
		return calendarprovider.ListEventsResult{}, &calendarprovider.SyncCursorInvalidError{Body: "gone"}
	}
	if m.eventsByCal != nil {
		if result, ok := m.eventsByCal[calendarID]; ok {
			if opts.SyncToken != "" {
				result.Incremental = true
			}
			return result, nil
		}
	}
	return calendarprovider.ListEventsResult{NextSyncToken: "evt-token"}, nil
}

type mockTokenRefresher struct {
	refresh calendarprovider.TokenSet
	err     error
}

func (m mockTokenRefresher) RefreshAccessToken(context.Context, string) (calendarprovider.TokenSet, error) {
	if m.err != nil {
		return calendarprovider.TokenSet{}, m.err
	}
	return m.refresh, nil
}

type mockCalendarAccountRepo struct {
	account  entity.ConnectedAccount
	accounts []entity.ConnectedAccount
	err      error
	recorded int
	cleared  int
	running  int
}

func (m *mockCalendarAccountRepo) Create(context.Context, entity.ConnectedAccount) (entity.ConnectedAccount, error) {
	return entity.ConnectedAccount{}, errors.New("unexpected Create")
}

func (m *mockCalendarAccountRepo) GetByID(context.Context, uuid.UUID) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
}

func (m *mockCalendarAccountRepo) GetByProviderAccount(context.Context, string, string) (entity.ConnectedAccount, error) {
	return entity.ConnectedAccount{}, errors.New("unexpected GetByProviderAccount")
}

func (m *mockCalendarAccountRepo) GetByUserAndProvider(context.Context, uuid.UUID, string) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
}

func (m *mockCalendarAccountRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.ConnectedAccount, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.accounts) > 0 {
		return m.accounts, nil
	}
	if m.account.ID == uuid.Nil {
		return []entity.ConnectedAccount{}, nil
	}
	return []entity.ConnectedAccount{m.account}, nil
}

func (m *mockCalendarAccountRepo) UpdateCredentials(_ context.Context, id uuid.UUID, credentialsRef string, tokenExpiresAt *time.Time, scopes []string, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CredentialsRef = credentialsRef
	m.account.TokenExpiresAt = tokenExpiresAt
	m.account.Scopes = scopes
	m.account.Status = status
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) UpdateProfile(_ context.Context, id uuid.UUID, displayName *string, providerMetadata []byte, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.DisplayName = displayName
	m.account.ProviderMetadata = providerMetadata
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) MarkCalendarSyncRunning(_ context.Context, id uuid.UUID, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.running++
	m.account.ID = id
	m.account.CalendarSyncStatus = status
	m.account.UpdatedAt = updatedAt
	for i := range m.accounts {
		if m.accounts[i].ID == id {
			m.accounts[i].CalendarSyncStatus = status
			m.accounts[i].UpdatedAt = updatedAt
		}
	}
	return m.account, nil
}

func (m *mockCalendarAccountRepo) RecordCalendarSync(_ context.Context, id uuid.UUID, record repository.CalendarSyncRecord) (entity.ConnectedAccount, error) {
	m.recorded++
	apply := func(a *entity.ConnectedAccount) {
		a.ID = id
		a.CalendarSyncStatus = record.Status
		a.CalendarListSyncToken = record.ListSyncToken
		a.LastSyncedAt = record.SuccessfulAt
		a.LastFailedSyncAt = record.FailedAt
		ms := record.DurationMs
		a.LastSyncDurationMs = &ms
		a.LastSyncCreatedCount = record.CreatedCount
		a.LastSyncUpdatedCount = record.UpdatedCount
		a.LastSyncDeletedCount = record.DeletedCount
		a.UpdatedAt = record.UpdatedAt
	}
	apply(&m.account)
	for i := range m.accounts {
		if m.accounts[i].ID == id {
			apply(&m.accounts[i])
		}
	}
	return m.account, nil
}

func (m *mockCalendarAccountRepo) ClearCalendarListSyncToken(_ context.Context, id uuid.UUID, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.cleared++
	m.account.ID = id
	m.account.CalendarListSyncToken = nil
	m.account.UpdatedAt = updatedAt
	for i := range m.accounts {
		if m.accounts[i].ID == id {
			m.accounts[i].CalendarListSyncToken = nil
			m.accounts[i].UpdatedAt = updatedAt
		}
	}
	return m.account, nil
}

func (m *mockCalendarAccountRepo) SoftDelete(_ context.Context, id uuid.UUID, deletedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.Status = constant.ConnectedAccountStatusDisconnected
	m.account.DeletedAt = &deletedAt
	m.account.UpdatedAt = deletedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) WithTx(pgx.Tx) repository.ConnectedAccountRepository { return m }

type mockCalendarSecretRepo struct {
	secret  entity.CredentialSecret
	secrets map[string]entity.CredentialSecret
	err     error
}

func (m *mockCalendarSecretRepo) Create(context.Context, entity.CredentialSecret) (entity.CredentialSecret, error) {
	return entity.CredentialSecret{}, errors.New("unexpected Create")
}

func (m *mockCalendarSecretRepo) GetByRef(_ context.Context, ref string) (entity.CredentialSecret, error) {
	if m.err != nil {
		return entity.CredentialSecret{}, m.err
	}
	if m.secrets != nil {
		if secret, ok := m.secrets[ref]; ok {
			return secret, nil
		}
	}
	return m.secret, nil
}

func (m *mockCalendarSecretRepo) UpdateCiphertext(_ context.Context, ref string, ciphertext []byte, updatedAt time.Time) (entity.CredentialSecret, error) {
	m.secret.Ref = ref
	m.secret.Ciphertext = ciphertext
	m.secret.UpdatedAt = updatedAt
	if m.secrets != nil {
		m.secrets[ref] = m.secret
	}
	return m.secret, nil
}

func (m *mockCalendarSecretRepo) WithTx(pgx.Tx) repository.CredentialSecretRepository { return m }

type mockSourceRepo struct {
	byKey   map[string]entity.CalendarSource
	byUser  []entity.CalendarSource
	created int
	updated int
	removed int64
}

func (m *mockSourceRepo) Create(_ context.Context, source entity.CalendarSource) (entity.CalendarSource, error) {
	m.created++
	key := source.ConnectedAccountID.String() + "|" + source.ProviderCalendarID
	m.byKey[key] = source
	return source, nil
}

func (m *mockSourceRepo) GetByAccountAndProviderCalendar(_ context.Context, accountID uuid.UUID, providerCalendarID string) (entity.CalendarSource, error) {
	key := accountID.String() + "|" + providerCalendarID
	source, ok := m.byKey[key]
	if !ok {
		return entity.CalendarSource{}, apperr.ErrNotFound
	}
	return source, nil
}

func (m *mockSourceRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.CalendarSource, error) {
	if len(m.byUser) > 0 {
		return m.byUser, nil
	}
	out := make([]entity.CalendarSource, 0, len(m.byKey))
	for _, source := range m.byKey {
		if source.DeletedAt == nil {
			out = append(out, source)
		}
	}
	return out, nil
}

func (m *mockSourceRepo) ListByConnectedAccountID(_ context.Context, accountID uuid.UUID) ([]entity.CalendarSource, error) {
	out := make([]entity.CalendarSource, 0)
	for _, source := range m.byKey {
		if source.ConnectedAccountID == accountID && source.DeletedAt == nil {
			out = append(out, source)
		}
	}
	return out, nil
}

func (m *mockSourceRepo) UpdateFromSync(_ context.Context, source entity.CalendarSource) (entity.CalendarSource, error) {
	m.updated++
	source.DeletedAt = nil
	key := source.ConnectedAccountID.String() + "|" + source.ProviderCalendarID
	m.byKey[key] = source
	return source, nil
}

func (m *mockSourceRepo) SoftDeleteMissing(_ context.Context, accountID uuid.UUID, keepProviderIDs []string, deletedAt time.Time) (int64, error) {
	keep := map[string]struct{}{}
	for _, id := range keepProviderIDs {
		keep[id] = struct{}{}
	}
	var n int64
	for key, source := range m.byKey {
		if source.ConnectedAccountID != accountID || source.DeletedAt != nil {
			continue
		}
		if _, ok := keep[source.ProviderCalendarID]; ok {
			continue
		}
		source.DeletedAt = &deletedAt
		m.byKey[key] = source
		n++
	}
	m.removed = n
	return n, nil
}

func (m *mockSourceRepo) SoftDeleteByProviderIDs(_ context.Context, accountID uuid.UUID, providerIDs []string, deletedAt time.Time) (int64, error) {
	want := map[string]struct{}{}
	for _, id := range providerIDs {
		want[id] = struct{}{}
	}
	var n int64
	for key, source := range m.byKey {
		if source.ConnectedAccountID != accountID || source.DeletedAt != nil {
			continue
		}
		if _, ok := want[source.ProviderCalendarID]; !ok {
			continue
		}
		source.DeletedAt = &deletedAt
		m.byKey[key] = source
		n++
	}
	m.removed = n
	return n, nil
}

func (m *mockSourceRepo) UpdateEventSyncState(_ context.Context, id uuid.UUID, syncCursor *string, lastSyncedAt, updatedAt time.Time) (entity.CalendarSource, error) {
	for key, source := range m.byKey {
		if source.ID == id {
			source.SyncCursor = syncCursor
			source.LastSyncedAt = &lastSyncedAt
			source.UpdatedAt = updatedAt
			m.byKey[key] = source
			return source, nil
		}
	}
	return entity.CalendarSource{}, apperr.ErrNotFound
}

func (m *mockSourceRepo) ClearEventSyncCursor(_ context.Context, id uuid.UUID, updatedAt time.Time) (entity.CalendarSource, error) {
	for key, source := range m.byKey {
		if source.ID == id {
			source.SyncCursor = nil
			source.UpdatedAt = updatedAt
			m.byKey[key] = source
			return source, nil
		}
	}
	return entity.CalendarSource{}, apperr.ErrNotFound
}

func (m *mockSourceRepo) UpdateSyncEnabledByAccount(_ context.Context, accountID uuid.UUID, syncEnabled bool, updatedAt time.Time) (int64, error) {
	var n int64
	for key, source := range m.byKey {
		if source.ConnectedAccountID != accountID || source.DeletedAt != nil {
			continue
		}
		source.SyncEnabled = syncEnabled
		source.UpdatedAt = updatedAt
		m.byKey[key] = source
		n++
	}
	return n, nil
}

func (m *mockSourceRepo) WithTx(pgx.Tx) repository.CalendarSourceRepository { return m }

type fakeCalendarTx struct{}

func (fakeCalendarTx) WithinTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

func testCalendarLog() *logger.Logger {
	return logger.NewFactory(logger.Options{
		Service:     "test",
		Environment: "test",
		Level:       "error",
		Output:      io.Discard,
	}).Module("calendar")
}

func sealedCred(t *testing.T, key []byte, access, refresh string, expiryUnix int64) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expiry_unix":   expiryUnix,
		"scope":         constant.GoogleScopeCalendar,
	})
	if err != nil {
		t.Fatal(err)
	}
	ct, err := seal.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	return ct
}

func newCalendarServiceForTest(
	t *testing.T,
	accounts *mockCalendarAccountRepo,
	secrets *mockCalendarSecretRepo,
	sources *mockSourceRepo,
	provider *mockCalendarProvider,
	tokens calendarprovider.TokenRefresher,
	key []byte,
) *business.CalendarService {
	t.Helper()
	providerName := constant.AuthProviderGoogle
	if provider != nil && provider.name != "" {
		providerName = provider.name
	}
	if tokens == nil {
		tokens = mockTokenRefresher{}
	}
	return business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts,
		Secrets:  secrets,
		Sources:  sources,
		Events:   &mockEventRepo{byKey: map[string]entity.CalendarEvent{}},
		Tx:       fakeCalendarTx{},
		Providers: map[string]calendarprovider.Provider{
			providerName: provider,
		},
		Tokens: map[string]calendarprovider.TokenRefresher{
			providerName: tokens,
		},
		SealKey: key,
		Log:     testCalendarLog(),
	})
}

func TestSyncSourcesCreatesAndUpdatesIdempotently(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000101")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000102")
	nowUnix := time.Now().Add(time.Hour).Unix()

	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:               accountID,
		UserID:           userID,
		Provider:         constant.AuthProviderGoogle,
		Status:           constant.ConnectedAccountStatusActive,
		Scopes:           []string{constant.GoogleScopeCalendar},
		CredentialsRef:   "cred_test",
		ProviderMetadata: []byte(`{}`),
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref:        "cred_test",
		Ciphertext: sealedCred(t, key, "access-1", "refresh-1", nowUnix),
	}}
	sources := &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}
	api := &mockCalendarProvider{listResult: calendarprovider.ListCalendarsResult{
		NextSyncToken: "sync-token-1",
		Calendars: []calendarprovider.RemoteCalendar{
			{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner", ETag: "etag-1", Color: "#ff0000", TimeZone: "UTC"},
			{ID: "work", Name: "Work", Primary: false, Writable: true, AccessRole: "writer", ETag: "etag-2"},
		},
	}}

	svc := newCalendarServiceForTest(t, accounts, secrets, sources, api, mockTokenRefresher{}, key)

	first, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatalf("first sync: %v", err)
	}
	if first.SourcesCreated != 2 || first.SourcesUpdated != 0 || first.SourcesDeleted != 0 {
		t.Fatalf("first counts = %+v", first)
	}
	if accounts.recorded != 1 {
		t.Fatalf("account record = %d", accounts.recorded)
	}
	if accounts.account.CalendarListSyncToken == nil || *accounts.account.CalendarListSyncToken != "sync-token-1" {
		t.Fatalf("sync token = %#v", accounts.account.CalendarListSyncToken)
	}

	api.listResult = calendarprovider.ListCalendarsResult{
		NextSyncToken: "sync-token-2",
		Incremental:   true,
		Calendars: []calendarprovider.RemoteCalendar{
			{ID: "primary", Name: "Personal Renamed", Primary: true, Writable: true, AccessRole: "owner", ETag: "etag-1b"},
			{ID: "work", Deleted: true},
		},
	}
	svc = newCalendarServiceForTest(t, accounts, secrets, sources, api, mockTokenRefresher{}, key)

	second, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if !second.Incremental {
		t.Fatal("expected incremental second sync")
	}
	if second.SourcesCreated != 0 || second.SourcesUpdated != 1 || second.SourcesDeleted != 1 {
		t.Fatalf("second counts = %+v", second)
	}
	if len(api.listCalls) < 2 || api.listCalls[len(api.listCalls)-1].SyncToken != "sync-token-1" {
		t.Fatalf("expected syncToken on second call, calls=%#v", api.listCalls)
	}

	keyPrimary := accountID.String() + "|primary"
	if sources.byKey[keyPrimary].Name != "Personal Renamed" {
		t.Fatalf("primary name = %q", sources.byKey[keyPrimary].Name)
	}
	keyWork := accountID.String() + "|work"
	if sources.byKey[keyWork].DeletedAt == nil {
		t.Fatal("expected work calendar soft-deleted")
	}
}

func TestSyncSourcesRecoversFromGoneSyncToken(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000501")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000502")
	tok := "stale-token"
	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:                    accountID,
		UserID:                userID,
		Provider:              constant.AuthProviderGoogle,
		Status:                constant.ConnectedAccountStatusActive,
		Scopes:                []string{constant.GoogleScopeCalendar},
		CredentialsRef:        "cred_test",
		CalendarListSyncToken: &tok,
		ProviderMetadata:      []byte(`{}`),
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref:        "cred_test",
		Ciphertext: sealedCred(t, key, "access-1", "refresh-1", time.Now().Add(time.Hour).Unix()),
	}}
	sources := &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}
	api := &mockCalendarProvider{
		gone: true,
		listAfterGone: calendarprovider.ListCalendarsResult{
			NextSyncToken: "fresh-token",
			Calendars: []calendarprovider.RemoteCalendar{
				{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner"},
			},
		},
	}
	svc := newCalendarServiceForTest(t, accounts, secrets, sources, api, mockTokenRefresher{}, key)

	result, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if accounts.cleared != 1 {
		t.Fatalf("cleared = %d", accounts.cleared)
	}
	if result.SourcesCreated != 1 || result.Incremental {
		t.Fatalf("result = %+v", result)
	}
	if accounts.account.CalendarListSyncToken == nil || *accounts.account.CalendarListSyncToken != "fresh-token" {
		t.Fatalf("token = %#v", accounts.account.CalendarListSyncToken)
	}
}

func TestEnsureFreshSkipsWhenRecent(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000601")
	synced := time.Now().UTC().Add(-30 * time.Second)
	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:                 uuid.MustParse("01900000-0000-7000-8000-000000000602"),
		UserID:             userID,
		Provider:           constant.AuthProviderGoogle,
		Status:             constant.ConnectedAccountStatusActive,
		Scopes:             []string{constant.GoogleScopeCalendar},
		CredentialsRef:     "cred_test",
		LastSyncedAt:       &synced,
		CalendarSyncStatus: constant.CalendarSyncStatusSucceeded,
	}}
	sources := &mockSourceRepo{
		byKey: map[string]entity.CalendarSource{},
		byUser: []entity.CalendarSource{{
			ID:                 uuid.MustParse("01900000-0000-7000-8000-000000000603"),
			UserID:             userID,
			ProviderCalendarID: "primary",
			Name:               "Personal",
			ProviderMetadata:   []byte(`{}`),
		}},
	}
	api := &mockCalendarProvider{}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, sources, api, mockTokenRefresher{}, key)

	result, err := svc.EnsureFresh(context.Background(), userID, 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped || len(api.listCalls) != 0 {
		t.Fatalf("result=%+v calls=%d", result, len(api.listCalls))
	}
}

func TestSyncSourcesRequiresCalendarScope(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000201")
	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:             uuid.MustParse("01900000-0000-7000-8000-000000000202"),
		UserID:         userID,
		Provider:       constant.AuthProviderGoogle,
		Status:         constant.ConnectedAccountStatusActive,
		Scopes:         []string{"openid", "email", "profile"},
		CredentialsRef: "cred_test",
	}}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}, &mockCalendarProvider{}, mockTokenRefresher{}, key)

	_, err = svc.SyncSources(context.Background(), userID)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestSyncSourcesMissingAccount(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	accounts := &mockCalendarAccountRepo{err: apperr.ErrNotFound}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}, &mockCalendarProvider{}, mockTokenRefresher{}, key)

	_, err = svc.SyncSources(context.Background(), uuid.MustParse("01900000-0000-7000-8000-000000000301"))
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestListSources(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000401")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000404")
	sources := &mockSourceRepo{
		byKey: map[string]entity.CalendarSource{},
		byUser: []entity.CalendarSource{{
			ID:                 uuid.MustParse("01900000-0000-7000-8000-000000000402"),
			PublicID:           "cal_abc",
			UserID:             userID,
			ProviderCalendarID: "primary",
			Name:               "Personal",
			IsWritable:         true,
			AccessRole:         strPtr("owner"),
			ProviderMetadata:   []byte(`{}`),
		}},
	}
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:       accountID,
		UserID:   userID,
		Provider: constant.AuthProviderGoogle,
		Status:   constant.ConnectedAccountStatusActive,
		Scopes:   []string{constant.GoogleScopeCalendar},
	}}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, sources, &mockCalendarProvider{}, mockTokenRefresher{}, key)

	got, err := svc.ListSources(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 1 || got.Sources[0].Name != "Personal" || !got.Sources[0].IsWritable {
		t.Fatalf("got = %#v", got)
	}
	if len(got.Accounts) != 1 || got.Account.ID != accountID {
		t.Fatalf("accounts = %#v account=%#v", got.Accounts, got.Account)
	}
}

func strPtr(v string) *string { return &v }
