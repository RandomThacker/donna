package business_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
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

type flakyEventsAPI struct {
	failCal string
	listed  googlecalendar.ListResult
	events  map[string]googlecalendar.EventListResult
}

func (m *flakyEventsAPI) ListCalendars(context.Context, string, googlecalendar.ListOptions) (googlecalendar.ListResult, error) {
	return m.listed, nil
}
func (m *flakyEventsAPI) ListEvents(_ context.Context, _ string, calendarID string, _ googlecalendar.EventListOptions) (googlecalendar.EventListResult, error) {
	if calendarID == m.failCal {
		return googlecalendar.EventListResult{}, errors.New("google events boom")
	}
	if m.events != nil {
		if r, ok := m.events[calendarID]; ok {
			return r, nil
		}
	}
	return googlecalendar.EventListResult{NextSyncToken: "ok"}, nil
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
	api := &flakyEventsAPI{
		failCal: "work",
		listed: googlecalendar.ListResult{
			NextSyncToken: "list-1",
			Calendars: []googlecalendar.RemoteCalendar{
				{ID: "primary", Name: "Personal", Primary: true, Writable: true, AccessRole: "owner"},
				{ID: "work", Name: "Work", Writable: true, AccessRole: "writer"},
			},
		},
		events: map[string]googlecalendar.EventListResult{
			"primary": {
				NextSyncToken: "e1",
				Events: []googlecalendar.RemoteEvent{{
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
		OAuth:    mockCalendarOAuth{},
		Calendar: api,
		SealKey:  key,
		Log:      testCalendarLog(),
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
