package business

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/RandomThacker/donna/services/api/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GoogleOAuthClient is the outbound Google OAuth port (login or integration).
type GoogleOAuthClient interface {
	AuthCodeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (googleoauth.TokenSet, error)
	FetchProfile(ctx context.Context, accessToken string) (googleoauth.Profile, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (googleoauth.TokenSet, error)
}

// AuthSession is the authenticated session returned to the client.
type AuthSession struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	ExpiresAt   time.Time
	User        entity.User
	IsNewUser   bool
}

// TxRunner runs work inside a database transaction.
type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// loginProfile is a provider-normalized identity used for login linking.
type loginProfile struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// loginCalendarLinker auto-connects the signed-in provider account for calendar sync.
type loginCalendarLinker interface {
	LinkGoogleFromLogin(ctx context.Context, userID uuid.UUID, profile googleoauth.Profile, tokenSet googleoauth.TokenSet) (entity.ConnectedAccount, error)
	LinkMicrosoftFromLogin(ctx context.Context, userID uuid.UUID, profile microsoftoauth.Profile, tokenSet microsoftoauth.TokenSet) (entity.ConnectedAccount, error)
}

// AuthService orchestrates multi-provider OAuth login (auth_identities + JWT).
// The signed-in provider account is also linked as a calendar connected_account when a linker is configured.
type AuthService struct {
	users      *UserService
	userRepo   repository.UserRepository
	identities repository.AuthIdentityRepository
	tx         TxRunner
	google     GoogleOAuthClient
	microsoft  MicrosoftOAuthClient
	state      *oauthstate.Manager
	tokens     *session.Issuer
	linker     loginCalendarLinker
	log        *logger.Logger
	now        func() time.Time
}

// AuthServiceDeps wires AuthService dependencies.
type AuthServiceDeps struct {
	Users      *UserService
	UserRepo   repository.UserRepository
	Identities repository.AuthIdentityRepository
	Tx         TxRunner
	Google     GoogleOAuthClient
	Microsoft  MicrosoftOAuthClient
	State      *oauthstate.Manager
	Tokens     *session.Issuer
	Log        *logger.Logger
}

// NewAuthService constructs an AuthService.
func NewAuthService(deps AuthServiceDeps) *AuthService {
	return &AuthService{
		users:      deps.Users,
		userRepo:   deps.UserRepo,
		identities: deps.Identities,
		tx:         deps.Tx,
		google:     deps.Google,
		microsoft:  deps.Microsoft,
		state:      deps.State,
		tokens:     deps.Tokens,
		log:        deps.Log,
		now:        time.Now,
	}
}

// SetLoginCalendarLinker registers the integrations hook that auto-links the login account.
func (s *AuthService) SetLoginCalendarLinker(linker loginCalendarLinker) {
	s.linker = linker
}

// BeginGoogleLogin creates CSRF state and the Google authorization URL.
func (s *AuthService) BeginGoogleLogin(ctx context.Context) (authURL, state string, err error) {
	if s.google == nil {
		return "", "", fmt.Errorf("%w: google oauth is not configured", apperr.ErrInvalid)
	}
	state, err = s.state.Create()
	if err != nil {
		return "", "", fmt.Errorf("oauth state: %w", err)
	}
	return s.google.AuthCodeURL(state), state, nil
}

// CompleteGoogleLogin exchanges the code, links auth_identity, and issues a JWT.
func (s *AuthService) CompleteGoogleLogin(ctx context.Context, code, state string) (AuthSession, error) {
	if s.google == nil {
		return AuthSession{}, fmt.Errorf("%w: google oauth is not configured", apperr.ErrInvalid)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return AuthSession{}, fmt.Errorf("%w: code is required", apperr.ErrValidation)
	}
	if err := s.state.Verify(state); err != nil {
		return AuthSession{}, fmt.Errorf("%w: %v", apperr.ErrInvalid, err)
	}

	tokenSet, err := s.google.ExchangeCode(ctx, code)
	if err != nil {
		return AuthSession{}, fmt.Errorf("%w: token exchange failed: %v", apperr.ErrInvalid, err)
	}
	profile, err := s.google.FetchProfile(ctx, tokenSet.AccessToken)
	if err != nil {
		return AuthSession{}, fmt.Errorf("%w: profile fetch failed: %v", apperr.ErrInvalid, err)
	}
	if profile.Subject == "" {
		return AuthSession{}, fmt.Errorf("%w: google profile missing subject", apperr.ErrInvalid)
	}

	email, err := normalizeLoginEmail(profile.Email)
	if err != nil {
		return AuthSession{}, err
	}

	session, err := s.resolveLogin(ctx, loginProfile{
		Provider:      constant.AuthProviderGoogle,
		Subject:       profile.Subject,
		Email:         email,
		EmailVerified: profile.EmailVerified,
		Name:          profile.Name,
		Picture:       profile.Picture,
	})
	if err != nil {
		return AuthSession{}, err
	}
	s.autoLinkGoogleCalendar(ctx, session.User.ID, profile, tokenSet)
	return session, nil
}

