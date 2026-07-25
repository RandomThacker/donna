package business

import (
	"context"
	"encoding/json"
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
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/RandomThacker/donna/services/api/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GoogleOAuthClient is the outbound Google OAuth port.
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

// sealedTokens is the JSON payload stored in credential_secrets.
type sealedTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiryUnix   int64  `json:"expiry_unix,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// TxRunner runs work inside a database transaction.
type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// AuthService orchestrates Google OAuth login and account linking.
type AuthService struct {
	users      *UserService
	userRepo   repository.UserRepository
	identities repository.AuthIdentityRepository
	accounts   repository.ConnectedAccountRepository
	secrets    repository.CredentialSecretRepository
	tx         TxRunner
	google     GoogleOAuthClient
	state      *oauthstate.Manager
	tokens     *session.Issuer
	sealKey    []byte
	log        *logger.Logger
	now        func() time.Time
	// onGoogleAccountReady runs after Google connected_accounts upsert (e.g. calendar sync bootstrap).
	onGoogleAccountReady func(ctx context.Context, accountID uuid.UUID) error
}

// AuthServiceDeps wires AuthService dependencies.
type AuthServiceDeps struct {
	Users      *UserService
	UserRepo   repository.UserRepository
	Identities repository.AuthIdentityRepository
	Accounts   repository.ConnectedAccountRepository
	Secrets    repository.CredentialSecretRepository
	Tx         TxRunner
	Google     GoogleOAuthClient
	State      *oauthstate.Manager
	Tokens     *session.Issuer
	SealKey    []byte
	Log        *logger.Logger
}

// NewAuthService constructs an AuthService.
func NewAuthService(deps AuthServiceDeps) *AuthService {
	return &AuthService{
		users:      deps.Users,
		userRepo:   deps.UserRepo,
		identities: deps.Identities,
		accounts:   deps.Accounts,
		secrets:    deps.Secrets,
		tx:         deps.Tx,
		google:     deps.Google,
		state:      deps.State,
		tokens:     deps.Tokens,
		sealKey:    deps.SealKey,
		log:        deps.Log,
		now:        time.Now,
	}
}

// SetOnGoogleAccountReady registers a post-connect hook (wired after calendar module).
func (s *AuthService) SetOnGoogleAccountReady(fn func(ctx context.Context, accountID uuid.UUID) error) {
	s.onGoogleAccountReady = fn
}

func (s *AuthService) notifyGoogleAccountReady(ctx context.Context, accountID uuid.UUID) {
	if s.onGoogleAccountReady == nil || accountID == uuid.Nil {
		return
	}
	if err := s.onGoogleAccountReady(ctx, accountID); err != nil && s.log != nil {
		s.log.Warn(ctx, "google account ready hook failed",
			constant.LogAttrError, err,
			"connected_account_id", accountID.String(),
		)
	}
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

// CompleteGoogleLogin exchanges the code, upserts identity/account, and issues a JWT.
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

	email := ""
	if strings.TrimSpace(profile.Email) != "" {
		normalized, emailErr := validation.EmailQuery(profile.Email)
		if emailErr != nil {
			return AuthSession{}, emailErr
		}
		email = normalized
	}

	identity, err := s.identities.GetByProviderSubject(ctx, constant.AuthProviderGoogle, profile.Subject)
	switch {
	case err == nil:
		return s.loginExisting(ctx, identity, profile, tokenSet)
	case errors.Is(err, apperr.ErrNotFound):
		return s.registerNew(ctx, profile, email, tokenSet)
	default:
		return AuthSession{}, err
	}
}

func (s *AuthService) loginExisting(
	ctx context.Context,
	identity entity.AuthIdentity,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) (AuthSession, error) {
	user, err := s.users.GetByID(ctx, identity.UserID)
	if err != nil {
		return AuthSession{}, err
	}

	now := s.now().UTC()
	if _, err := s.userRepo.TouchLastLogin(ctx, user.ID, now); err != nil {
		return AuthSession{}, err
	}
	user.LastLoginAt = &now

	account, err := s.upsertConnectedAccount(ctx, user, profile, tokenSet)
	if err != nil {
		return AuthSession{}, err
	}
	s.notifyGoogleAccountReady(ctx, account.ID)

	s.log.AuthEvent(ctx, logger.AuthEventLogin,
		constant.LogAttrUserID, user.ID.String(),
		constant.LogAttrProvider, constant.AuthProviderGoogle,
	)

	return s.issueSession(user, false)
}

func (s *AuthService) registerNew(
	ctx context.Context,
	profile googleoauth.Profile,
	email string,
	tokenSet googleoauth.TokenSet,
) (AuthSession, error) {
	if email == "" {
		return AuthSession{}, fmt.Errorf("%w: google account email is required", apperr.ErrValidation)
	}

	var created entity.User
	var connected entity.ConnectedAccount
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := s.userRepo.WithTx(tx)
		identityRepo := s.identities.WithTx(tx)
		accountRepo := s.accounts.WithTx(tx)
		secretRepo := s.secrets.WithTx(tx)

		var displayName *string
		if name := strings.TrimSpace(profile.Name); name != "" {
			displayName = &name
		}
		var avatar *string
		if pic := strings.TrimSpace(profile.Picture); pic != "" {
			avatar = &pic
		}

		user, err := s.users.CreateWithRepo(ctx, userRepo, CreateUserInput{
			Email:         email,
			EmailVerified: profile.EmailVerified,
			DisplayName:   displayName,
			AvatarURL:     avatar,
			Timezone:      constant.DefaultUserTimezone,
		})
		if err != nil {
			return err
		}

		now := s.now().UTC()
		emailCopy := email
		identityID, err := idgen.NewUUIDv7()
		if err != nil {
			return err
		}
		_, err = identityRepo.Create(ctx, entity.AuthIdentity{
			ID:              identityID,
			PublicID:        idgen.PublicID(constant.PublicIDPrefixAuthIdentity, identityID),
			UserID:          user.ID,
			Provider:        constant.AuthProviderGoogle,
			ProviderSubject: profile.Subject,
			Email:           &emailCopy,
			EmailVerified:   profile.EmailVerified,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
		if err != nil {
			return err
		}

		account, err := s.storeConnectedAccount(ctx, accountRepo, secretRepo, user, profile, tokenSet)
		if err != nil {
			return err
		}
		connected = account

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

	s.notifyGoogleAccountReady(ctx, connected.ID)

	s.log.AuthEvent(ctx, logger.AuthEventLogin,
		constant.LogAttrUserID, created.ID.String(),
		constant.LogAttrProvider, constant.AuthProviderGoogle,
		constant.LogAttrEvent, "auth.signup",
	)

	return s.issueSession(created, true)
}

func (s *AuthService) upsertConnectedAccount(
	ctx context.Context,
	user entity.User,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	existing, err := s.accounts.GetByProviderAccount(ctx, constant.AuthProviderGoogle, profile.Subject)
	if errors.Is(err, apperr.ErrNotFound) {
		return s.storeConnectedAccount(ctx, s.accounts, s.secrets, user, profile, tokenSet)
	}
	if err != nil {
		return entity.ConnectedAccount{}, err
	}

	now := s.now().UTC()
	ciphertext, expiresAt, scopes, err := s.sealTokenSet(ctx, tokenSet, existing, now)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}

	_, err = s.secrets.UpdateCiphertext(ctx, existing.CredentialsRef, ciphertext, now)
	if errors.Is(err, apperr.ErrNotFound) {
		ref, cerr := s.createSecret(ctx, s.secrets, ciphertext, now)
		if cerr != nil {
			return entity.ConnectedAccount{}, cerr
		}
		return s.accounts.UpdateCredentials(ctx, existing.ID, ref, expiresAt, scopes, constant.ConnectedAccountStatusActive, now)
	}
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	return s.accounts.UpdateCredentials(ctx, existing.ID, existing.CredentialsRef, expiresAt, scopes, constant.ConnectedAccountStatusActive, now)
}

func (s *AuthService) storeConnectedAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	user entity.User,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	now := s.now().UTC()
	ciphertext, expiresAt, scopes, err := s.sealTokenSet(ctx, tokenSet, entity.ConnectedAccount{}, now)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	ref, err := s.createSecret(ctx, secrets, ciphertext, now)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}

	accountID, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	var displayName *string
	if name := strings.TrimSpace(profile.Name); name != "" {
		displayName = &name
	}
	return accounts.Create(ctx, entity.ConnectedAccount{
		ID:                accountID,
		PublicID:          idgen.PublicID(constant.PublicIDPrefixConnectedAccount, accountID),
		UserID:            user.ID,
		Provider:          constant.AuthProviderGoogle,
		ProviderAccountID: profile.Subject,
		DisplayName:       displayName,
		Status:            constant.ConnectedAccountStatusActive,
		Scopes:            scopes,
		CredentialsRef:    ref,
		TokenExpiresAt:    expiresAt,
		ProviderMetadata:  []byte("{}"),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *AuthService) createSecret(ctx context.Context, secrets repository.CredentialSecretRepository, ciphertext []byte, now time.Time) (string, error) {
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return "", err
	}
	ref := idgen.PublicID(constant.PublicIDPrefixCredential, id)
	_, err = secrets.Create(ctx, entity.CredentialSecret{
		ID:         id,
		Ref:        ref,
		Ciphertext: ciphertext,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		return "", err
	}
	return ref, nil
}

func (s *AuthService) sealTokenSet(ctx context.Context, tokenSet googleoauth.TokenSet, existing entity.ConnectedAccount, now time.Time) ([]byte, *time.Time, []string, error) {
	refresh := tokenSet.RefreshToken
	if refresh == "" && existing.CredentialsRef != "" {
		// Preserve previously sealed refresh token when Google omits a new one.
		prev, err := s.secrets.GetByRef(ctx, existing.CredentialsRef)
		if err == nil {
			if plain, derr := seal.Decrypt(s.sealKey, prev.Ciphertext); derr == nil {
				var old sealedTokens
				if json.Unmarshal(plain, &old) == nil && old.RefreshToken != "" {
					refresh = old.RefreshToken
				}
			}
		}
	}
	if refresh == "" {
		return nil, nil, nil, fmt.Errorf("%w: google did not return a refresh token", apperr.ErrInvalid)
	}

	var expiresAt *time.Time
	expiryUnix := int64(0)
	if tokenSet.ExpiresIn > 0 {
		t := now.Add(time.Duration(tokenSet.ExpiresIn) * time.Second)
		expiresAt = &t
		expiryUnix = t.Unix()
	}

	payload, err := json.Marshal(sealedTokens{
		AccessToken:  tokenSet.AccessToken,
		RefreshToken: refresh,
		TokenType:    tokenSet.TokenType,
		ExpiryUnix:   expiryUnix,
		Scope:        tokenSet.Scope,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("marshal tokens: %w", err)
	}
	ciphertext, err := seal.Encrypt(s.sealKey, payload)
	if err != nil {
		return nil, nil, nil, err
	}

	scopes := splitScopes(tokenSet.Scope)
	if len(scopes) == 0 {
		scopes = []string{
			constant.GoogleScopeOpenID,
			constant.GoogleScopeEmail,
			constant.GoogleScopeProfile,
		}
	}
	return ciphertext, expiresAt, scopes, nil
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
