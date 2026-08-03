package business_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/calendarsyncmetrics"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
)

func TestCoordinatorEnsureFreshSkipsWithinCooldown(t *testing.T) {
	key, err := seal.KeyFromSecret("calendar-coord-test-secret-key!!")
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000901")
	synced := time.Now().UTC().Add(-5 * time.Minute)
	accounts := &mockCalendarAccountRepo{account: entity.ConnectedAccount{
		ID:                 uuid.MustParse("01900000-0000-7000-8000-000000000902"),
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
			ID:                 uuid.MustParse("01900000-0000-7000-8000-000000000903"),
			UserID:             userID,
			ProviderCalendarID: "primary",
			Name:               "Personal",
			ProviderMetadata:   []byte(`{}`),
		}},
	}
	api := &mockCalendarProvider{}
	svc := newCalendarServiceForTest(t, accounts, &mockCalendarSecretRepo{}, sources, api, mockTokenRefresher{}, key)
	metrics := calendarsyncmetrics.New()
	log := logger.NewFactory(logger.Options{
		Service:     "api",
		Environment: "test",
		Level:       "info",
		Output:      io.Discard,
	}).Module(constant.ModuleCalendar)
	coord := business.NewCalendarSyncCoordinator(svc, log, metrics)

	result, err := coord.EnsureFresh(context.Background(), userID, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped {
		t.Fatalf("expected skip, got %+v", result)
	}
	if len(api.listCalls) != 0 {
		t.Fatalf("provider should not be called, calls=%d", len(api.listCalls))
	}
	snap := metrics.Snapshot()
	if snap.SyncRequestedTotal != 1 || snap.SyncSkippedTotal != 1 {
		t.Fatalf("metrics = %+v", snap)
	}
}
