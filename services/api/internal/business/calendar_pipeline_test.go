package business_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockSyncRunRepo struct {
	runs []entity.CalendarSyncRun
}

func (m *mockSyncRunRepo) Create(_ context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error) {
	m.runs = append(m.runs, run)
	return run, nil
}
func (m *mockSyncRunRepo) Finish(_ context.Context, run entity.CalendarSyncRun) (entity.CalendarSyncRun, error) {
	for i := range m.runs {
		if m.runs[i].ID == run.ID {
			m.runs[i] = run
			return run, nil
		}
	}
	m.runs = append(m.runs, run)
	return run, nil
}
func (m *mockSyncRunRepo) WithTx(pgx.Tx) repository.CalendarSyncRunRepository { return m }

type flakyEventsProvider struct {
	name    string
	failCal string
	listed  calendarprovider.ListCalendarsResult
	events  map[string]calendarprovider.ListEventsResult
}

func (m *flakyEventsProvider) Name() string {
	if m.name != "" {
		return m.name
	}
	return constant.AuthProviderGoogle
}
func (m *flakyEventsProvider) ListCalendars(context.Context, string, calendarprovider.ListCalendarsOptions) (calendarprovider.ListCalendarsResult, error) {
	return m.listed, nil
}
func (m *flakyEventsProvider) ListEvents(_ context.Context, _ string, calendarID string, _ calendarprovider.ListEventsOptions) (calendarprovider.ListEventsResult, error) {
	if calendarID == m.failCal {
		return calendarprovider.ListEventsResult{}, errors.New("calendar events boom")
	}
	if m.events != nil {
		if r, ok := m.events[calendarID]; ok {
			return r, nil
		}
	}
	return calendarprovider.ListEventsResult{NextSyncToken: "ok"}, nil
}

func TestSyncPipelineContinuesWhenOneCalendarFails(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-pipeline-test-secret-key!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000901")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000902")
	nowUnix := time.Now().Add(time.Hour).Unix()

	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID: accountID, UserID: userID, Provider: constant.AuthProviderGoogle,
		Status: constant.ConnectedAccountStatusActive, Scopes: []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_pipe", ProviderMetadata: []byte(`{}`),
	}}
	secrets := &mockCalendarSecretRepo{secret: entity.CredentialSecret{
		Ref: "cred_pipe", Ciphertext: sealedCred(t, key, "access", "refresh", nowUnix),
	}}
	sources := &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}
	runs := &mockSyncRunRepo{}
	start := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	api := &flakyEventsProvider{
		failCal: "work",
		listed: calendarprovider.ListCalendarsResult{
			NextSyncToken: "list-1",
			Calendars: []calendarprovider.RemoteCalendar{
				{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner"},
				{ID: "work", Name: "Work", Writable: true, AccessRole: "writer"},
			},
		},
		events: map[string]calendarprovider.ListEventsResult{
			"primary": {
				NextSyncToken: "e1",
				Events: []calendarprovider.RemoteEvent{{
					ID: "evt-1", Title: "Ok", StartsAt: start, EndsAt: start.Add(time.Hour),
					Status: "confirmed", Raw: map[string]any{"id": "evt-1"},
				}},
			},
		},
	}

	svc := business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts,
		Secrets:  secrets,
		Sources:  sources,
		Events:   &mockEventRepo{byKey: map[string]entity.CalendarEvent{}},
		SyncRuns: runs,
		Tx:       fakeCalendarTx{},
		Providers: map[string]calendarprovider.Provider{
			constant.AuthProviderGoogle: api,
		},
		Tokens: map[string]calendarprovider.TokenRefresher{
			constant.AuthProviderGoogle: mockTokenRefresher{},
		},
		SealKey: key,
		Log:     testCalendarLog(),
	})

	result, err := svc.SyncPipeline(context.Background(), userID, constant.CalendarSyncTriggerManual)
	if err != nil {
		t.Fatalf("pipeline err: %v", err)
	}
	if result.Status != constant.CalendarSyncRunStatusPartial {
		t.Fatalf("status = %s", result.Status)
	}
	if result.CalendarsProcessed != 2 {
		t.Fatalf("calendars_processed = %d", result.CalendarsProcessed)
	}
	if result.EventsCreated != 1 {
		t.Fatalf("events_created = %d", result.EventsCreated)
	}
	if len(result.Failures) != 1 || result.Failures[0].ProviderCalendarID != "work" {
		t.Fatalf("failures = %#v", result.Failures)
	}
	if len(runs.runs) == 0 || runs.runs[len(runs.runs)-1].Status != constant.CalendarSyncRunStatusPartial {
		t.Fatalf("persisted runs = %#v", runs.runs)
	}
}

