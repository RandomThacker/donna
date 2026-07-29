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
	"github.com/RandomThacker/donna/services/api/internal/icscalendar"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ConnectICSRequest creates a universal ICS calendar feed integration.
type ConnectICSRequest struct {
	Name        string
	ICSURL      string
	SyncEnabled *bool
}

// UpdateICSRequest patches an ICS feed integration (never returns the sealed URL).
type UpdateICSRequest struct {
	Name        *string
	ICSURL      *string
	SyncEnabled *bool
}

// ICSIntegrationView is the API projection for an ICS calendar feed.
type ICSIntegrationView struct {
	Account     entity.ConnectedAccount
	SyncEnabled bool
	EventCount  int64
}

// ConnectICS registers a provider-independent ICS feed as a connected account.
func (s *IntegrationService) ConnectICS(ctx context.Context, userID uuid.UUID, req ConnectICSRequest) (ICSIntegrationView, error) {
	if userID == uuid.Nil {
		return ICSIntegrationView{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return ICSIntegrationView{}, fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	feedURL, err := icscalendar.NormalizeFeedURL(req.ICSURL)
	if err != nil {
		return ICSIntegrationView{}, fmt.Errorf("%w: %v", apperr.ErrValidation, err)
	}
	syncEnabled := true
	if req.SyncEnabled != nil {
		syncEnabled = *req.SyncEnabled
	}

	providerAccountID := icscalendar.FeedAccountID(feedURL)
	now := s.now().UTC()

	var account entity.ConnectedAccount
	txErr := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		accounts := s.accounts.WithTx(tx)
		secrets := s.secrets.WithTx(tx)

		existing, getErr := accounts.GetByProviderAccount(ctx, constant.AuthProviderICS, providerAccountID)
		if getErr == nil {
			if existing.UserID != userID {
				return fmt.Errorf("%w: ics feed already linked to another user", apperr.ErrConflict)
			}
			updated, uerr := s.updateICSAccount(ctx, accounts, secrets, existing, name, feedURL, syncEnabled, now)
			if uerr != nil {
				return uerr
			}
			account = updated
			return nil
		}
		if !errors.Is(getErr, apperr.ErrNotFound) {
			return getErr
		}

		created, cerr := s.storeICSAccount(ctx, accounts, secrets, userID, providerAccountID, name, feedURL, syncEnabled, now)
		if cerr != nil {
			return cerr
		}
		account = created
		return nil
	})
	if txErr != nil {
		return ICSIntegrationView{}, txErr
	}

	_ = s.applyICSSourceSyncEnabled(ctx, account.ID, syncEnabled, now)
	s.notifyAccountReady(ctx, account.ID)
	return s.icsView(ctx, account)
}

// ListICS returns ICS feed integrations for the user (URL never exposed).
func (s *IntegrationService) ListICS(ctx context.Context, userID uuid.UUID) ([]ICSIntegrationView, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	accounts, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ICSIntegrationView, 0)
	for _, account := range accounts {
		if account.Provider != constant.AuthProviderICS {
			continue
		}
		view, verr := s.icsView(ctx, account)
		if verr != nil {
			return nil, verr
		}
		out = append(out, view)
	}
	return out, nil
}

