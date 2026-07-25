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
	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GoogleCalendarClient lists calendars from Google Calendar API.
type GoogleCalendarClient interface {
	ListCalendars(ctx context.Context, accessToken string, opts googlecalendar.ListOptions) (googlecalendar.ListResult, error)
}

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

// CalendarService syncs and lists calendar sources (Donna DB is the query source of truth).
type CalendarService struct {
	accounts repository.ConnectedAccountRepository
	secrets  repository.CredentialSecretRepository
	sources  repository.CalendarSourceRepository
	jobs     repository.SchedulerJobRepository
	tx       TxRunner
	oauth    GoogleOAuthClient
	calendar GoogleCalendarClient
	sealKey  []byte
	log      *logger.Logger
	now      func() time.Time
}

// CalendarServiceDeps wires CalendarService.
type CalendarServiceDeps struct {
	Accounts repository.ConnectedAccountRepository
	Secrets  repository.CredentialSecretRepository
	Sources  repository.CalendarSourceRepository
	Jobs     repository.SchedulerJobRepository
	Tx       TxRunner
	OAuth    GoogleOAuthClient
	Calendar GoogleCalendarClient
	SealKey  []byte
	Log      *logger.Logger
}

// NewCalendarService constructs a CalendarService.
func NewCalendarService(deps CalendarServiceDeps) *CalendarService {
	return &CalendarService{
		accounts: deps.Accounts,
		secrets:  deps.Secrets,
		sources:  deps.Sources,
		jobs:     deps.Jobs,
		tx:       deps.Tx,
		oauth:    deps.OAuth,
		calendar: deps.Calendar,
		sealKey:  deps.SealKey,
		log:      deps.Log,
		now:      time.Now,
	}
}

// CalendarSourcesView is the Donna-DB read model for sources + account sync observability.
type CalendarSourcesView struct {
	Sources []entity.CalendarSource
	Account entity.ConnectedAccount
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
	view := CalendarSourcesView{Sources: sources}
	account, err := s.accounts.GetByUserAndProvider(ctx, userID, constant.AuthProviderGoogle)
	if err == nil {
		view.Account = account
	} else if !errors.Is(err, apperr.ErrNotFound) {
		return CalendarSourcesView{}, err
	}
	return view, nil
}

// SyncSources performs manual / background sync (incremental when syncToken exists).
func (s *CalendarService) SyncSources(ctx context.Context, userID uuid.UUID) (CalendarSyncResult, error) {
	account, err := s.requireGoogleAccount(ctx, userID)
	if err != nil {
		return CalendarSyncResult{}, err
	}
	return s.syncAccount(ctx, account, false)
}

// SyncSourcesForAccount runs sync for a connected account (scheduler worker).
func (s *CalendarService) SyncSourcesForAccount(ctx context.Context, accountID uuid.UUID) (CalendarSyncResult, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return CalendarSyncResult{}, err
	}
	return s.syncAccount(ctx, account, false)
}

// EnsureFresh syncs only when last successful sync is older than maxAge (default 2m).
// Used for app startup and AI workflows that depend on calendar data — call this before
// any calendar-dependent tool; always read Donna DB afterward, never Google directly.
func (s *CalendarService) EnsureFresh(ctx context.Context, userID uuid.UUID, maxAge time.Duration) (CalendarSyncResult, error) {
	if maxAge <= 0 {
		maxAge = constant.CalendarSyncStaleAfter
	}
	account, err := s.requireGoogleAccount(ctx, userID)
	if err != nil {
		return CalendarSyncResult{}, err
	}
	now := s.now().UTC()
	if account.LastSyncedAt != nil && now.Sub(account.LastSyncedAt.UTC()) < maxAge {
		sources, listErr := s.sources.ListByUserID(ctx, userID)
		if listErr != nil {
			return CalendarSyncResult{}, listErr
		}
		return CalendarSyncResult{
			Sources:    sources,
			SyncedAt:   account.LastSyncedAt.UTC(),
			Skipped:    true,
			SyncStatus: account.CalendarSyncStatus,
		}, nil
	}
	return s.syncAccount(ctx, account, false)
}

