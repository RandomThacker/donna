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
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockAccountRepo struct {
	account  entity.ConnectedAccount
	accounts []entity.ConnectedAccount
	err      error
	created  int
	deleted  int
	lastSoft entity.ConnectedAccount
}

func (m *mockAccountRepo) Create(_ context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, error) {
	m.created++
	m.account = account
	m.accounts = append(m.accounts, account)
	return account, nil
}

func (m *mockAccountRepo) GetByID(_ context.Context, id uuid.UUID) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	if m.account.ID == id {
		return m.account, nil
	}
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return entity.ConnectedAccount{}, apperr.ErrNotFound
}

func (m *mockAccountRepo) GetByProviderAccount(context.Context, string, string) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	if m.account.ID == uuid.Nil {
		return entity.ConnectedAccount{}, apperr.ErrNotFound
	}
	return m.account, nil
}

func (m *mockAccountRepo) GetByUserAndProvider(context.Context, uuid.UUID, string) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
}

func (m *mockAccountRepo) ListByUserID(context.Context, uuid.UUID) ([]entity.ConnectedAccount, error) {
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

func (m *mockAccountRepo) UpdateCredentials(_ context.Context, id uuid.UUID, credentialsRef string, tokenExpiresAt *time.Time, scopes []string, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CredentialsRef = credentialsRef
	m.account.TokenExpiresAt = tokenExpiresAt
	m.account.Scopes = scopes
	m.account.Status = status
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockAccountRepo) UpdateProfile(_ context.Context, id uuid.UUID, displayName *string, providerMetadata []byte, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.DisplayName = displayName
	m.account.ProviderMetadata = providerMetadata
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockAccountRepo) MarkCalendarSyncRunning(_ context.Context, id uuid.UUID, status string, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CalendarSyncStatus = status
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockAccountRepo) RecordCalendarSync(_ context.Context, id uuid.UUID, record repository.CalendarSyncRecord) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CalendarSyncStatus = record.Status
	m.account.CalendarListSyncToken = record.ListSyncToken
	m.account.LastSyncedAt = record.SuccessfulAt
	m.account.LastFailedSyncAt = record.FailedAt
	m.account.UpdatedAt = record.UpdatedAt
	return m.account, nil
}

func (m *mockAccountRepo) ClearCalendarListSyncToken(_ context.Context, id uuid.UUID, updatedAt time.Time) (entity.ConnectedAccount, error) {
	m.account.ID = id
	m.account.CalendarListSyncToken = nil
	m.account.UpdatedAt = updatedAt
	return m.account, nil
}

func (m *mockAccountRepo) SoftDelete(_ context.Context, id uuid.UUID, deletedAt time.Time) (entity.ConnectedAccount, error) {
	m.deleted++
	m.account.ID = id
	m.account.Status = constant.ConnectedAccountStatusDisconnected
	m.account.DeletedAt = &deletedAt
	m.account.UpdatedAt = deletedAt
	m.lastSoft = m.account
	return m.account, nil
}

func (m *mockAccountRepo) WithTx(pgx.Tx) repository.ConnectedAccountRepository { return m }

type mockSecretRepo struct {
	byRef   map[string]entity.CredentialSecret
	created int
}

func (m *mockSecretRepo) Create(_ context.Context, secret entity.CredentialSecret) (entity.CredentialSecret, error) {
	if m.byRef == nil {
		m.byRef = map[string]entity.CredentialSecret{}
	}
	m.byRef[secret.Ref] = secret
	m.created++
	return secret, nil
}

func (m *mockSecretRepo) GetByRef(_ context.Context, ref string) (entity.CredentialSecret, error) {
	if s, ok := m.byRef[ref]; ok {
		return s, nil
	}
	return entity.CredentialSecret{}, apperr.ErrNotFound
}

func (m *mockSecretRepo) UpdateCiphertext(_ context.Context, ref string, ciphertext []byte, updatedAt time.Time) (entity.CredentialSecret, error) {
	s, ok := m.byRef[ref]
	if !ok {
		return entity.CredentialSecret{}, apperr.ErrNotFound
	}
	s.Ciphertext = ciphertext
	s.UpdatedAt = updatedAt
	m.byRef[ref] = s
	return s, nil
}

