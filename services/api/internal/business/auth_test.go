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
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
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

type mockMicrosoft struct {
	tokenSet microsoftoauth.TokenSet
	profile  microsoftoauth.Profile
	exchErr  error
}

func (m mockMicrosoft) AuthCodeURL(state string) string {
	return "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?state=" + state
}

func (m mockMicrosoft) Exchange(context.Context, string) (microsoftoauth.TokenSet, microsoftoauth.Profile, error) {
	if m.exchErr != nil {
		return microsoftoauth.TokenSet{}, microsoftoauth.Profile{}, m.exchErr
	}
	return m.tokenSet, m.profile, nil
}

func (m mockMicrosoft) RefreshAccessToken(context.Context, string) (microsoftoauth.TokenSet, error) {
	return m.tokenSet, nil
}

func (m mockMicrosoft) FetchProfile(context.Context, string) (microsoftoauth.Profile, error) {
	return m.profile, nil
}

type mockLoginLinker struct {
	googleCalls     int
	microsoftCalls  int
	lastGoogleUser  uuid.UUID
	lastGoogleSub   string
	lastMSUser      uuid.UUID
	lastMSSub       string
	err             error
}

func (m *mockLoginLinker) LinkGoogleFromLogin(
	_ context.Context,
	userID uuid.UUID,
	profile googleoauth.Profile,
	_ googleoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	m.googleCalls++
	m.lastGoogleUser = userID
	m.lastGoogleSub = profile.Subject
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return entity.ConnectedAccount{UserID: userID, Provider: constant.AuthProviderGoogle, ProviderAccountID: profile.Subject}, nil
}

func (m *mockLoginLinker) LinkMicrosoftFromLogin(
	_ context.Context,
	userID uuid.UUID,
	profile microsoftoauth.Profile,
	_ microsoftoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	m.microsoftCalls++
	m.lastMSUser = userID
	m.lastMSSub = profile.Subject
	if m.err != nil {
		return entity.ConnectedAccount{}, m.err
	}
	return entity.ConnectedAccount{UserID: userID, Provider: constant.AuthProviderMicrosoft, ProviderAccountID: profile.Subject}, nil
}

type mockIdentityRepo struct {
	identity entity.AuthIdentity
	err      error
	created  int
	last     entity.AuthIdentity
}

func (m *mockIdentityRepo) Create(_ context.Context, identity entity.AuthIdentity) (entity.AuthIdentity, error) {
	m.created++
	m.last = identity
	return identity, nil
}

func (m *mockIdentityRepo) GetByProviderSubject(context.Context, string, string) (entity.AuthIdentity, error) {
	if m.err != nil {
		return entity.AuthIdentity{}, m.err
	}
	return m.identity, nil
}

func (m *mockIdentityRepo) WithTx(pgx.Tx) repository.AuthIdentityRepository { return m }

type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	return fn(ctx, nil)
}

func newTestAuth(
	t *testing.T,
	google mockGoogle,
	microsoft business.MicrosoftOAuthClient,
	identities *mockIdentityRepo,
	users mockUserRepo,
	tx business.TxRunner,
) *business.AuthService {
	t.Helper()
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
		Tx:         tx,
		Google:     google,
		Microsoft:  microsoft,
		State:      oauthstate.NewManager("test-jwt-secret-value"),
		Tokens:     iss,
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
	svc := newTestAuth(t, mockGoogle{}, nil, &mockIdentityRepo{}, mockUserRepo{}, fakeTx{})
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
			return entity.User{ID: id, PublicID: "usr_x", Email: "a@b.com", EmailVerified: true}, nil
		},
	}
	identities := &mockIdentityRepo{
		identity: entity.AuthIdentity{UserID: userID, Provider: constant.AuthProviderGoogle, ProviderSubject: "sub-1"},
	}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", ExpiresIn: 3600, Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "sub-1", Email: "a@b.com", EmailVerified: true, Name: "Ada"},
	}
	svc := newTestAuth(t, google, nil, identities, users, fakeTx{})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if session.AccessToken == "" || session.IsNewUser {
		t.Fatalf("session = %#v", session)
	}
	if identities.created != 0 {
		t.Fatalf("login must not create identities; created=%d", identities.created)
	}
}

