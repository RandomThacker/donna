package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CalendarSyncResult summarizes an idempotent source sync.
type CalendarSyncResult struct {
	Sources      []entity.CalendarSource
	CreatedCount int
	UpdatedCount int
	RemovedCount int
	SyncedAt     time.Time
	DurationMs   int
	Incremental  bool
	Skipped      bool // EnsureFresh found data fresh enough
	SyncStatus   string
}

// CalendarEventSyncResult summarizes an events sync across active sources.
type CalendarEventSyncResult struct {
	Events       []entity.CalendarEvent
	CreatedCount int
	UpdatedCount int
	RemovedCount int
	SyncedAt     time.Time
	DurationMs   int
	SourceCount  int
}

// CalendarService syncs and lists calendar sources/events (Donna DB is the query source of truth).
type CalendarService struct {
	accounts  repository.ConnectedAccountRepository
	secrets   repository.CredentialSecretRepository
	sources   repository.CalendarSourceRepository
	events    repository.CalendarEventRepository
	syncRuns  repository.CalendarSyncRunRepository
	jobs      repository.SchedulerJobRepository
	tx        TxRunner
	providers map[string]calendarprovider.Provider
	tokens    map[string]calendarprovider.TokenRefresher
	sealKey   []byte
	log       *logger.Logger
	now       func() time.Time
}

// CalendarServiceDeps wires CalendarService.
type CalendarServiceDeps struct {
	Accounts  repository.ConnectedAccountRepository
	Secrets   repository.CredentialSecretRepository
	Sources   repository.CalendarSourceRepository
	Events    repository.CalendarEventRepository
	SyncRuns  repository.CalendarSyncRunRepository
	Jobs      repository.SchedulerJobRepository
	Tx        TxRunner
	Providers map[string]calendarprovider.Provider
	Tokens    map[string]calendarprovider.TokenRefresher
	SealKey   []byte
	Log       *logger.Logger
}

// NewCalendarService constructs a CalendarService.
func NewCalendarService(deps CalendarServiceDeps) *CalendarService {
	providers := deps.Providers
	if providers == nil {
		providers = map[string]calendarprovider.Provider{}
	}
	tokens := deps.Tokens
	if tokens == nil {
		tokens = map[string]calendarprovider.TokenRefresher{}
	}
	return &CalendarService{
		accounts:  deps.Accounts,
		secrets:   deps.Secrets,
		sources:   deps.Sources,
		events:    deps.Events,
		syncRuns:  deps.SyncRuns,
		jobs:      deps.Jobs,
		tx:        deps.Tx,
		providers: providers,
		tokens:    tokens,
		sealKey:   deps.SealKey,
		log:       deps.Log,
		now:       time.Now,
	}
}

// googleTokenAdapter adapts GoogleOAuthClient to calendarprovider.TokenRefresher.
type googleTokenAdapter struct {
	client GoogleOAuthClient
}

func (a googleTokenAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (calendarprovider.TokenSet, error) {
	ts, err := a.client.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return calendarprovider.TokenSet{}, err
	}
	return calendarprovider.TokenSet{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
		TokenType:    ts.TokenType,
		ExpiresIn:    int(ts.ExpiresIn),
		Scope:        ts.Scope,
	}, nil
}

// GoogleTokenRefresher wraps a GoogleOAuthClient as a provider-neutral TokenRefresher.
func GoogleTokenRefresher(client GoogleOAuthClient) calendarprovider.TokenRefresher {
	if client == nil {
		return nil
	}
	return googleTokenAdapter{client: client}
}

// NewGoogleTokenRefresher wraps a GoogleOAuthClient as a provider-neutral TokenRefresher.
func NewGoogleTokenRefresher(client GoogleOAuthClient) calendarprovider.TokenRefresher {
	return GoogleTokenRefresher(client)
}

type microsoftTokenAdapter struct {
	refresh func(ctx context.Context, refreshToken string) (calendarprovider.TokenSet, error)
}