// BeginMicrosoftLogin creates CSRF state and the Microsoft authorization URL.
func (s *AuthService) BeginMicrosoftLogin(ctx context.Context) (authURL, state string, err error) {
	if s.microsoft == nil {
		return "", "", fmt.Errorf("%w: microsoft oauth is not configured", apperr.ErrInvalid)
	}
	state, err = s.state.Create()
	if err != nil {
		return "", "", fmt.Errorf("oauth state: %w", err)
	}
	return s.microsoft.AuthCodeURL(state), state, nil
}

// CompleteMicrosoftLogin exchanges the code, links auth_identity, and issues a JWT.
func (s *AuthService) CompleteMicrosoftLogin(ctx context.Context, code, state string) (AuthSession, error) {
	if s.microsoft == nil {
		return AuthSession{}, fmt.Errorf("%w: microsoft oauth is not configured", apperr.ErrInvalid)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return AuthSession{}, fmt.Errorf("%w: code is required", apperr.ErrValidation)
	}
	if err := s.state.Verify(state); err != nil {
		return AuthSession{}, fmt.Errorf("%w: %v", apperr.ErrInvalid, err)
	}

	tokenSet, profile, err := s.microsoft.Exchange(ctx, code)
	if err != nil {
		return AuthSession{}, fmt.Errorf("%w: token exchange failed: %v", apperr.ErrInvalid, err)
	}
	if strings.TrimSpace(profile.Subject) == "" {
		return AuthSession{}, fmt.Errorf("%w: microsoft profile missing subject", apperr.ErrInvalid)
	}

	email, err := normalizeLoginEmail(profile.Email)
	if err != nil {
		return AuthSession{}, err
	}

	session, err := s.resolveLogin(ctx, loginProfile{
		Provider:      constant.AuthProviderMicrosoft,
		Subject:       profile.Subject,
		Email:         email,
		EmailVerified: profile.EmailVerified,
		Name:          profile.Name,
	})
	if err != nil {
		return AuthSession{}, err
	}
	s.autoLinkMicrosoftCalendar(ctx, session.User.ID, profile, tokenSet)
	return session, nil
}

func (s *AuthService) resolveLogin(ctx context.Context, profile loginProfile) (AuthSession, error) {
	identity, err := s.identities.GetByProviderSubject(ctx, profile.Provider, profile.Subject)
	switch {
	case err == nil:
		return s.loginExisting(ctx, identity, profile.Provider)
	case errors.Is(err, apperr.ErrNotFound):
		return s.loginOrRegister(ctx, profile)
	default:
		return AuthSession{}, err
	}
}

func (s *AuthService) loginOrRegister(ctx context.Context, profile loginProfile) (AuthSession, error) {
	if profile.Email == "" {
		return AuthSession{}, fmt.Errorf("%w: %s account email is required", apperr.ErrValidation, profile.Provider)
	}

	existing, err := s.users.GetByEmail(ctx, profile.Email)
	switch {
	case err == nil:
		if profile.EmailVerified && existing.EmailVerified {
			return s.linkIdentity(ctx, existing, profile)
		}
		return AuthSession{}, fmt.Errorf("%w: email already registered", apperr.ErrConflict)
	case errors.Is(err, apperr.ErrNotFound):
		return s.registerNew(ctx, profile)
	default:
		return AuthSession{}, err
	}
}

func (s *AuthService) loginExisting(ctx context.Context, identity entity.AuthIdentity, provider string) (AuthSession, error) {
	user, err := s.users.GetByID(ctx, identity.UserID)
	if err != nil {
		return AuthSession{}, err
	}

	now := s.now().UTC()
	if _, err := s.userRepo.TouchLastLogin(ctx, user.ID, now); err != nil {
		return AuthSession{}, err
	}
	user.LastLoginAt = &now

	s.log.AuthEvent(ctx, logger.AuthEventLogin,
		constant.LogAttrUserID, user.ID.String(),
		constant.LogAttrProvider, provider,
	)

	return s.issueSession(user, false)
}

func (s *AuthService) linkIdentity(ctx context.Context, user entity.User, profile loginProfile) (AuthSession, error) {
	now := s.now().UTC()
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		identityRepo := s.identities.WithTx(tx)
		userRepo := s.userRepo.WithTx(tx)
		if err := s.createIdentity(ctx, identityRepo, user.ID, profile, now); err != nil {
			return err
		}
		if _, err := userRepo.TouchLastLogin(ctx, user.ID, now); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return AuthSession{}, err
	}
	user.LastLoginAt = &now

	s.log.AuthEvent(ctx, logger.AuthEventLogin,
		constant.LogAttrUserID, user.ID.String(),
		constant.LogAttrProvider, profile.Provider,
		constant.LogAttrEvent, "auth.link_identity",
	)

	return s.issueSession(user, false)
}