func TestSyncPipelineAggregatesTwoProviders(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-pipeline-test-secret-key!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000911")
	googleID := uuid.MustParse("01900000-0000-7000-8000-000000000912")
	msID := uuid.MustParse("01900000-0000-7000-8000-000000000913")
	nowUnix := time.Now().Add(time.Hour).Unix()

	googleAccount := entity.ConnectedAccount{
		ID: googleID, UserID: userID, Provider: constant.AuthProviderGoogle,
		Status: constant.ConnectedAccountStatusActive, Scopes: []string{constant.GoogleScopeCalendar},
		CredentialsRef: "cred_google", ProviderMetadata: []byte(`{}`),
	}
	msAccount := entity.ConnectedAccount{
		ID: msID, UserID: userID, Provider: constant.AuthProviderMicrosoft,
		Status: constant.ConnectedAccountStatusActive,
		Scopes: []string{constant.MicrosoftScopeCalendarsReadWrite, constant.MicrosoftScopeOfflineAccess},
		CredentialsRef: "cred_ms", ProviderMetadata: []byte(`{}`),
	}
	accounts := &mockCalendarAccountRepo{
		account:  googleAccount,
		accounts: []entity.ConnectedAccount{googleAccount, msAccount},
	}
	secrets := &mockCalendarSecretRepo{
		secrets: map[string]entity.CredentialSecret{
			"cred_google": {Ref: "cred_google", Ciphertext: sealedCred(t, key, "g-access", "g-refresh", nowUnix)},
			"cred_ms":     {Ref: "cred_ms", Ciphertext: sealedCred(t, key, "m-access", "m-refresh", nowUnix)},
		},
	}
	sources := &mockSourceRepo{byKey: map[string]entity.CalendarSource{}}
	googleAPI := &mockCalendarProvider{
		name: constant.AuthProviderGoogle,
		listResult: calendarprovider.ListCalendarsResult{
			NextSyncToken: "g-list",
			Calendars: []calendarprovider.RemoteCalendar{
				{ID: "g-primary", Name: "Google", Primary: true, Writable: true, AccessRole: "owner"},
			},
		},
		eventsByCal: map[string]calendarprovider.ListEventsResult{
			"g-primary": {NextSyncToken: "g-evt"},
		},
	}
	msAPI := &mockCalendarProvider{
		name: constant.AuthProviderMicrosoft,
		listResult: calendarprovider.ListCalendarsResult{
			NextSyncToken: "m-list",
			Calendars: []calendarprovider.RemoteCalendar{
				{ID: "m-primary", Name: "Outlook", Primary: true, Writable: true, AccessRole: "owner"},
			},
		},
		eventsByCal: map[string]calendarprovider.ListEventsResult{
			"m-primary": {NextSyncToken: "m-evt"},
		},
	}

	svc := business.NewCalendarService(business.CalendarServiceDeps{
		Accounts: accounts,
		Secrets:  secrets,
		Sources:  sources,
		Events:   &mockEventRepo{byKey: map[string]entity.CalendarEvent{}},
		Tx:       fakeCalendarTx{},
		Providers: map[string]calendarprovider.Provider{
			constant.AuthProviderGoogle:    googleAPI,
			constant.AuthProviderMicrosoft: msAPI,
		},
		Tokens: map[string]calendarprovider.TokenRefresher{
			constant.AuthProviderGoogle:    mockTokenRefresher{},
			constant.AuthProviderMicrosoft: mockTokenRefresher{},
		},
		SealKey: key,
		Log:     testCalendarLog(),
	})

	result, err := svc.SyncPipeline(context.Background(), userID, constant.CalendarSyncTriggerManual)
	if err != nil {
		t.Fatalf("pipeline err: %v", err)
	}
	if result.Status != constant.CalendarSyncRunStatusSucceeded {
		t.Fatalf("status = %s failures=%#v", result.Status, result.Failures)
	}
	if result.SourcesCreated != 2 {
		t.Fatalf("sources_created = %d", result.SourcesCreated)
	}
	if result.CalendarsProcessed != 2 {
		t.Fatalf("calendars_processed = %d", result.CalendarsProcessed)
	}
	if len(googleAPI.listCalls) != 1 || len(msAPI.listCalls) != 1 {
		t.Fatalf("provider list calls google=%d microsoft=%d", len(googleAPI.listCalls), len(msAPI.listCalls))
	}
	if len(result.Sources) != 2 {
		t.Fatalf("combined sources = %d", len(result.Sources))
	}
}