// UpdateICS patches name, optional feed URL, and/or sync_enabled.
func (s *IntegrationService) UpdateICS(ctx context.Context, userID, accountID uuid.UUID, req UpdateICSRequest) (ICSIntegrationView, error) {
	if userID == uuid.Nil || accountID == uuid.Nil {
		return ICSIntegrationView{}, fmt.Errorf("%w: user id and account id are required", apperr.ErrValidation)
	}
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return ICSIntegrationView{}, err
	}
	if account.UserID != userID {
		return ICSIntegrationView{}, apperr.ErrForbidden
	}
	if account.Provider != constant.AuthProviderICS {
		return ICSIntegrationView{}, fmt.Errorf("%w: not an ics integration", apperr.ErrInvalid)
	}

	name := ""
	if account.DisplayName != nil {
		name = *account.DisplayName
	}
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
		if name == "" {
			return ICSIntegrationView{}, fmt.Errorf("%w: name is required", apperr.ErrValidation)
		}
	}
	syncEnabled := icsSyncEnabledFromMetadata(account.ProviderMetadata)
	if req.SyncEnabled != nil {
		syncEnabled = *req.SyncEnabled
	}

	feedURL := ""
	if req.ICSURL != nil {
		normalized, nerr := icscalendar.NormalizeFeedURL(*req.ICSURL)
		if nerr != nil {
			return ICSIntegrationView{}, fmt.Errorf("%w: %v", apperr.ErrValidation, nerr)
		}
		feedURL = normalized
		if icscalendar.FeedAccountID(feedURL) != account.ProviderAccountID {
			return ICSIntegrationView{}, fmt.Errorf(
				"%w: changing the calendar url to a different feed is not supported; delete and add a new feed",
				apperr.ErrInvalid,
			)
		}
	}

	now := s.now().UTC()
	updated, err := s.updateICSAccount(ctx, s.accounts, s.secrets, account, name, feedURL, syncEnabled, now)
	if err != nil {
		return ICSIntegrationView{}, err
	}
	_ = s.applyICSSourceSyncEnabled(ctx, updated.ID, syncEnabled, now)
	_ = s.applyICSSourceName(ctx, updated.ID, name, now)
	return s.icsView(ctx, updated)
}

// DeleteICS disconnects an ICS feed and permanently removes its calendar rows.
func (s *IntegrationService) DeleteICS(ctx context.Context, userID, accountID uuid.UUID) error {
	return s.Disconnect(ctx, userID, accountID)
}

// SyncICS runs the calendar pipeline for one ICS feed account.
func (s *IntegrationService) SyncICS(ctx context.Context, userID, accountID uuid.UUID) (ICSIntegrationView, error) {
	if userID == uuid.Nil || accountID == uuid.Nil {
		return ICSIntegrationView{}, fmt.Errorf("%w: user id and account id are required", apperr.ErrValidation)
	}
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return ICSIntegrationView{}, err
	}
	if account.UserID != userID {
		return ICSIntegrationView{}, apperr.ErrForbidden
	}
	if account.Provider != constant.AuthProviderICS {
		return ICSIntegrationView{}, fmt.Errorf("%w: not an ics integration", apperr.ErrInvalid)
	}
	if s.syncAccount == nil {
		return ICSIntegrationView{}, fmt.Errorf("%w: calendar sync is not configured", apperr.ErrInvalid)
	}
	if _, err := s.syncAccount(ctx, accountID); err != nil {
		return ICSIntegrationView{}, err
	}
	fresh, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return ICSIntegrationView{}, err
	}
	return s.icsView(ctx, fresh)
}