func (s *AuthService) registerNew(ctx context.Context, profile loginProfile) (AuthSession, error) {
	var created entity.User
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := s.userRepo.WithTx(tx)
		identityRepo := s.identities.WithTx(tx)

		var displayName *string
		if name := strings.TrimSpace(profile.Name); name != "" {
			displayName = &name
		}
		var avatar *string
		if pic := strings.TrimSpace(profile.Picture); pic != "" {
			avatar = &pic
		}

		user, err := s.users.CreateWithRepo(ctx, userRepo, CreateUserInput{
			Email:         profile.Email,
			EmailVerified: profile.EmailVerified,
			DisplayName:   displayName,
			AvatarURL:     avatar,
			Timezone:      constant.DefaultUserTimezone,
		})
		if err != nil {
			return err
		}

		now := s.now().UTC()
		if err := s.createIdentity(ctx, identityRepo, user.ID, profile, now); err != nil {
			return err
		}
		if _, err := userRepo.TouchLastLogin(ctx, user.ID, now); err != nil {
			return err
		}
		user.LastLoginAt = &now
		created = user
		return nil
	})
	if err != nil {
		return AuthSession{}, err
	}

	s.log.AuthEvent(ctx, logger.AuthEventLogin,
		constant.LogAttrUserID, created.ID.String(),
		constant.LogAttrProvider, profile.Provider,
		constant.LogAttrEvent, "auth.signup",
	)

	return s.issueSession(created, true)
}

func (s *AuthService) createIdentity(
	ctx context.Context,
	identityRepo repository.AuthIdentityRepository,
	userID uuid.UUID,
	profile loginProfile,
	now time.Time,
) error {
	identityID, err := idgen.NewUUIDv7()
	if err != nil {
		return err
	}
	var emailCopy *string
	if profile.Email != "" {
		email := profile.Email
		emailCopy = &email
	}
	_, err = identityRepo.Create(ctx, entity.AuthIdentity{
		ID:              identityID,
		PublicID:        idgen.PublicID(constant.PublicIDPrefixAuthIdentity, identityID),
		UserID:          userID,
		Provider:        profile.Provider,
		ProviderSubject: profile.Subject,
		Email:           emailCopy,
		EmailVerified:   profile.EmailVerified,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	return err
}

func (s *AuthService) issueSession(user entity.User, isNew bool) (AuthSession, error) {
	issued, err := s.tokens.Issue(user.ID, user.PublicID, user.Email)
	if err != nil {
		return AuthSession{}, err
	}
	return AuthSession{
		AccessToken: issued.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   issued.ExpiresIn,
		ExpiresAt:   issued.ExpiresAt,
		User:        user,
		IsNewUser:   isNew,
	}, nil
}

func (s *AuthService) autoLinkGoogleCalendar(
	ctx context.Context,
	userID uuid.UUID,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) {
	if s.linker == nil {
		return
	}
	if _, err := s.linker.LinkGoogleFromLogin(ctx, userID, profile, tokenSet); err != nil && s.log != nil {
		s.log.Warn(ctx, "auto-link google calendar from login failed",
			constant.LogAttrError, err,
			constant.LogAttrUserID, userID.String(),
		)
	}
}

func (s *AuthService) autoLinkMicrosoftCalendar(
	ctx context.Context,
	userID uuid.UUID,
	profile microsoftoauth.Profile,
	tokenSet microsoftoauth.TokenSet,
) {
	if s.linker == nil {
		return
	}
	if _, err := s.linker.LinkMicrosoftFromLogin(ctx, userID, profile, tokenSet); err != nil && s.log != nil {
		s.log.Warn(ctx, "auto-link microsoft calendar from login failed",
			constant.LogAttrError, err,
			constant.LogAttrUserID, userID.String(),
		)
	}
}

func normalizeLoginEmail(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return validation.EmailQuery(raw)
}

func splitScopes(scope string) []string {
	fields := strings.Fields(scope)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func ensureScope(scopes []string, required string) []string {
	required = strings.TrimSpace(required)
	if required == "" {
		return scopes
	}
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == required {
			return scopes
		}
	}
	return append(append([]string{}, scopes...), required)
}

func tokenHasCalendarScope(scope string) bool {
	return hasCalendarScope(splitScopes(scope))
}

func unionScopes(parts ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, list := range parts {
		for _, scope := range list {
			scope = strings.TrimSpace(scope)
			if scope == "" {
				continue
			}
			if _, ok := seen[scope]; ok {
				continue
			}
			seen[scope] = struct{}{}
			out = append(out, scope)
		}
	}
	return out
}
