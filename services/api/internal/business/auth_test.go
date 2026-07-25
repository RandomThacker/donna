package business_test

import (
	"context"
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
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockGoogle struct {
	tokenSet googleoauth.TokenSet
	profile  googleoauth.Profile
	exchErr  error
	profErr  error
}

func (m mockGoogle) AuthCodeURL(state string) string {
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
}

func (m mockGoogle) ExchangeCode(context.Context, string) (googleoauth.TokenSet, error) {
	if m.exchErr != nil {
		return googleoauth.TokenSet{}, m.exchErr
	}
	return m.tokenSet, nil
}

func (m mockGoogle) FetchProfile(context.Context, string) (googleoauth.Profile, error) {
	if m.profErr != nil {
		return googleoauth.Profile{}, m.profErr
	}
	return m.profile, nil
}

func (m mockGoogle) RefreshAccessToken(context.Context, string) (googleoauth.TokenSet, error) {
	return m.tokenSet, nil
}

type mockIdentityRepo struct {
	identity entity.AuthIdentity
	err      error
	created  int
}

func (m *mockIdentityRepo) Create(_ context.Context, identity entity.AuthIdentity) (entity.AuthIdentity, error) {
	m.created++
	return identity, nil
}

func (m *mockIdentityRepo) GetByProviderSubject(context.Context, string, string) (entity.AuthIdentity, error) {
	if m.err != nil {
		return entity.AuthIdentity{}, m.err
	}
	return m.identity, nil
}

func (m *mockIdentityRepo) WithTx(pgx.Tx) repository.AuthIdentityRepository { return m }

type mockAccountRepo struct {
	account entity.ConnectedAccount
	err     error
	created int
}

func (m *mockAccountRepo) Create(_ context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, error) {
	m.created++
	m.account = account
	return account, nil
}

func (m *mockAccountRepo) GetByID(context.Context, uuid.UUID) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
}

func (m *mockAccountRepo) GetByProviderAccount(context.Context, string, string) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
}

func (m *mockAccountRepo) GetByUserAndProvider(context.Context, uuid.UUID, string) (entity.ConnectedAccount, error) {
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return m.account, nil
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

type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

func newTestAuth(
	t *testing.T,
	google mockGoogle,
	identities *mockIdentityRepo,
	accounts *mockAccountRepo,
	secrets *mockSecretRepo,
	users mockUserRepo,
	tx business.TxRunner,
) *business.AuthService {
	t.Helper()
	key, err := seal.KeyFromSecret("test-credentials-key")
	if err != nil {
		t.Fatal(err)
	}
	iss, err := session.NewIssuer("test-jwt-secret-value", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	log := logger.NewFactory(logger.Options{
		Environment: constant.EnvDevelopment,
		Level:       "error",
		Output:      io.Discard,
	}).Module(constant.ModuleAuth)

	return business.NewAuthService(business.AuthServiceDeps{
		Users:      business.NewUserService(users, log),
		UserRepo:   users,
		Identities: identities,
		Accounts:   accounts,
		Secrets:    secrets,
		Tx:         tx,
		Google:     google,
		State:      oauthstate.NewManager("test-jwt-secret-value"),
		Tokens:     iss,
		SealKey:    key,
		Log:        log,
	})
}

func validState(t *testing.T) string {
	t.Helper()
	state, err := oauthstate.NewManager("test-jwt-secret-value").Create()
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestBeginGoogleLogin(t *testing.T) {
	svc := newTestAuth(t, mockGoogle{}, &mockIdentityRepo{}, &mockAccountRepo{}, &mockSecretRepo{}, mockUserRepo{}, fakeTx{})
	url, state, err := svc.BeginGoogleLogin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if url == "" || state == "" {
		t.Fatalf("url=%q state=%q", url, state)
	}
}

func TestCompleteGoogleLoginExistingUser(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000010")
	users := mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (entity.User, error) {
			return entity.User{ID: id, PublicID: "usr_x", Email: "a@b.com"}, nil
		},
	}
	identities := &mockIdentityRepo{
		identity: entity.AuthIdentity{UserID: userID, Provider: constant.AuthProviderGoogle, ProviderSubject: "sub-1"},
	}
	accounts := &mockAccountRepo{err: apperr.ErrNotFound}
	secrets := &mockSecretRepo{}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600, Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "sub-1", Email: "a@b.com", EmailVerified: true, Name: "Ada"},
	}
	svc := newTestAuth(t, google, identities, accounts, secrets, users, fakeTx{})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if session.AccessToken == "" || session.IsNewUser {
		t.Fatalf("session = %#v", session)
	}
	if accounts.created != 1 || secrets.created != 1 {
		t.Fatalf("accounts=%d secrets=%d", accounts.created, secrets.created)
	}
}

func TestCompleteGoogleLoginNewUser(t *testing.T) {
	users := mockUserRepo{
		createFn: func(_ context.Context, user entity.User) (entity.User, error) {
			return user, nil
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	accounts := &mockAccountRepo{err: apperr.ErrNotFound}
	secrets := &mockSecretRepo{}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600, Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "sub-new", Email: "new@example.com", EmailVerified: true, Name: "New"},
	}
	svc := newTestAuth(t, google, identities, accounts, secrets, users, fakeTx{})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if !session.IsNewUser || session.User.Email != "new@example.com" {
		t.Fatalf("session = %#v", session)
	}
	if identities.created != 1 || accounts.created != 1 || secrets.created != 1 {
		t.Fatalf("identity=%d account=%d secret=%d", identities.created, accounts.created, secrets.created)
	}
	if !session.User.EmailVerified {
		t.Fatal("expected email_verified from Google")
	}
}

func TestCompleteGoogleLoginRejectsBadState(t *testing.T) {
	svc := newTestAuth(t, mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", RefreshToken: "rt"},
		profile:  googleoauth.Profile{Subject: "sub"},
	}, &mockIdentityRepo{err: apperr.ErrNotFound}, &mockAccountRepo{}, &mockSecretRepo{}, mockUserRepo{}, fakeTx{})

	_, err := svc.CompleteGoogleLogin(context.Background(), "code", "not-a-valid-state")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