func TestCompleteGoogleLoginNewUser(t *testing.T) {
	users := mockUserRepo{
		createFn: func(_ context.Context, user entity.User) (entity.User, error) {
			return user, nil
		},
		getByEmailFn: func(context.Context, string) (entity.User, error) {
			return entity.User{}, apperr.ErrNotFound
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", ExpiresIn: 3600, Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "sub-new", Email: "new@example.com", EmailVerified: true, Name: "New"},
	}
	svc := newTestAuth(t, google, nil, identities, users, fakeTx{})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if !session.IsNewUser || session.User.Email != "new@example.com" {
		t.Fatalf("session = %#v", session)
	}
	if identities.created != 1 {
		t.Fatalf("identity=%d", identities.created)
	}
	if !session.User.EmailVerified {
		t.Fatal("expected email_verified from Google")
	}
}

func TestCompleteGoogleLoginLinksVerifiedEmailIdentity(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000020")
	users := mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (entity.User, error) {
			return entity.User{
				ID:            userID,
				PublicID:      "usr_existing",
				Email:         email,
				EmailVerified: true,
			}, nil
		},
		getByIDFn: func(_ context.Context, id uuid.UUID) (entity.User, error) {
			return entity.User{ID: id, PublicID: "usr_existing", Email: "link@example.com", EmailVerified: true}, nil
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "google-link-sub", Email: "link@example.com", EmailVerified: true, Name: "Link"},
	}
	svc := newTestAuth(t, google, nil, identities, users, fakeTx{})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if session.IsNewUser || session.User.ID != userID {
		t.Fatalf("expected existing user link, got %#v", session)
	}
	if identities.created != 1 {
		t.Fatalf("expected one linked identity, got %d", identities.created)
	}
	if identities.last.Provider != constant.AuthProviderGoogle || identities.last.ProviderSubject != "google-link-sub" {
		t.Fatalf("identity = %#v", identities.last)
	}
	if identities.last.UserID != userID {
		t.Fatalf("linked to wrong user: %s", identities.last.UserID)
	}
}

func TestCompleteGoogleLoginAutoLinksCalendar(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000010")
	users := mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (entity.User, error) {
			return entity.User{ID: id, PublicID: "usr_x", Email: "a@b.com", EmailVerified: true}, nil
		},
	}
	identities := &mockIdentityRepo{
		identity: entity.AuthIdentity{UserID: userID, Provider: constant.AuthProviderGoogle, ProviderSubject: "sub-1"},
	}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", RefreshToken: "rt", Scope: "openid email profile calendar"},
		profile:  googleoauth.Profile{Subject: "sub-1", Email: "a@b.com", EmailVerified: true, Name: "Ada"},
	}
	linker := &mockLoginLinker{}
	svc := newTestAuth(t, google, nil, identities, users, fakeTx{})
	svc.SetLoginCalendarLinker(linker)

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteGoogleLogin: %v", err)
	}
	if linker.googleCalls != 1 {
		t.Fatalf("expected calendar auto-link once, got %d", linker.googleCalls)
	}
	if linker.lastGoogleUser != session.User.ID || linker.lastGoogleSub != "sub-1" {
		t.Fatalf("auto-link args user=%s sub=%s", linker.lastGoogleUser, linker.lastGoogleSub)
	}
}

func TestCompleteMicrosoftLoginAutoLinksCalendar(t *testing.T) {
	users := mockUserRepo{
		createFn: func(_ context.Context, user entity.User) (entity.User, error) {
			return user, nil
		},
		getByEmailFn: func(context.Context, string) (entity.User, error) {
			return entity.User{}, apperr.ErrNotFound
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	ms := mockMicrosoft{
		tokenSet: microsoftoauth.TokenSet{AccessToken: "at", RefreshToken: "rt", Scope: "Calendars.ReadWrite"},
		profile:  microsoftoauth.Profile{Subject: "ms-sub", Email: "ms@example.com", EmailVerified: true, Name: "Ms"},
	}
	linker := &mockLoginLinker{}
	svc := newTestAuth(t, mockGoogle{}, ms, identities, users, fakeTx{})
	svc.SetLoginCalendarLinker(linker)

	session, err := svc.CompleteMicrosoftLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteMicrosoftLogin: %v", err)
	}
	if linker.microsoftCalls != 1 {
		t.Fatalf("expected calendar auto-link once, got %d", linker.microsoftCalls)
	}
	if linker.lastMSUser != session.User.ID || linker.lastMSSub != "ms-sub" {
		t.Fatalf("auto-link args user=%s sub=%s", linker.lastMSUser, linker.lastMSSub)
	}
}

func TestCompleteGoogleLoginContinuesWhenAutoLinkFails(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000010")
	users := mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (entity.User, error) {
			return entity.User{ID: id, PublicID: "usr_x", Email: "a@b.com", EmailVerified: true}, nil
		},
	}
	identities := &mockIdentityRepo{
		identity: entity.AuthIdentity{UserID: userID, Provider: constant.AuthProviderGoogle, ProviderSubject: "sub-1"},
	}
	google := mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at", Scope: "openid email profile"},
		profile:  googleoauth.Profile{Subject: "sub-1", Email: "a@b.com", EmailVerified: true, Name: "Ada"},
	}
	svc := newTestAuth(t, google, nil, identities, users, fakeTx{})
	svc.SetLoginCalendarLinker(&mockLoginLinker{err: errors.New("link boom")})

	session, err := svc.CompleteGoogleLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("login should succeed even if auto-link fails: %v", err)
	}
	if session.AccessToken == "" {
		t.Fatal("expected session token")
	}
}