func (m *mockSecretRepo) WithTx(pgx.Tx) repository.CredentialSecretRepository { return m }

func newTestIntegration(
	t *testing.T,
	google business.GoogleOAuthClient,
	microsoft business.MicrosoftOAuthClient,
	accounts *mockAccountRepo,
	secrets *mockSecretRepo,
) *business.IntegrationService {
	t.Helper()
	key, err := seal.KeyFromSecret("test-credentials-key")
	if err != nil {
		t.Fatal(err)
	}
	log := logger.NewFactory(logger.Options{
		Environment: constant.EnvDevelopment,
		Level:       "error",
		Output:      io.Discard,
	}).Module(constant.ModuleCalendar)

	return business.NewIntegrationService(business.IntegrationServiceDeps{
		Accounts:  accounts,
		Secrets:   secrets,
		Tx:        fakeTx{},
		Google:    google,
		Microsoft: microsoft,
		State:     oauthstate.NewManager("test-jwt-secret-value"),
		SealKey:   key,
		Log:       log,
	})
}

func userBoundState(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	state, err := oauthstate.NewManager("test-jwt-secret-value").CreateWithUser(userID.String())
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestListConnectedAccounts(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000100")
	accounts := &mockAccountRepo{
		accounts: []entity.ConnectedAccount{
			{ID: uuid.MustParse("01900000-0000-7000-8000-000000000101"), UserID: userID, Provider: constant.AuthProviderGoogle, ProviderMetadata: []byte(`{"email":"a@example.com"}`)},
			{ID: uuid.MustParse("01900000-0000-7000-8000-000000000102"), UserID: userID, Provider: constant.AuthProviderMicrosoft, ProviderMetadata: []byte(`{"email":"b@example.com"}`)},
		},
	}
	svc := newTestIntegration(t, nil, nil, accounts, &mockSecretRepo{})
	got, err := svc.ListConnectedAccounts(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d accounts", len(got))
	}
}

func TestListConnectedAccountsBackfillsEmail(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000160")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000161")
	name := "Aryan Thacker"
	key, err := seal.KeyFromSecret("test-credentials-key")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"access_token":  "access",
		"refresh_token": "refresh",
		"expiry_unix":   time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := seal.Encrypt(key, payload)
	if err != nil {
		t.Fatal(err)
	}
	secrets := &mockSecretRepo{byRef: map[string]entity.CredentialSecret{
		"cred_1": {Ref: "cred_1", Ciphertext: ciphertext},
	}}
	accounts := &mockAccountRepo{
		accounts: []entity.ConnectedAccount{{
			ID:               accountID,
			UserID:           userID,
			Provider:         constant.AuthProviderGoogle,
			DisplayName:      &name,
			CredentialsRef:   "cred_1",
			ProviderMetadata: []byte("{}"),
		}},
		account: entity.ConnectedAccount{
			ID:               accountID,
			UserID:           userID,
			Provider:         constant.AuthProviderGoogle,
			DisplayName:      &name,
			CredentialsRef:   "cred_1",
			ProviderMetadata: []byte("{}"),
		},
	}
	google := mockGoogle{
		profile: googleoauth.Profile{
			Subject: "sub",
			Email:   "aryan@example.com",
			Name:    "Aryan Thacker",
			Picture: "https://lh3.googleusercontent.com/a/example",
		},
	}
	svc := newTestIntegration(t, google, nil, accounts, secrets)

	got, err := svc.ListConnectedAccounts(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d", len(got))
	}
	var meta struct {
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if json.Unmarshal(got[0].ProviderMetadata, &meta) != nil || meta.Email != "aryan@example.com" {
		t.Fatalf("metadata = %s", got[0].ProviderMetadata)
	}
	if meta.AvatarURL != "https://lh3.googleusercontent.com/a/example" {
		t.Fatalf("avatar_url = %q", meta.AvatarURL)
	}
}

func TestCompleteGoogleConnectCreatesAccount(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000110")
	accounts := &mockAccountRepo{err: apperr.ErrNotFound}
	secrets := &mockSecretRepo{}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresIn:    3600,
			Scope:        "openid email profile https://www.googleapis.com/auth/calendar",
		},
		profile: googleoauth.Profile{Subject: "g-sub", Email: "g@example.com", EmailVerified: true, Name: "G"},
	}
	svc := newTestIntegration(t, google, nil, accounts, secrets)

	var ready uuid.UUID
	svc.SetOnAccountReady(func(_ context.Context, accountID uuid.UUID) error {
		ready = accountID
		return nil
	})

	account, err := svc.CompleteGoogleConnect(context.Background(), "code", userBoundState(t, userID))
	if err != nil {
		t.Fatalf("CompleteGoogleConnect: %v", err)
	}
	if accounts.created != 1 || secrets.created != 1 {
		t.Fatalf("accounts=%d secrets=%d", accounts.created, secrets.created)
	}
	if account.UserID != userID || account.Provider != constant.AuthProviderGoogle {
		t.Fatalf("account = %#v", account)
	}
	if ready != account.ID {
		t.Fatalf("ready hook not called for %s", account.ID)
	}
}