func (s *IntegrationService) storeICSAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	userID uuid.UUID,
	providerAccountID, name, feedURL string,
	syncEnabled bool,
	now time.Time,
) (entity.ConnectedAccount, error) {
	ciphertext, err := s.sealICSCredential(feedURL)
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
	displayName := name
	return accounts.Create(ctx, entity.ConnectedAccount{
		ID:                accountID,
		PublicID:          idgen.PublicID(constant.PublicIDPrefixConnectedAccount, accountID),
		UserID:            userID,
		Provider:          constant.AuthProviderICS,
		ProviderAccountID: providerAccountID,
		DisplayName:       &displayName,
		Status:            constant.ConnectedAccountStatusActive,
		Scopes:            []string{constant.ICSScopeCalendar},
		CredentialsRef:    ref,
		ProviderMetadata:  icsProviderMetadata(name, syncEnabled),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *IntegrationService) updateICSAccount(
	ctx context.Context,
	accounts repository.ConnectedAccountRepository,
	secrets repository.CredentialSecretRepository,
	existing entity.ConnectedAccount,
	name, feedURL string,
	syncEnabled bool,
	now time.Time,
) (entity.ConnectedAccount, error) {
	if feedURL != "" {
		ciphertext, err := s.sealICSCredential(feedURL)
		if err != nil {
			return entity.ConnectedAccount{}, err
		}
		_, err = secrets.UpdateCiphertext(ctx, existing.CredentialsRef, ciphertext, now)
		if errors.Is(err, apperr.ErrNotFound) {
			ref, cerr := s.createSecret(ctx, secrets, ciphertext, now)
			if cerr != nil {
				return entity.ConnectedAccount{}, cerr
			}
			if _, uerr := accounts.UpdateCredentials(
				ctx,
				existing.ID,
				ref,
				nil,
				[]string{constant.ICSScopeCalendar},
				constant.ConnectedAccountStatusActive,
				now,
			); uerr != nil {
				return entity.ConnectedAccount{}, uerr
			}
		} else if err != nil {
			return entity.ConnectedAccount{}, err
		} else if _, uerr := accounts.UpdateCredentials(
			ctx,
			existing.ID,
			existing.CredentialsRef,
			nil,
			[]string{constant.ICSScopeCalendar},
			constant.ConnectedAccountStatusActive,
			now,
		); uerr != nil {
			return entity.ConnectedAccount{}, uerr
		}
	}

	displayName := name
	return accounts.UpdateProfile(ctx, existing.ID, &displayName, icsProviderMetadata(name, syncEnabled), now)
}

func (s *IntegrationService) sealICSCredential(feedURL string) ([]byte, error) {
	payload, err := json.Marshal(sealedTokens{
		AccessToken: feedURL,
		TokenType:   "ics_url",
		Scope:       constant.ICSScopeCalendar,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal ics credential: %w", err)
	}
	return seal.Encrypt(s.sealKey, payload)
}

func (s *IntegrationService) applyICSSourceSyncEnabled(ctx context.Context, accountID uuid.UUID, syncEnabled bool, now time.Time) error {
	if s.sources == nil {
		return nil
	}
	_, err := s.sources.UpdateSyncEnabledByAccount(ctx, accountID, syncEnabled, now)
	return err
}

func (s *IntegrationService) applyICSSourceName(ctx context.Context, accountID uuid.UUID, name string, now time.Time) error {
	if s.sources == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	sources, err := s.sources.ListByConnectedAccountID(ctx, accountID)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if source.Name == name {
			continue
		}
		source.Name = name
		source.UpdatedAt = now
		if _, err := s.sources.UpdateFromSync(ctx, source); err != nil {
			return err
		}
	}
	return nil
}

func (s *IntegrationService) icsView(ctx context.Context, account entity.ConnectedAccount) (ICSIntegrationView, error) {
	view := ICSIntegrationView{
		Account:     account,
		SyncEnabled: icsSyncEnabledFromMetadata(account.ProviderMetadata),
	}
	if s.sources != nil {
		sources, err := s.sources.ListByConnectedAccountID(ctx, account.ID)
		if err != nil {
			return ICSIntegrationView{}, err
		}
		if len(sources) > 0 {
			view.SyncEnabled = sources[0].SyncEnabled
		}
	}
	if s.events != nil {
		n, err := s.events.CountLiveByConnectedAccountID(ctx, account.ID)
		if err != nil {
			return ICSIntegrationView{}, err
		}
		view.EventCount = n
	}
	return view, nil
}

func icsProviderMetadata(name string, syncEnabled bool) []byte {
	payload, err := json.Marshal(map[string]any{
		"name":         strings.TrimSpace(name),
		"sync_enabled": syncEnabled,
		"kind":         "ics_feed",
	})
	if err != nil {
		return []byte(`{"kind":"ics_feed","sync_enabled":true}`)
	}
	return payload
}

func icsSyncEnabledFromMetadata(raw []byte) bool {
	if len(raw) == 0 {
		return true
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return true
	}
	v, ok := meta["sync_enabled"]
	if !ok {
		return true
	}
	b, ok := v.(bool)
	if !ok {
		return true
	}
	return b
}