func TestCompleteMicrosoftLoginNewUser(t *testing.T) {
	users := mockUserRepo{
		createFn: func(_ context.Context, user entity.User) (entity.User, error) {
			return user, nil
		},
		getByEmailFn: func(context.Context, string) (entity.User, error) {
			return entity.User{}, apperr.ErrNotFound
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	ms := mockMicrosoft{
		tokenSet: microsoftoauth.TokenSet{AccessToken: "at", Scope: "openid email profile User.Read"},
		profile:  microsoftoauth.Profile{Subject: "ms-sub", Email: "ms@example.com", EmailVerified: true, Name: "Ms"},
	}
	svc := newTestAuth(t, mockGoogle{}, ms, identities, users, fakeTx{})

	session, err := svc.CompleteMicrosoftLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteMicrosoftLogin: %v", err)
	}
	if !session.IsNewUser || session.User.Email != "ms@example.com" {
		t.Fatalf("session = %#v", session)
	}
	if identities.created != 1 || identities.last.Provider != constant.AuthProviderMicrosoft {
		t.Fatalf("identity=%#v created=%d", identities.last, identities.created)
	}
}

func TestCompleteMicrosoftLoginLinksVerifiedEmail(t *testing.T) {
	userID := uuid.MustParse("01900000-0000-7000-8000-000000000030")
	users := mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (entity.User, error) {
			return entity.User{
				ID:            userID,
				PublicID:      "usr_ms_link",
				Email:         email,
				EmailVerified: true,
			}, nil
		},
		getByIDFn: func(_ context.Context, id uuid.UUID) (entity.User, error) {
			return entity.User{ID: id, PublicID: "usr_ms_link", Email: "shared@example.com", EmailVerified: true}, nil
		},
	}
	identities := &mockIdentityRepo{err: apperr.ErrNotFound}
	ms := mockMicrosoft{
		tokenSet: microsoftoauth.TokenSet{AccessToken: "at", Scope: "openid email profile User.Read"},
		profile:  microsoftoauth.Profile{Subject: "ms-link-sub", Email: "shared@example.com", EmailVerified: true, Name: "Shared"},
	}
	svc := newTestAuth(t, mockGoogle{}, ms, identities, users, fakeTx{})

	session, err := svc.CompleteMicrosoftLogin(context.Background(), "code", validState(t))
	if err != nil {
		t.Fatalf("CompleteMicrosoftLogin: %v", err)
	}
	if session.IsNewUser || session.User.ID != userID {
		t.Fatalf("expected email link to existing user, got %#v", session)
	}
	if identities.created != 1 || identities.last.Provider != constant.AuthProviderMicrosoft {
		t.Fatalf("identity=%#v created=%d", identities.last, identities.created)
	}
	if identities.last.UserID != userID {
		t.Fatalf("linked to wrong user: %s", identities.last.UserID)
	}
}

func TestCompleteGoogleLoginRejectsBadState(t *testing.T) {
	svc := newTestAuth(t, mockGoogle{
		tokenSet: googleoauth.TokenSet{AccessToken: "at"},
		profile:  googleoauth.Profile{Subject: "sub"},
	}, nil, &mockIdentityRepo{err: apperr.ErrNotFound}, mockUserRepo{}, fakeTx{})

	_, err := svc.CompleteGoogleLogin(context.Background(), "code", "not-a-valid-state")
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