func TestCompleteMicrosoftConnectCreatesAccount(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000120")
	accounts := &mockAccountRepo{err: apperr.ErrNotFound}
	secrets := &mockSecretRepo{}
	ms := mockMicrosoft{
		tokenSet: microsoftoauth.TokenSet{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresIn:    3600,
			Scope:        "openid email profile offline_access User.Read Calendars.ReadWrite",
		},
		profile: microsoftoauth.Profile{Subject: "ms-sub", Email: "m@example.com", EmailVerified: true, Name: "M"},
	}
	svc := newTestIntegration(t, nil, ms, accounts, secrets)

	account, err := svc.CompleteMicrosoftConnect(context.Background(), "code", userBoundState(t, userID))
	if err != nil {
		t.Fatalf("CompleteMicrosoftConnect: %v", err)
	}
	if accounts.created != 1 || account.Provider != constant.AuthProviderMicrosoft {
		t.Fatalf("account=%#v created=%d", account, accounts.created)
	}
}

func TestDisconnectSoftDeletesOwnedAccount(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000130")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000131")
	accounts := &mockAccountRepo{
		account: entity.ConnectedAccount{
			ID:       accountID,
			UserID:   userID,
			Provider: constant.AuthProviderGoogle,
			Status:   constant.ConnectedAccountStatusActive,
		},
	}
	svc := newTestIntegration(t, nil, nil, accounts, &mockSecretRepo{})

	if err := svc.Disconnect(context.Background(), userID, accountID); err != nil {
		t.Fatal(err)
	}
	if accounts.deleted != 1 || accounts.lastSoft.Status != constant.ConnectedAccountStatusDisconnected {
		t.Fatalf("soft delete = %#v deleted=%d", accounts.lastSoft, accounts.deleted)
	}
}

func TestCompleteGoogleConnectKeepsCallerUser(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000150")
	accounts := &mockAccountRepo{err: apperr.ErrNotFound}
	secrets := &mockSecretRepo{}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresIn:    3600,
			Scope:        "openid email profile https://www.googleapis.com/auth/calendar",
		},
		profile: googleoauth.Profile{Subject: "g-other", Email: "other@example.com", EmailVerified: true, Name: "Other"},
	}
	svc := newTestIntegration(t, google, nil, accounts, secrets)

	account, err := svc.CompleteGoogleConnect(context.Background(), "code", userBoundState(t, userID))
	if err != nil {
		t.Fatalf("CompleteGoogleConnect: %v", err)
	}
	if account.UserID != userID {
		t.Fatalf("integration must bind to session user %s, got %s", userID, account.UserID)
	}
}

func TestDisconnectRejectsOtherUsersAccount(t *testing.T) {
	owner := uuid.MustParse("01900000-0000-7000-8000-000000000140")
	other := uuid.MustParse("01900000-0000-7000-8000-000000000141")
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000142")
	accounts := &mockAccountRepo{
		account: entity.ConnectedAccount{ID: accountID, UserID: owner, Provider: constant.AuthProviderGoogle},
	}
	svc := newTestIntegration(t, nil, nil, accounts, &mockSecretRepo{})

	err := svc.Disconnect(context.Background(), other, accountID)
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}