func (a microsoftTokenAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (calendarprovider.TokenSet, error) {
	return a.refresh(ctx, refreshToken)
}

// MicrosoftTokenRefresher adapts a Microsoft OAuth refresh function.
func MicrosoftTokenRefresher(refresh func(ctx context.Context, refreshToken string) (calendarprovider.TokenSet, error)) calendarprovider.TokenRefresher {
	if refresh == nil {
		return nil
	}
	return microsoftTokenAdapter{refresh: refresh}
}

// CalendarSourcesView is the Donna-DB read model for sources + account sync observability.
type CalendarSourcesView struct {
	Sources  []entity.CalendarSource
	Accounts []entity.ConnectedAccount
	Account  entity.ConnectedAccount // first syncable, back-compat
}

// ListSources returns live calendar sources for the user from Donna DB only.
func (s *CalendarService) ListSources(ctx context.Context, userID uuid.UUID) (CalendarSourcesView, error) {
	if userID == uuid.Nil {
		return CalendarSourcesView{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	sources, err := s.sources.ListByUserID(ctx, userID)
	if err != nil {
		return CalendarSourcesView{}, err
	}
	view := CalendarSourcesView{
		Sources:  sources,
		Accounts: []entity.ConnectedAccount{},
	}
	accounts, err := s.listSyncableAccounts(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalid) {
			return view, nil
		}
		return CalendarSourcesView{}, err
	}
	view.Accounts = accounts
	if len(accounts) > 0 {
		view.Account = accounts[0]
	}
	return view, nil
}

// SyncSources is POST /calendar/sync — full orchestration (sources + events).
func (s *CalendarService) SyncSources(ctx context.Context, userID uuid.UUID) (CalendarPipelineResult, error) {
	return s.SyncPipeline(ctx, userID, constant.CalendarSyncTriggerManual)
}

// SyncSourcesForAccount runs the full pipeline for a connected account (scheduler).
func (s *CalendarService) SyncSourcesForAccount(ctx context.Context, accountID uuid.UUID) (CalendarPipelineResult, error) {
	return s.SyncPipelineForAccount(ctx, accountID, constant.CalendarSyncTriggerScheduler)
}

// EnsureFresh runs the full pipeline only when any syncable account is stale.
// Used for app startup and AI workflows — always read Donna DB afterward.
func (s *CalendarService) EnsureFresh(ctx context.Context, userID uuid.UUID, maxAge time.Duration) (CalendarPipelineResult, error) {
	if maxAge <= 0 {
		maxAge = constant.CalendarSyncStaleAfter
	}
	accounts, err := s.listSyncableAccounts(ctx, userID)
	if err != nil {
		return CalendarPipelineResult{}, err
	}
	if len(accounts) == 0 {
		return CalendarPipelineResult{}, fmt.Errorf("%w: connect a calendar provider and grant Calendar access first", apperr.ErrNotFound)
	}

	now := s.now().UTC()
	allFresh := true
	var referenceSyncedAt *time.Time
	var syncStatus string
	for _, account := range accounts {
		if account.LastSyncedAt == nil || now.Sub(account.LastSyncedAt.UTC()) >= maxAge {
			allFresh = false
			break
		}
		if referenceSyncedAt == nil || account.LastSyncedAt.After(*referenceSyncedAt) {
			t := account.LastSyncedAt.UTC()
			referenceSyncedAt = &t
			syncStatus = account.CalendarSyncStatus
		}
	}
	if allFresh && referenceSyncedAt != nil {
		sources, listErr := s.sources.ListByUserID(ctx, userID)
		if listErr != nil {
			return CalendarPipelineResult{}, listErr
		}
		return CalendarPipelineResult{
			Trigger:    constant.CalendarSyncTriggerEnsure,
			Status:     constant.CalendarSyncRunStatusSkipped,
			StartedAt:  *referenceSyncedAt,
			FinishedAt: *referenceSyncedAt,
			Sources:    sources,
			Skipped:    true,
			SyncStatus: syncStatus,
			Failures:   []CalendarSyncFailure{},
		}, nil
	}
	return s.SyncPipeline(ctx, userID, constant.CalendarSyncTriggerEnsure)
}

func (s *CalendarService) providerFor(account entity.ConnectedAccount) (calendarprovider.Provider, error) {
	provider, ok := s.providers[account.Provider]
	if !ok || provider == nil {
		return nil, fmt.Errorf("%w: calendar provider %q is not configured", apperr.ErrInvalid, account.Provider)
	}
	return provider, nil
}

func (s *CalendarService) tokenRefresherFor(account entity.ConnectedAccount) (calendarprovider.TokenRefresher, error) {
	refresher, ok := s.tokens[account.Provider]
	if !ok || refresher == nil {
		return nil, fmt.Errorf("%w: calendar token refresher for %q is not configured", apperr.ErrInvalid, account.Provider)
	}
	return refresher, nil
}

func (s *CalendarService) listSyncableAccounts(ctx context.Context, userID uuid.UUID) ([]entity.ConnectedAccount, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if len(s.providers) == 0 {
		return nil, fmt.Errorf("%w: calendar is not configured", apperr.ErrInvalid)
	}
	all, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.ConnectedAccount, 0, len(all))
	for _, account := range all {
		if account.Status != constant.ConnectedAccountStatusActive {
			continue
		}
		// Google may omit calendar from the token `scope` string on incremental
		// grants. Accounts that already synced calendars are still syncable.
		if !hasCalendarScope(account.Scopes) && !calendarAccessProven(account) {
			continue
		}
		if _, ok := s.providers[account.Provider]; !ok {
			continue
		}
		out = append(out, account)
	}
	return out, nil
}

func (s *CalendarService) requireAnyCalendarAccount(ctx context.Context, userID uuid.UUID) (entity.ConnectedAccount, error) {
	accounts, err := s.listSyncableAccounts(ctx, userID)
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	if len(accounts) == 0 {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: connect a calendar provider and grant Calendar access first", apperr.ErrNotFound)
	}
	return accounts[0], nil
}

func (s *CalendarService) syncAccount(ctx context.Context, account entity.ConnectedAccount, forceFull bool) (result CalendarSyncResult, err error) {
	started := s.now().UTC()
	_, _ = s.accounts.MarkCalendarSyncRunning(ctx, account.ID, constant.CalendarSyncStatusRunning, started)

	var syncCommitted bool
	defer func() {
		// Never overwrite a committed success (e.g. post-commit readback errors).
		if err == nil || syncCommitted {
			return
		}
		failedAt := s.now().UTC()
		duration := int(failedAt.Sub(started).Milliseconds())
		_, _ = s.accounts.RecordCalendarSync(ctx, account.ID, repository.CalendarSyncRecord{
			Status:       constant.CalendarSyncStatusFailed,
			FailedAt:     &failedAt,
			DurationMs:   duration,
			UpdatedAt:    failedAt,
			UpdateCounts: false, // preserve last successful record counts
		})
	}()

	provider, providerErr := s.providerFor(account)
	if providerErr != nil {
		err = providerErr
		return CalendarSyncResult{}, err
	}

	accessToken, tokenErr := s.resolveAccessToken(ctx, account)
	if tokenErr != nil {
		err = tokenErr
		return CalendarSyncResult{}, err
	}

	syncToken := ""
	if !forceFull && account.CalendarListSyncToken != nil {
		syncToken = strings.TrimSpace(*account.CalendarListSyncToken)
	}

	listed, listErr := provider.ListCalendars(ctx, accessToken, calendarprovider.ListCalendarsOptions{SyncToken: syncToken})
	if listErr != nil {
		var gone *calendarprovider.SyncCursorInvalidError
		if errors.As(listErr, &gone) {
			if s.log != nil {
				s.log.Warn(ctx, "calendar list sync token invalid; recovering with full sync",
					constant.LogAttrUserID, account.UserID.String(),
					"provider", account.Provider,
				)
			}
			clearedAt := s.now().UTC()
			if _, clearErr := s.accounts.ClearCalendarListSyncToken(ctx, account.ID, clearedAt); clearErr != nil {
				err = clearErr
				return CalendarSyncResult{}, err
			}
			account.CalendarListSyncToken = nil
			listed, listErr = provider.ListCalendars(ctx, accessToken, calendarprovider.ListCalendarsOptions{})
		}
	}
	if listErr != nil {
		var authErr *calendarprovider.AuthError
		if errors.As(listErr, &authErr) {
			detail := strings.TrimSpace(authErr.Body)
			if detail == "" {
				detail = authErr.Error()
			}
			if s.log != nil {
				s.log.Warn(ctx, "calendar provider list denied",
					constant.LogAttrUserID, account.UserID.String(),
					"provider", account.Provider,
					"status", authErr.Status,
					"provider_error", detail,
				)
			}
			err = fmt.Errorf(
				"%w: calendar provider denied access (%d). Check API enablement and reauthorize. Details: %s",
				apperr.ErrForbidden,
				authErr.Status,
				detail,
			)
			return CalendarSyncResult{}, err
		}
		err = fmt.Errorf("list calendars from provider: %w", listErr)
		return CalendarSyncResult{}, err
	}

	now := s.now().UTC()
	result = CalendarSyncResult{SyncedAt: now, Incremental: listed.Incremental}
	keepIDs := make([]string, 0, len(listed.Calendars))
	deletedIDs := make([]string, 0)

	txErr := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sourcesRepo := s.sources.WithTx(tx)

		for _, remote := range listed.Calendars {
			if remote.Deleted {
				deletedIDs = append(deletedIDs, remote.ID)
				continue
			}
			keepIDs = append(keepIDs, remote.ID)
			_, created, upsertErr := s.upsertSource(ctx, sourcesRepo, account, remote, now)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				result.CreatedCount++
			} else {
				result.UpdatedCount++
			}
		}

		if listed.Incremental {
			removed, delErr := sourcesRepo.SoftDeleteByProviderIDs(ctx, account.ID, deletedIDs, now)
			if delErr != nil {
				return delErr
			}
			result.RemovedCount = int(removed)
		} else {
			removed, delErr := sourcesRepo.SoftDeleteMissing(ctx, account.ID, keepIDs, now)
			if delErr != nil {
				return delErr
			}
			result.RemovedCount = int(removed)
		}
		return nil
	})
	if txErr != nil {
		err = txErr
		return CalendarSyncResult{}, err
	}

	finished := s.now().UTC()
	result.DurationMs = int(finished.Sub(started).Milliseconds())
	result.SyncStatus = constant.CalendarSyncStatusSucceeded

	var nextToken *string
	if listed.NextSyncToken != "" {
		tok := listed.NextSyncToken
		nextToken = &tok
	} else if account.CalendarListSyncToken != nil && listed.Incremental {
		nextToken = account.CalendarListSyncToken
	}

	if _, recErr := s.accounts.RecordCalendarSync(ctx, account.ID, repository.CalendarSyncRecord{
		Status:        constant.CalendarSyncStatusSucceeded,
		SuccessfulAt:  &finished,
		DurationMs:    result.DurationMs,
		CreatedCount:  result.CreatedCount,
		UpdatedCount:  result.UpdatedCount,
		DeletedCount:  result.RemovedCount,
		ListSyncToken: nextToken,
		UpdatedAt:     finished,
		UpdateCounts:  true,
	}); recErr != nil {
		err = recErr
		return CalendarSyncResult{}, err
	}
	syncCommitted = true
	if nextToken != nil {
		account.CalendarListSyncToken = nextToken
	}

	if jobErr := s.ensureBackgroundSyncJob(ctx, account, finished, ensureJobOpts{}); jobErr != nil {
		if s.log != nil {
			s.log.Warn(ctx, "failed to ensure calendar sync scheduler job", constant.LogAttrError, jobErr)
		}
	}

	live, listErr := s.sources.ListByConnectedAccountID(ctx, account.ID)
	if listErr != nil {
		// Sync is committed; return counts without failing the request.
		if s.log != nil {
			s.log.Warn(ctx, "calendar sync committed but source readback failed", constant.LogAttrError, listErr)
		}
		return result, nil
	}
	result.Sources = live

	if s.log != nil {
		s.log.Info(ctx, "calendar sources synced",
			constant.LogAttrUserID, account.UserID.String(),
			"provider", account.Provider,
			"incremental", result.Incremental,
			"created", result.CreatedCount,
			"updated", result.UpdatedCount,
			"removed", result.RemovedCount,
			"duration_ms", result.DurationMs,
		)
	}
	return result, nil
}