func (s *CalendarService) requireGoogleAccount(ctx context.Context, userID uuid.UUID) (entity.ConnectedAccount, error) {
	if userID == uuid.Nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if s.oauth == nil || s.calendar == nil {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google calendar is not configured", apperr.ErrInvalid)
	}
	account, err := s.accounts.GetByUserAndProvider(ctx, userID, constant.AuthProviderGoogle)
	if errors.Is(err, apperr.ErrNotFound) {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: connect Google and grant Calendar access first", apperr.ErrNotFound)
	}
	if err != nil {
		return entity.ConnectedAccount{}, err
	}
	if account.Status != constant.ConnectedAccountStatusActive {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google account needs reauthorization", apperr.ErrForbidden)
	}
	if !hasCalendarScope(account.Scopes) {
		return entity.ConnectedAccount{}, fmt.Errorf("%w: google calendar scope missing; sign in again to grant Calendar access", apperr.ErrForbidden)
	}
	return account, nil
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

	accessToken, tokenErr := s.resolveAccessToken(ctx, account)
	if tokenErr != nil {
		err = tokenErr
		return CalendarSyncResult{}, err
	}

	syncToken := ""
	if !forceFull && account.CalendarListSyncToken != nil {
		syncToken = strings.TrimSpace(*account.CalendarListSyncToken)
	}

	listed, listErr := s.calendar.ListCalendars(ctx, accessToken, googlecalendar.ListOptions{SyncToken: syncToken})
	if listErr != nil {
		var gone *googlecalendar.GoneError
		if errors.As(listErr, &gone) {
			if s.log != nil {
				s.log.Warn(ctx, "calendar list sync token invalid; recovering with full sync",
					constant.LogAttrUserID, account.UserID.String(),
				)
			}
			clearedAt := s.now().UTC()
			if _, clearErr := s.accounts.ClearCalendarListSyncToken(ctx, account.ID, clearedAt); clearErr != nil {
				err = clearErr
				return CalendarSyncResult{}, err
			}
			account.CalendarListSyncToken = nil
			listed, listErr = s.calendar.ListCalendars(ctx, accessToken, googlecalendar.ListOptions{})
		}
	}
	if listErr != nil {
		var authErr *googlecalendar.AuthError
		if errors.As(listErr, &authErr) {
			detail := strings.TrimSpace(authErr.Body)
			if detail == "" {
				detail = authErr.Error()
			}
			if s.log != nil {
				s.log.Warn(ctx, "google calendar list denied",
					constant.LogAttrUserID, account.UserID.String(),
					"status", authErr.Status,
					"google_error", detail,
				)
			}
			err = fmt.Errorf(
				"%w: google calendar API denied access (%d). Enable the Google Calendar API for this Cloud project, then sign in again. Details: %s",
				apperr.ErrForbidden,
				authErr.Status,
				detail,
			)
			return CalendarSyncResult{}, err
		}
		err = fmt.Errorf("list google calendars: %w", listErr)
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

	live, listErr := s.sources.ListByUserID(ctx, account.UserID)
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

// BootstrapCalendarSyncJob enqueues an immediate pending calendar_sync after Google connect.
// No-op when the account lacks calendar scope.
func (s *CalendarService) BootstrapCalendarSyncJob(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if !hasCalendarScope(account.Scopes) {
		return nil
	}
	return s.ensureBackgroundSyncJob(ctx, account, s.now().UTC(), ensureJobOpts{Immediate: true})
}

func (s *CalendarService) upsertSource(
	ctx context.Context,
	repo repository.CalendarSourceRepository,
	account entity.ConnectedAccount,
	remote googlecalendar.RemoteCalendar,
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
	var syncCursor *string
	if e := strings.TrimSpace(remote.ETag); e != "" {
		syncCursor = &e
	}
	var accessRole *string
	if role := strings.TrimSpace(remote.AccessRole); role != "" {
		accessRole = &role
	}
	syncedAt := now

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
			Name:                remote.Name,
			Color:               color,
			IsPrimaryOnProvider: remote.Primary,
			IsWritable:          remote.Writable,
			AccessRole:          accessRole,
			SyncEnabled:         true,
			SyncCursor:          syncCursor,
			LastSyncedAt:        &syncedAt,
			Timezone:            tz,
			ProviderMetadata:    meta,
			CreatedAt:           now,
			UpdatedAt:           now,
		})
		return created, true, cErr
	case err != nil:
		return entity.CalendarSource{}, false, err
	default:
		existing.Name = remote.Name
		existing.Color = color
		existing.IsPrimaryOnProvider = remote.Primary
		existing.IsWritable = remote.Writable
		existing.AccessRole = accessRole
		existing.SyncCursor = syncCursor
		existing.LastSyncedAt = &syncedAt
		existing.Timezone = tz
		existing.ProviderMetadata = meta
		existing.UpdatedAt = now
		updated, uErr := repo.UpdateFromSync(ctx, existing)
		return updated, false, uErr
	}
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
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return "", fmt.Errorf("%w: google credentials incomplete; sign in again", apperr.ErrForbidden)
	}

	now := s.now().UTC()
	needsRefresh := tokens.ExpiryUnix > 0 && now.Unix() >= (tokens.ExpiryUnix-60)
	if !needsRefresh && account.TokenExpiresAt != nil && !account.TokenExpiresAt.After(now.Add(60*time.Second)) {
		needsRefresh = true
	}
	if !needsRefresh {
		return tokens.AccessToken, nil
	}

	refreshed, err := s.oauth.RefreshAccessToken(ctx, tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("%w: google token refresh failed; sign in again", apperr.ErrForbidden)
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tokens.RefreshToken
	}
	if refreshed.Scope == "" {
		refreshed.Scope = tokens.Scope
	}

	ciphertext, expiresAt, scopes, err := s.sealRefreshedTokens(refreshed, now)
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

func (s *CalendarService) sealRefreshedTokens(tokenSet googleoauth.TokenSet, now time.Time) ([]byte, *time.Time, []string, error) {
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
		scopes = []string{constant.GoogleScopeCalendar}
	}
	return ciphertext, expiresAt, scopes, nil
}

func hasCalendarScope(scopes []string) bool {
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == constant.GoogleScopeCalendar ||
			scope == "https://www.googleapis.com/auth/calendar.readonly" ||
			strings.HasSuffix(scope, "/auth/calendar") ||
			strings.HasSuffix(scope, "/auth/calendar.readonly") {
			return true
		}
	}
	return false
}

func marshalSourceMetadata(remote googlecalendar.RemoteCalendar) ([]byte, error) {
	payload := map[string]any{
		"etag":     remote.ETag,
		"hidden":   remote.Hidden,
		"selected": remote.Selected,
	}
	if remote.Description != "" {
		payload["description"] = remote.Description
	}
	return json.Marshal(payload)
}
