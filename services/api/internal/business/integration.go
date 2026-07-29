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
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MicrosoftOAuthClient is the outbound Microsoft OAuth port (login or integration).
type MicrosoftOAuthClient interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (microsoftoauth.TokenSet, microsoftoauth.Profile, error)
	FetchProfile(ctx context.Context, accessToken string) (microsoftoauth.Profile, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (microsoftoauth.TokenSet, error)
}

// sealedTokens is the JSON payload stored in credential_secrets.
type sealedTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiryUnix   int64  `json:"expiry_unix,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// IntegrationService connects third-party calendar accounts (not login).
// Connecting never changes the logged-in Donna user.
type IntegrationService struct {
	accounts       repository.ConnectedAccountRepository
	secrets        repository.CredentialSecretRepository
	sources        repository.CalendarSourceRepository
	events         repository.CalendarEventRepository
	tx             TxRunner
	google         GoogleOAuthClient
	microsoft      MicrosoftOAuthClient
	state          *oauthstate.Manager
	sealKey        []byte
	log            *logger.Logger
	now            func() time.Time
	onAccountReady func(ctx context.Context, accountID uuid.UUID) error
	syncAccount    func(ctx context.Context, accountID uuid.UUID) (CalendarPipelineResult, error)
}

// IntegrationServiceDeps wires IntegrationService.
type IntegrationServiceDeps struct {
	Accounts  repository.ConnectedAccountRepository
	Secrets   repository.CredentialSecretRepository
	Sources   repository.CalendarSourceRepository
	Events    repository.CalendarEventRepository
	Tx        TxRunner
	Google    GoogleOAuthClient
	Microsoft MicrosoftOAuthClient
	State     *oauthstate.Manager
	SealKey   []byte
	Log       *logger.Logger
}

// NewIntegrationService constructs an IntegrationService.
func NewIntegrationService(deps IntegrationServiceDeps) *IntegrationService {
	return &IntegrationService{
		accounts:  deps.Accounts,
		secrets:   deps.Secrets,
		sources:   deps.Sources,
		events:    deps.Events,
		tx:        deps.Tx,
		google:    deps.Google,
		microsoft: deps.Microsoft,
		state:     deps.State,
		sealKey:   deps.SealKey,
		log:       deps.Log,
		now:       time.Now,
	}
}

// SetOnAccountReady registers a post-connect hook (e.g. calendar sync bootstrap).
func (s *IntegrationService) SetOnAccountReady(fn func(ctx context.Context, accountID uuid.UUID) error) {
	s.onAccountReady = fn
}

// SetSyncAccount registers a manual sync hook used by ICS sync-now.
func (s *IntegrationService) SetSyncAccount(fn func(ctx context.Context, accountID uuid.UUID) (CalendarPipelineResult, error)) {
	s.syncAccount = fn
}

// ListConnectedAccounts returns live connected accounts for a user.
// Accounts missing a stored email are enriched from the provider profile once.
func (s *IntegrationService) ListConnectedAccounts(ctx context.Context, userID uuid.UUID) ([]entity.ConnectedAccount, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	accounts, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if !accountNeedsProfileBackfill(accounts[i]) {
			continue
		}
		if enriched, ok := s.backfillAccountProfile(ctx, accounts[i]); ok {
			accounts[i] = enriched
		}
	}
	return accounts, nil
}

// Disconnect soft-deletes the connected account and permanently removes its
// calendar sources + events from Donna.
func (s *IntegrationService) Disconnect(ctx context.Context, userID, accountID uuid.UUID) error {
	if userID == uuid.Nil || accountID == uuid.Nil {
		return fmt.Errorf("%w: user id and account id are required", apperr.ErrValidation)
	}
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.UserID != userID {
		return apperr.ErrForbidden
	}
	now := s.now().UTC()

	cleanup := func(
		ctx context.Context,
		accounts repository.ConnectedAccountRepository,
		sources repository.CalendarSourceRepository,
		events repository.CalendarEventRepository,
	) error {
		// Events first (FK RESTRICT on calendar_source_id), then sources, then account.
		if events != nil {
			if _, err := events.DeleteByConnectedAccountID(ctx, accountID); err != nil {
				return err
			}
		}
		if sources != nil {
			if _, err := sources.DeleteByConnectedAccountID(ctx, accountID); err != nil {
				return err
			}
		}
		if _, err := accounts.SoftDelete(ctx, accountID, now); err != nil {
			return err
		}
		return nil
	}

	if s.tx != nil && (s.sources != nil || s.events != nil) {
		return s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			var sourcesRepo repository.CalendarSourceRepository
			var eventsRepo repository.CalendarEventRepository
			if s.sources != nil {
				sourcesRepo = s.sources.WithTx(tx)
			}
			if s.events != nil {
				eventsRepo = s.events.WithTx(tx)
			}
			return cleanup(ctx, s.accounts.WithTx(tx), sourcesRepo, eventsRepo)
		})
	}
	return cleanup(ctx, s.accounts, s.sources, s.events)
}

// BeginGoogleConnect starts Google calendar OAuth for an authenticated Donna user.
func (s *IntegrationService) BeginGoogleConnect(ctx context.Context, userID uuid.UUID) (authURL string, err error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if s.google == nil {
		return "", fmt.Errorf("%w: google oauth is not configured", apperr.ErrInvalid)
	}
	state, err := s.state.CreateWithUser(userID.String())
	if err != nil {
		return "", err
	}
	return s.google.AuthCodeURL(state), nil
}

// CompleteGoogleConnect exchanges the OAuth code and upserts a google connected_account.
func (s *IntegrationService) CompleteGoogleConnect(ctx context.Context, code, state string) (entity.ConnectedAccount, error) {
	if s.google == nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google oauth is not configured", apperr.ErrInvalid)
	}
	userID, err := s.verifyBoundUser(state)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: authorization code is required", apperr.ErrValidation)
	}

	tokenSet, err := s.google.ExchangeCode(ctx, code)
	if err != nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google oauth exchange failed", apperr.ErrInvalid)
	}
	profile, err := s.google.FetchProfile(ctx, tokenSet.AccessToken)
	if err != nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google profile fetch failed", apperr.ErrInvalid)
	}
	subject := strings.TrimSpace(profile.Subject)
	if subject == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google profile missing subject", apperr.ErrInvalid)
	}

	var account entity.ConnectedAccount
	err = s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		created, upsertErr := s.upsertGoogleAccount(ctx, s.accounts.WithTx(tx), s.secrets.WithTx(tx), userID, profile, tokenSet)
		if upsertErr != nil {
			return upsertErr
		}
		account = created
		return nil
	})
	if err != nil {
		return entity.ConnectedAccount{}, err
	}

	s.notifyAccountReady(ctx, account.ID)
	return account, nil
}

// LinkGoogleFromLogin upserts a Google connected_account for the signed-in Donna user.
// Used so the login identity is automatically available for calendar sync.
func (s *IntegrationService) LinkGoogleFromLogin(
	ctx context.Context,
	userID uuid.UUID,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	if userID == uuid.Nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if strings.TrimSpace(profile.Subject) == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google profile missing subject", apperr.ErrValidation)
	}
	var account entity.ConnectedAccount
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		created, upsertErr := s.upsertGoogleAccount(ctx, s.accounts.WithTx(tx), s.secrets.WithTx(tx), userID, profile, tokenSet)
		if upsertErr != nil {
			return upsertErr
		}
		account = created
		return nil
	})
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	s.notifyAccountReady(ctx, account.ID)
	return account, nil
}

// LinkMicrosoftFromLogin upserts a Microsoft connected_account for the signed-in Donna user.
func (s *IntegrationService) LinkMicrosoftFromLogin(
	ctx context.Context,
	userID uuid.UUID,
	profile microsoftoauth.Profile,
	tokenSet microsoftoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	if userID == uuid.Nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if strings.TrimSpace(profile.Subject) == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: microsoft profile missing subject", apperr.ErrValidation)
	}
	var account entity.ConnectedAccount
	err := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		created, upsertErr := s.upsertMicrosoftAccount(ctx, s.accounts.WithTx(tx), s.secrets.WithTx(tx), userID, profile, tokenSet)
		if upsertErr != nil {
			return upsertErr
		}
		account = created
		return nil
	})
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	s.notifyAccountReady(ctx, account.ID)
	return account, nil
}

// BeginMicrosoftConnect starts Microsoft calendar OAuth for an authenticated Donna user.
func (s *IntegrationService) BeginMicrosoftConnect(ctx context.Context, userID uuid.UUID) (authURL string, err error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if s.microsoft == nil {
		return "", fmt.Errorf("%w: microsoft oauth is not configured", apperr.ErrInvalid)
	}
	state, err := s.state.CreateWithUser(userID.String())
	if err != nil {
		return "", err
	}
	return s.microsoft.AuthCodeURL(state), nil
}

// CompleteMicrosoftConnect exchanges the OAuth code and upserts a microsoft connected_account.
func (s *IntegrationService) CompleteMicrosoftConnect(ctx context.Context, code, state string) (entity.ConnectedAccount, error) {
	if s.microsoft == nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: microsoft oauth is not configured", apperr.ErrInvalid)
	}
	userID, err := s.verifyBoundUser(state)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: authorization code is required", apperr.ErrValidation)
	}

	tokenSet, profile, err := s.microsoft.Exchange(ctx, code)
	if err != nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: microsoft oauth exchange failed", apperr.ErrInvalid)
	}
	subject := strings.TrimSpace(profile.Subject)
	if subject == "" {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: microsoft profile missing id", apperr.ErrInvalid)
	}

	var account entity.ConnectedAccount
	err = s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		created, upsertErr := s.upsertMicrosoftAccount(ctx, s.accounts.WithTx(tx), s.secrets.WithTx(tx), userID, profile, tokenSet)
		if upsertErr != nil {
			return upsertErr
		}
		account = created
		return nil
	})
	if err != nil {
		return entity.ConnectedAccount{}, err
	}

	s.notifyAccountReady(ctx, account.ID)
	return account, nil
}

func (s *IntegrationService) verifyBoundUser(state string) (uuid.UUID, error) {
	userIDRaw, err := s.state.VerifyWithUser(state)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid oauth state", apperr.ErrValidation)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid oauth state user", apperr.ErrValidation)
	}
	return userID, nil
}

func (s *IntegrationService) notifyAccountReady(ctx context.Context, accountID uuid.UUID) {
	if s.onAccountReady == nil || accountID == uuid.Nil {
		return
	}
	if err := s.onAccountReady(ctx, accountID); err != nil && s.log != nil {
		s.log.Warn(ctx, "integration account ready hook failed",
			constant.LogAttrError, err,
			"connected_account_id", accountID.String(),
		)
	}
}

func (s *IntegrationService) upsertGoogleAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	userID uuid.UUID,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	now := s.now().UTC()
	existing, err := accounts.GetByProviderAccount(ctx, constant.AuthProviderGoogle, profile.Subject)
	if errors.Is(err, apperr.ErrNotFound) {
		if !tokenHasCalendarScope(tokenSet.Scope) {
			return entity.ConnectedAccount{}, fmt.Errorf(
				"%w: google did not grant calendar access; reconnect with calendar permission",
				apperr.ErrInvalid,
			)
		}
		return s.storeGoogleAccount(ctx, accounts, secrets, userID, profile, tokenSet, now)
	}
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	if existing.UserID != userID {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google account already linked to another user", apperr.ErrConflict)
	}

	// Login auto-link can return a weaker token (no calendar). Never overwrite a
	// working calendar connection with a downgraded grant.
	if !tokenHasCalendarScope(tokenSet.Scope) && (hasCalendarScope(existing.Scopes) || calendarAccessProven(existing)) {
		return accounts.UpdateProfile(
			ctx,
			existing.ID,
			integrationDisplayName(profile.Name, profile.Email),
			integrationProviderMetadata(profile.Email, profile.Picture),
			now,
		)
	}

	ciphertext, expiresAt, scopes, sealErr := s.sealGoogleTokens(ctx, secrets, tokenSet, existing, now)
	if sealErr != nil {
		return entity.ConnectedAccount{}, sealErr
	}
	_, err = secrets.UpdateCiphertext(ctx, existing.CredentialsRef, ciphertext, now)
	if errors.Is(err, apperr.ErrNotFound) {
		ref, cerr := s.createSecret(ctx, secrets, ciphertext, now)
		if cerr != nil {
			return entity.ConnectedAccount{}, cerr
		}
		if _, uerr := accounts.UpdateCredentials(ctx, existing.ID, ref, expiresAt, scopes, constant.ConnectedAccountStatusActive, now); uerr != nil {
			return entity.ConnectedAccount{}, uerr
		}
	} else if err != nil {
		return entity.ConnectedAccount{}, err
	} else if _, uerr := accounts.UpdateCredentials(ctx, existing.ID, existing.CredentialsRef, expiresAt, scopes, constant.ConnectedAccountStatusActive, now); uerr != nil {
		return entity.ConnectedAccount{}, uerr
	}
	return accounts.UpdateProfile(ctx, existing.ID, integrationDisplayName(profile.Name, profile.Email), integrationProviderMetadata(profile.Email, profile.Picture), now)
}

func (s *IntegrationService) storeGoogleAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	userID uuid.UUID,
	profile googleoauth.Profile,
	tokenSet googleoauth.TokenSet,
	now time.Time,
) (entity.ConnectedAccount, error) {
	ciphertext, expiresAt, scopes, err := s.sealGoogleTokens(ctx, secrets, tokenSet, entity.ConnectedAccount{}, now)
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
	} else if email := strings.TrimSpace(profile.Email); email != "" {
		displayName = &email
	}
	return accounts.Create(ctx, entity.ConnectedAccount{
		ID:                accountID,
		PublicID:          idgen.PublicID(constant.PublicIDPrefixConnectedAccount, accountID),
		UserID:            userID,
		Provider:          constant.AuthProviderGoogle,
		ProviderAccountID: profile.Subject,
		DisplayName:       displayName,
		Status:            constant.ConnectedAccountStatusActive,
		Scopes:            scopes,
		CredentialsRef:    ref,
		TokenExpiresAt:    expiresAt,
		ProviderMetadata:  integrationProviderMetadata(profile.Email, profile.Picture),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *IntegrationService) upsertMicrosoftAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	userID uuid.UUID,
	profile microsoftoauth.Profile,
	tokenSet microsoftoauth.TokenSet,
) (entity.ConnectedAccount, error) {
	now := s.now().UTC()
	existing, err := accounts.GetByProviderAccount(ctx, constant.AuthProviderMicrosoft, profile.Subject)
	if errors.Is(err, apperr.ErrNotFound) {
		return s.storeMicrosoftAccount(ctx, accounts, secrets, userID, profile, tokenSet, now)
	}
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	if existing.UserID != userID {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: microsoft account already linked to another user", apperr.ErrConflict)
	}

	ciphertext, expiresAt, scopes, sealErr := s.sealMicrosoftTokens(tokenSet, now)
	if sealErr != nil {
		return entity.ConnectedAccount{}, sealErr
	}
	_, err = secrets.UpdateCiphertext(ctx, existing.CredentialsRef, ciphertext, now)
	if errors.Is(err, apperr.ErrNotFound) {
		ref, cerr := s.createSecret(ctx, secrets, ciphertext, now)
		if cerr != nil {
			return entity.ConnectedAccount{}, cerr
		}
		if _, uerr := accounts.UpdateCredentials(ctx, existing.ID, ref, expiresAt, scopes, constant.ConnectedAccountStatusActive, now); uerr != nil {
			return entity.ConnectedAccount{}, uerr
		}
	} else if err != nil {
		return entity.ConnectedAccount{}, err
	} else if _, uerr := accounts.UpdateCredentials(ctx, existing.ID, existing.CredentialsRef, expiresAt, scopes, constant.ConnectedAccountStatusActive, now); uerr != nil {
		return entity.ConnectedAccount{}, uerr
	}
	return accounts.UpdateProfile(ctx, existing.ID, integrationDisplayName(profile.Name, profile.Email), integrationProviderMetadata(profile.Email, ""), now)
}

func (s *IntegrationService) storeMicrosoftAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	userID uuid.UUID,
	profile microsoftoauth.Profile,
	tokenSet microsoftoauth.TokenSet,
	now time.Time,
) (entity.ConnectedAccount, error) {
	ciphertext, expiresAt, scopes, err := s.sealMicrosoftTokens(tokenSet, now)
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
	} else if email := strings.TrimSpace(profile.Email); email != "" {
		displayName = &email
	}
	return accounts.Create(ctx, entity.ConnectedAccount{
		ID:                accountID,
		PublicID:          idgen.PublicID(constant.PublicIDPrefixConnectedAccount, accountID),
		UserID:            userID,
		Provider:          constant.AuthProviderMicrosoft,
		ProviderAccountID: profile.Subject,
		DisplayName:       displayName,
		Status:            constant.ConnectedAccountStatusActive,
		Scopes:            scopes,
		CredentialsRef:    ref,
		TokenExpiresAt:    expiresAt,
		ProviderMetadata:  integrationProviderMetadata(profile.Email, ""),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *IntegrationService) sealGoogleTokens(
	ctx context.Context,
	secrets repository.CredentialSecretRepository,
	tokenSet googleoauth.TokenSet,
	existing entity.ConnectedAccount,
	now time.Time,
) ([]byte, *time.Time, []string, error) {
	refresh := tokenSet.RefreshToken
	if refresh == "" && existing.CredentialsRef != "" {
		prev, err := secrets.GetByRef(ctx, existing.CredentialsRef)
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
			constant.GoogleScopeCalendar,
		}
	}
	// Keep previously stored scopes when Google omits them from an incremental
	// token response, but never invent calendar if this grant didn't include it
	// and the account wasn't already calendar-capable.
	scopes = unionScopes(scopes, existing.Scopes)
	if tokenHasCalendarScope(tokenSet.Scope) {
		scopes = ensureScope(scopes, constant.GoogleScopeCalendar)
	}
	return ciphertext, expiresAt, scopes, nil
}

func (s *IntegrationService) sealMicrosoftTokens(tokenSet microsoftoauth.TokenSet, now time.Time) ([]byte, *time.Time, []string, error) {
	var expiresAt *time.Time
	expiryUnix := int64(0)
	if tokenSet.ExpiresIn > 0 {
		t := now.Add(time.Duration(tokenSet.ExpiresIn) * time.Second)
		expiresAt = &t
		expiryUnix = t.Unix()
	}
	payload, err := json.Marshal(sealedTokens{
		AccessToken:  tokenSet.AccessToken,
		RefreshToken: tokenSet.RefreshToken,
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
			constant.MicrosoftScopeOpenID,
			constant.MicrosoftScopeEmail,
			constant.MicrosoftScopeProfile,
			constant.MicrosoftScopeOfflineAccess,
			constant.MicrosoftScopeUserRead,
			constant.MicrosoftScopeCalendarsReadWrite,
		}
	}
	return ciphertext, expiresAt, scopes, nil
}

func (s *IntegrationService) createSecret(
	ctx context.Context,
	secrets repository.CredentialSecretRepository,
	ciphertext []byte,
	now time.Time,
) (string, error) {
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

func integrationDisplayName(name, email string) *string {
	if n := strings.TrimSpace(name); n != "" {
		return &n
	}
	if e := strings.TrimSpace(email); e != "" {
		return &e
	}
	return nil
}

func integrationProviderMetadata(email, avatarURL string) []byte {
	meta := map[string]string{}
	if email = strings.TrimSpace(email); email != "" {
		meta["email"] = email
	}
	if avatarURL = strings.TrimSpace(avatarURL); avatarURL != "" {
		meta["avatar_url"] = avatarURL
	}
	if len(meta) == 0 {
		return []byte("{}")
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return []byte("{}")
	}
	return payload
}

func accountNeedsProfileBackfill(account entity.ConnectedAccount) bool {
	email := emailFromMetadata(account.ProviderMetadata)
	avatar := avatarURLFromMetadata(account.ProviderMetadata)
	hasEmail := email != ""
	if !hasEmail && account.DisplayName != nil {
		name := strings.TrimSpace(*account.DisplayName)
		hasEmail = strings.Contains(name, "@")
	}
	if !hasEmail {
		return true
	}
	// Google profiles include a public picture URL; backfill once if missing.
	if account.Provider == constant.AuthProviderGoogle && avatar == "" {
		return true
	}
	return false
}

func emailFromMetadata(raw []byte) string {
	return stringFieldFromMetadata(raw, "email")
}

func avatarURLFromMetadata(raw []byte) string {
	return stringFieldFromMetadata(raw, "avatar_url")
}

func stringFieldFromMetadata(raw []byte, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return ""
	}
	v, ok := meta[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func (s *IntegrationService) backfillAccountProfile(ctx context.Context, account entity.ConnectedAccount) (entity.ConnectedAccount, bool) {
	accessToken, err := s.integrationAccessToken(ctx, account)
	if err != nil || accessToken == "" {
		return account, false
	}

	var name, email, avatarURL string
	switch account.Provider {
	case constant.AuthProviderGoogle:
		if s.google == nil {
			return account, false
		}
		profile, ferr := s.google.FetchProfile(ctx, accessToken)
		if ferr != nil {
			return account, false
		}
		name, email, avatarURL = profile.Name, profile.Email, profile.Picture
	case constant.AuthProviderMicrosoft:
		if s.microsoft == nil {
			return account, false
		}
		profile, ferr := s.microsoft.FetchProfile(ctx, accessToken)
		if ferr != nil {
			return account, false
		}
		name, email = profile.Name, profile.Email
	default:
		return account, false
	}

	email = strings.TrimSpace(email)
	avatarURL = strings.TrimSpace(avatarURL)
	if email == "" && avatarURL == "" {
		return account, false
	}
	// Preserve existing email/avatar when the provider omits one field.
	if email == "" {
		email = emailFromMetadata(account.ProviderMetadata)
	}
	if avatarURL == "" {
		avatarURL = avatarURLFromMetadata(account.ProviderMetadata)
	}

	updated, err := s.accounts.UpdateProfile(
		ctx,
		account.ID,
		integrationDisplayName(name, email),
		integrationProviderMetadata(email, avatarURL),
		s.now().UTC(),
	)
	if err != nil {
		return account, false
	}
	return updated, true
}

func (s *IntegrationService) integrationAccessToken(ctx context.Context, account entity.ConnectedAccount) (string, error) {
	if account.CredentialsRef == "" || len(s.sealKey) == 0 {
		return "", fmt.Errorf("credentials unavailable")
	}
	secret, err := s.secrets.GetByRef(ctx, account.CredentialsRef)
	if err != nil {
		return "", err
	}
	plain, err := seal.Decrypt(s.sealKey, secret.Ciphertext)
	if err != nil {
		return "", err
	}
	var tokens sealedTokens
	if err := json.Unmarshal(plain, &tokens); err != nil {
		return "", err
	}
	if tokens.AccessToken == "" {
		return "", fmt.Errorf("missing access token")
	}

	now := s.now().UTC()
	needsRefresh := tokens.ExpiryUnix > 0 && now.Unix() >= (tokens.ExpiryUnix-60)
	if !needsRefresh && account.TokenExpiresAt != nil && !account.TokenExpiresAt.After(now.Add(60*time.Second)) {
		needsRefresh = true
	}
	if !needsRefresh {
		return tokens.AccessToken, nil
	}
	if tokens.RefreshToken == "" {
		return tokens.AccessToken, nil
	}

	switch account.Provider {
	case constant.AuthProviderGoogle:
		if s.google == nil {
			return tokens.AccessToken, nil
		}
		refreshed, rerr := s.google.RefreshAccessToken(ctx, tokens.RefreshToken)
		if rerr != nil {
			return tokens.AccessToken, nil
		}
		if refreshed.AccessToken != "" {
			return refreshed.AccessToken, nil
		}
	case constant.AuthProviderMicrosoft:
		if s.microsoft == nil {
			return tokens.AccessToken, nil
		}
		refreshed, rerr := s.microsoft.RefreshAccessToken(ctx, tokens.RefreshToken)
		if rerr != nil {
			return tokens.AccessToken, nil
		}
		if refreshed.AccessToken != "" {
			return refreshed.AccessToken, nil
		}
	}
	return tokens.AccessToken, nil
}