type ensureJobOpts struct {
	Payload   []byte
	Immediate bool // run_at = now (OAuth bootstrap); otherwise now + payload interval
}

func (s *CalendarService) ensureBackgroundSyncJob(ctx context.Context, account entity.ConnectedAccount, now time.Time, opts ensureJobOpts) error {
	if s.jobs == nil {
		return nil
	}

	payload := opts.Payload
	existing, getErr := s.jobs.GetPendingByTypeAndAccount(ctx, constant.SchedulerJobTypeCalendarSync, account.ID)

	switch {
	case getErr == nil:
		if len(payload) == 0 {
			payload = existing.Payload
		}
		intervalMinutes := repository.DecodeCalendarSyncPayload(payload).IntervalMinutes
		runAt := now.Add(time.Duration(intervalMinutes) * time.Minute)
		if opts.Immediate {
			runAt = now
		}
		_, err := s.jobs.ReschedulePending(ctx, existing.ID, runAt, now)
		return err

	case errors.Is(getErr, apperr.ErrNotFound):
		intervalMinutes := constant.CalendarSyncIntervalMinutes
		if len(payload) == 0 {
			encoded, encErr := repository.EncodeCalendarSyncPayload(intervalMinutes)
			if encErr != nil {
				return encErr
			}
			payload = encoded
		} else {
			intervalMinutes = repository.DecodeCalendarSyncPayload(payload).IntervalMinutes
		}
		runAt := now.Add(time.Duration(intervalMinutes) * time.Minute)
		if opts.Immediate {
			runAt = now
		}
		id, idErr := idgen.NewUUIDv7()
		if idErr != nil {
			return idErr
		}
		accountID := account.ID
		_, createErr := s.jobs.Create(ctx, entity.SchedulerJob{
			ID:                 id,
			PublicID:           idgen.PublicID(constant.PublicIDPrefixSchedulerJob, id),
			UserID:             account.UserID,
			JobType:            constant.SchedulerJobTypeCalendarSync,
			Status:             constant.SchedulerJobStatusPending,
			RunAt:              runAt,
			AttemptCount:       0,
			MaxAttempts:        5,
			Payload:            payload,
			ConnectedAccountID: &accountID,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
		if createErr == nil {
			return nil
		}
		if !errors.Is(createErr, apperr.ErrConflict) {
			return createErr
		}
		// Unique pending index race: another writer won; reschedule that row.
		existing, getErr = s.jobs.GetPendingByTypeAndAccount(ctx, constant.SchedulerJobTypeCalendarSync, account.ID)
		if getErr != nil {
			return createErr
		}
		_, rescheduleErr := s.jobs.ReschedulePending(ctx, existing.ID, runAt, now)
		return rescheduleErr

	default:
		return getErr
	}
}

// EnsureBackgroundSyncJob creates or reschedules the next pending calendar_sync job.
func (s *CalendarService) EnsureBackgroundSyncJob(ctx context.Context, accountID uuid.UUID) error {
	return s.EnsureBackgroundSyncJobWithPayload(ctx, accountID, nil)
}

// EnsureBackgroundSyncJobWithPayload uses interval_minutes from payload when provided.
func (s *CalendarService) EnsureBackgroundSyncJobWithPayload(ctx context.Context, accountID uuid.UUID, payload []byte) error {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	return s.ensureBackgroundSyncJob(ctx, account, s.now().UTC(), ensureJobOpts{Payload: payload})
}

// BootstrapCalendarSyncJob enqueues an immediate pending calendar_sync after any calendar provider connect.
// No-op when the account lacks calendar scope.
func (s *CalendarService) BootstrapCalendarSyncJob(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if !hasCalendarScope(account.Scopes) {
		return nil
	}
	now := s.now().UTC()
	opts := ensureJobOpts{Immediate: true}
	if account.Provider == constant.AuthProviderICS {
		payload, encErr := repository.EncodeCalendarSyncPayload(constant.ICSDefaultSyncIntervalMin)
		if encErr != nil {
			return encErr
		}
		opts.Payload = payload
	}
	return s.ensureBackgroundSyncJob(ctx, account, now, opts)
}

func (s *CalendarService) upsertSource(
	ctx context.Context,
	repo repository.CalendarSourceRepository,
	account entity.ConnectedAccount,
	remote calendarprovider.RemoteCalendar,
	now time.Time,
) (entity.CalendarSource, bool, error) {
	meta, err := marshalSourceMetadata(remote)
	if err != nil {
		return entity.CalendarSource{}, false, err
	}
	var color *string
	if c := strings.TrimSpace(remote.Color); c != "" {
		color = &c
	}
	var tz *string
	if t := strings.TrimSpace(remote.TimeZone); t != "" {
		tz = &t
	}
	var accessRole *string
	if role := strings.TrimSpace(remote.AccessRole); role != "" {
		accessRole = &role
	}
	name := calendarSourceName(account, remote.Name)

	existing, err := repo.GetByAccountAndProviderCalendar(ctx, account.ID, remote.ID)
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		id, idErr := idgen.NewUUIDv7()
		if idErr != nil {
			return entity.CalendarSource{}, false, idErr
		}
		created, cErr := repo.Create(ctx, entity.CalendarSource{
			ID:                  id,
			PublicID:            idgen.PublicID(constant.PublicIDPrefixCalendarSource, id),
			UserID:              account.UserID,
			ConnectedAccountID:  account.ID,
			ProviderCalendarID:  remote.ID,
			Name:                name,
			Color:               color,
			IsPrimaryOnProvider: remote.Primary,
			IsWritable:          remote.Writable,
			AccessRole:          accessRole,
			SyncEnabled:         true,
			// sync_cursor / last_synced_at belong to events.list sync, not calendarList.
			Timezone:         tz,
			ProviderMetadata: meta,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
		return created, true, cErr
	case err != nil:
		return entity.CalendarSource{}, false, err
	default:
		existing.Name = name
		existing.Color = color
		existing.IsPrimaryOnProvider = remote.Primary
		existing.IsWritable = remote.Writable
		existing.AccessRole = accessRole
		existing.Timezone = tz
		existing.ProviderMetadata = meta
		existing.UpdatedAt = now
		updated, uErr := repo.UpdateFromSync(ctx, existing)
		return updated, false, uErr
	}
}

// calendarSourceName prefers the ICS feed label the user chose over generic
// X-WR-CALNAME values like "Calendar" from the remote ICS body.
func calendarSourceName(account entity.ConnectedAccount, remoteName string) string {
	if account.Provider == constant.AuthProviderICS && account.DisplayName != nil {
		if name := strings.TrimSpace(*account.DisplayName); name != "" {
			return name
		}
	}
	if name := strings.TrimSpace(remoteName); name != "" {
		return name
	}
	return "Calendar"
}

func (s *CalendarService) resolveAccessToken(ctx context.Context, account entity.ConnectedAccount) (string, error) {
	secret, err := s.secrets.GetByRef(ctx, account.CredentialsRef)
	if err != nil {
		return "", fmt.Errorf("%w: credentials unavailable", apperr.ErrForbidden)
	}
	plain, err := seal.Decrypt(s.sealKey, secret.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("decrypt credentials: %w", err)
	}
	var tokens sealedTokens
	if err := json.Unmarshal(plain, &tokens); err != nil {
		return "", fmt.Errorf("decode credentials: %w", err)
	}
	if tokens.AccessToken == "" {
		return "", fmt.Errorf("%w: calendar credentials incomplete; sign in again", apperr.ErrForbidden)
	}
	// ICS (and similar) stores a non-expiring opaque credential with no refresh token.
	if tokens.RefreshToken == "" {
		if _, err := s.tokenRefresherFor(account); err != nil {
			return tokens.AccessToken, nil
		}
		return "", fmt.Errorf("%w: calendar credentials incomplete; sign in again", apperr.ErrForbidden)
	}

	now := s.now().UTC()
	needsRefresh := tokens.ExpiryUnix > 0 && now.Unix() >= (tokens.ExpiryUnix-60)
	if !needsRefresh && account.TokenExpiresAt != nil && !account.TokenExpiresAt.After(now.Add(60*time.Second)) {
		needsRefresh = true
	}
	if !needsRefresh {
		return tokens.AccessToken, nil
	}

	refresher, err := s.tokenRefresherFor(account)
	if err != nil {
		return "", err
	}
	refreshed, err := refresher.RefreshAccessToken(ctx, tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("%w: calendar token refresh failed; sign in again", apperr.ErrForbidden)
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tokens.RefreshToken
	}
	if refreshed.Scope == "" {
		refreshed.Scope = tokens.Scope
	}

	ciphertext, expiresAt, scopes, err := s.sealRefreshedTokens(refreshed, account.Provider, now)
	if err != nil {
		return "", err
	}
	if _, err := s.secrets.UpdateCiphertext(ctx, account.CredentialsRef, ciphertext, now); err != nil {
		return "", err
	}
	if _, err := s.accounts.UpdateCredentials(ctx, account.ID, account.CredentialsRef, expiresAt, scopes, constant.ConnectedAccountStatusActive, now); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func (s *CalendarService) sealRefreshedTokens(tokenSet calendarprovider.TokenSet, provider string, now time.Time) ([]byte, *time.Time, []string, error) {
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
		scopes = defaultCalendarScopes(provider)
	}
	return ciphertext, expiresAt, scopes, nil
}

func defaultCalendarScopes(provider string) []string {
	switch provider {
	case constant.AuthProviderMicrosoft:
		return []string{constant.MicrosoftScopeCalendarsReadWrite, constant.MicrosoftScopeOfflineAccess}
	case constant.AuthProviderICS:
		return []string{constant.ICSScopeCalendar}
	default:
		return []string{constant.GoogleScopeCalendar}
	}
}

func hasCalendarScope(scopes []string) bool {
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if scope == constant.ICSScopeCalendar {
			return true
		}
		if scope == constant.GoogleScopeCalendar ||
			scope == "https://www.googleapis.com/auth/calendar.readonly" ||
			strings.HasSuffix(scope, "/auth/calendar") ||
			strings.HasSuffix(scope, "/auth/calendar.readonly") {
			return true
		}
		if scope == constant.MicrosoftScopeCalendarsRead ||
			scope == constant.MicrosoftScopeCalendarsReadWrite ||
			strings.EqualFold(scope, "https://graph.microsoft.com/Calendars.Read") ||
			strings.EqualFold(scope, "https://graph.microsoft.com/Calendars.ReadWrite") ||
			strings.HasSuffix(scope, "/Calendars.Read") ||
			strings.HasSuffix(scope, "/Calendars.ReadWrite") {
			return true
		}
	}
	return false
}

func calendarAccessProven(account entity.ConnectedAccount) bool {
	if account.LastSyncedAt != nil {
		return true
	}
	switch account.CalendarSyncStatus {
	case constant.CalendarSyncStatusSucceeded, constant.CalendarSyncStatusRunning:
		return true
	default:
		return false
	}
}

func marshalSourceMetadata(remote calendarprovider.RemoteCalendar) ([]byte, error) {
	payload := map[string]any{
		"etag":     remote.ETag,
		"hidden":   remote.Hidden,
		"selected": remote.Selected,
	}
	if remote.Description != "" {
		payload["description"] = remote.Description
	}
	if len(remote.Raw) > 0 {
		for k, v := range remote.Raw {
			if _, exists := payload[k]; !exists {
				payload[k] = v
			}
		}
	}
	return json.Marshal(payload)
}
