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
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockCalendarAPI struct {
	calls  []googlecalendar.ListOptions
	result googlecalendar.ListResult
	err    error
	gone   bool
	after  googlecalendar.ListResult
}

func (m *mockCalendarAPI) ListCalendars(_ context.Context, _ string, opts googlecalendar.ListOptions) (googlecalendar.ListResult, error) {
	m.calls = append(m.calls, opts)
	if m.gone && opts.SyncToken != "" {
		return googlecalendar.ListResult{}, &googlecalendar.GoneError{Body: "sync token expired"}
	}
	if m.err != nil {
		return googlecalendar.ListResult{}, m.err
	}
	if m.gone && opts.SyncToken == "" && len(m.after.Calendars) > 0 {
		return m.after, nil
	}
	return m.result, nil
}

type mockCalendarAccountRepo struct {
	account  entity.ConnectedAccount
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

func (m *mockCalendarAccountRepo) UpdateCredentials(_ context.Context, id uuid.UUID, credentialsRef string, tokenExpiresAt *time.Time, scopes []string, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CredentialsRef = credentialsRef
	m.account.TokenExpiresAt = tokenExpiresAt
	m.account.Scopes = scopes
	m.account.Status = status
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) MarkCalendarSyncRunning(_ context.Context, id uuid.UUID, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.running++
	m.account.ID = id
	m.account.CalendarSyncStatus = status
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) RecordCalendarSync(_ context.Context, id uuid.UUID, record repository.CalendarSyncRecord) (entity.ConnectedAccount, error) {
	m.recorded++
	m.account.ID = id
	m.account.CalendarSyncStatus = record.Status
	m.account.CalendarListSyncToken = record.ListSyncToken
	m.account.LastSyncedAt = record.SuccessfulAt
	m.account.LastFailedSyncAt = record.FailedAt
	ms := record.DurationMs
	m.account.LastSyncDurationMs = &ms
	m.account.LastSyncCreatedCount = record.CreatedCount
	m.account.LastSyncUpdatedCount = record.UpdatedCount
	m.account.LastSyncDeletedCount = record.DeletedCount
	m.account.UpdatedAt = record.UpdatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) ClearCalendarListSyncToken(_ context.Context, id uuid.UUID, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.cleared++
	m.account.ID = id
	m.account.CalendarListSyncToken = nil
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockCalendarAccountRepo) WithTx(pgx.Tx) repository.ConnectedAccountRepository { return m }

type mockCalendarSecretRepo struct {
	secret entity.CredentialSecret
	err    error
}

func (m *mockCalendarSecretRepo) Create(context.Context, entity.CredentialSecret) (entity.CredentialSecret, error) {
	return entity.CredentialSecret{}, errors.New("unexpected Create")
}

func (m *mockCalendarSecretRepo) GetByRef(context.Context, string) (entity.CredentialSecret, error) {
	if m.err != nil {
		return entity.CredentialSecret{}, m.err
	}
	return m.secret, nil
}

func (m *mockCalendarSecretRepo) UpdateCiphertext(_ context.Context, ref string, ciphertext []byte, updatedAt time.Time) (entity.CredentialSecret, error) {
	m.secret.Ref = ref
	m.secret.Ciphertext = ciphertext
	m.secret.UpdatedAt = updatedAt
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
	out := make([]entity.CalendarSource, 0, len(m.byKey))
	for _, source := range m.byKey {
		if source.DeletedAt == nil {
			out = append(out, source)
		}
	}
	if len(m.byUser) > 0 {
		return m.byUser, nil
	}
	return out, nil
}

func (m *mockSourceRepo) ListByConnectedAccountID(context.Context, uuid.UUID) ([]entity.CalendarSource, error) {
	return m.ListByUserID(context.Background(), uuid.Nil)
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

func (m *mockSourceRepo) WithTx(pgx.Tx) repository.CalendarSourceRepository { return m }

type fakeCalendarTx struct{}

func (fakeCalendarTx) WithinTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

type mockCalendarOAuth struct {
	refresh googleoauth.TokenSet
	err     error
}

func (m mockCalendarOAuth) AuthCodeURL(string) string { return "" }
func (m mockCalendarOAuth) ExchangeCode(context.Context, string) (googleoauth.TokenSet, error) {
	return googleoauth.TokenSet{}, errors.New("unused")
}
func (m mockCalendarOAuth) FetchProfile(context.Context, string) (googleoauth.Profile, error) {
	return googleoauth.Profile{}, errors.New("unused")
}
func (m mockCalendarOAuth) RefreshAccessToken(context.Context, string) (googleoauth.TokenSet, error) {
	if m.err != nil {
		return googleoauth.TokenSet{}, m.err
	}
	return m.refresh, nil
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
	api *mockCalendarAPI,
	oauth mockCalendarOAuth,
	key []byte,
) *business.CalendarService {
	t.Helper()
	return business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts,
		Secrets:  secrets,
		Sources:  sources,
		Tx:       fakeCalendarTx{},
		OAuth:    oauth,
		Calendar: api,
		SealKey:  key,
		Log:      testCalendarLog(),
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
		ID:             accountID,
		UserID:         userID,
		Provider:       constant.AuthProviderGoogle,
		Status:         constant.ConnectedAccountStatusActive,
		Scopes:         []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_test",
		ProviderMetadata: []byte(`{}`),
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref:        "cred_test",
		Ciphertext: sealedCred(t, key, "access-1", "refresh-1", nowUnix),
	}}
	sources := &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}
	api := &mockCalendarAPI{result: googlecalendar.ListResult{
		NextSyncToken: "sync-token-1",
		Calendars: []googlecalendar.RemoteCalendar{
			{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner", ETag: "etag-1", Color: "#ff0000", TimeZone: "UTC"},
			{ID: "work", Name: "Work", Primary: false, Writable: true, AccessRole: "writer", ETag: "etag-2"},
		},
	}}

	svc := newCalendarServiceForTest(t, accounts, secrets, sources, api, mockCalendarOAuth{}, key)

	first, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatalf("first sync: %v", err)
	}
	if first.CreatedCount != 2 || first.UpdatedCount != 0 || first.RemovedCount != 0 {
		t.Fatalf("first counts = %+v", first)
	}
	if accounts.recorded != 1 {
		t.Fatalf("account record = %d", accounts.recorded)
	}
	if accounts.account.CalendarListSyncToken == nil || *accounts.account.CalendarListSyncToken != "sync-token-1" {
		t.Fatalf("sync token = %#v", accounts.account.CalendarListSyncToken)
	}

	// Incremental second sync: rename + explicit delete.
	api.result = googlecalendar.ListResult{
		NextSyncToken: "sync-token-2",
		Incremental:   true,
		Calendars: []googlecalendar.RemoteCalendar{
			{ID: "primary", Name: "Personal Renamed", Primary: true, Writable: true, AccessRole: "owner", ETag: "etag-1b"},
			{ID: "work", Deleted: true},
		},
	}
	svc = newCalendarServiceForTest(t, accounts, secrets, sources, api, mockCalendarOAuth{}, key)

	second, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if !second.Incremental {
		t.Fatal("expected incremental second sync")
	}
	if second.CreatedCount != 0 || second.UpdatedCount != 1 || second.RemovedCount != 1 {
		t.Fatalf("second counts = %+v", second)
	}
	if len(api.calls) < 2 || api.calls[len(api.calls)-1].SyncToken != "sync-token-1" {
		t.Fatalf("expected syncToken on second call, calls=%#v", api.calls)
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
	api := &mockCalendarAPI{
		gone: true,
		after: googlecalendar.ListResult{
			NextSyncToken: "fresh-token",
			Calendars: []googlecalendar.RemoteCalendar{
				{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner"},
			},
		},
	}
	svc := newCalendarServiceForTest(t, accounts, secrets, sources, api, mockCalendarOAuth{}, key)

	result, err := svc.SyncSources(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if accounts.cleared != 1 {
		t.Fatalf("cleared = %d", accounts.cleared)
	}
	if result.CreatedCount != 1 || result.Incremental {
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
	api := &mockCalendarAPI{}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, sources, api, mockCalendarOAuth{}, key)

	result, err := svc.EnsureFresh(context.Background(), userID, 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped || len(api.calls) != 0 {
		t.Fatalf("result=%+v calls=%d", result, len(api.calls))
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
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}, &mockCalendarAPI{}, mockCalendarOAuth{}, key)

	_, err = svc.SyncSources(context.Background(), userID)
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestSyncSourcesMissingAccount(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-test-secret-key-32b!!")
	if err != nil {
		t.Fatal(err)
	}
	accounts := &mockCalendarAccountRepo{err: apperr.ErrNotFound}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}, &mockCalendarAPI{}, mockCalendarOAuth{}, key)

	_, err = svc.SyncSources(context.Background(), uuid.MustParse("01900000-0000-7000-8000-000000000301"))
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestListSources(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000401")
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
	svc := newCalendarServiceForTest(t, &mockCalendarAccountRepo{}, &mockCalendarSecretRepo{}, sources, &mockCalendarAPI{}, mockCalendarOAuth{}, key)

	got, err := svc.ListSources(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 1 || got.Sources[0].Name != "Personal" || !got.Sources[0].IsWritable {
		t.Fatalf("got = %#v", got)
	}
}

func strPtr(v string) *string { return &v }
