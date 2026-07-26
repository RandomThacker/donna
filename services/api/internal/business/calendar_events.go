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
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListEvents returns live events overlapping [from, to) from Donna DB only.
func (s *CalendarService) ListEvents(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.CalendarEvent, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if s.events == nil {
		return nil, fmt.Errorf("%w: calendar events are not configured", apperr.ErrInvalid)
	}
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("%w: from and to are required", apperr.ErrValidation)
	}
	if !to.After(from) {
		return nil, fmt.Errorf("%w: to must be after from", apperr.ErrValidation)
	}
	return s.events.ListByUserInRange(ctx, userID, from.UTC(), to.UTC())
}

// SyncEvents syncs events for every active sync_enabled calendar source across all syncable accounts.
func (s *CalendarService) SyncEvents(ctx context.Context, userID uuid.UUID) (CalendarEventSyncResult, error) {
	started := s.now().UTC()
	accounts, err := s.listSyncableAccounts(ctx, userID)
	if err != nil {
		return CalendarEventSyncResult{}, err
	}
	if len(accounts) == 0 {
		return CalendarEventSyncResult{}, fmt.Errorf("%w: connect a calendar provider and grant Calendar access first", apperr.ErrNotFound)
	}
	if s.events == nil {
		return CalendarEventSyncResult{}, fmt.Errorf("%w: calendar events are not configured", apperr.ErrInvalid)
	}

	result := CalendarEventSyncResult{SyncedAt: started}
	var failures []CalendarSyncFailure

	for _, account := range accounts {
		accessToken, tokenErr := s.resolveAccessToken(ctx, account)
		if tokenErr != nil {
			failures = append(failures, CalendarSyncFailure{
				Stage: "events",
				Error: tokenErr.Error(),
			})
			if s.log != nil {
				s.log.Warn(ctx, "calendar events sync token resolve failed",
					constant.LogAttrUserID, userID.String(),
					"provider", account.Provider,
					constant.LogAttrError, tokenErr,
				)
			}
			continue
		}

		sources, listErr := s.sources.ListByConnectedAccountID(ctx, account.ID)
		if listErr != nil {
			return CalendarEventSyncResult{}, listErr
		}

		for _, source := range sources {
			if source.ConnectedAccountID != account.ID {
				continue
			}
			if !source.SyncEnabled || source.DeletedAt != nil {
				continue
			}
			result.SourceCount++
			partial, syncErr := s.syncSourceEvents(ctx, account, source, accessToken)
			if syncErr != nil {
				failures = append(failures, CalendarSyncFailure{
					CalendarSourceID:   source.ID.String(),
					ProviderCalendarID: source.ProviderCalendarID,
					Name:               source.Name,
					Stage:              "events",
					Error:              syncErr.Error(),
				})
				if s.log != nil {
					s.log.Warn(ctx, "calendar events sync failed for source",
						constant.LogAttrUserID, userID.String(),
						"calendar_source_id", source.ID.String(),
						constant.LogAttrError, syncErr,
					)
				}
				continue
			}
			result.CreatedCount += partial.CreatedCount
			result.UpdatedCount += partial.UpdatedCount
			result.RemovedCount += partial.RemovedCount
		}
	}
	_ = failures

	finished := s.now().UTC()
	result.DurationMs = int(finished.Sub(started).Milliseconds())
	result.SyncedAt = finished

	// Default list window for response payload (same as GET without params).
	from := finished.Add(-7 * 24 * time.Hour)
	to := finished.Add(30 * 24 * time.Hour)
	live, listErr := s.events.ListByUserInRange(ctx, userID, from, to)
	if listErr != nil {
		if s.log != nil {
			s.log.Warn(ctx, "events sync committed but readback failed", constant.LogAttrError, listErr)
		}
	} else {
		result.Events = live
	}

	if s.log != nil {
		s.log.Info(ctx, "calendar events synced",
			constant.LogAttrUserID, userID.String(),
			"sources", result.SourceCount,
			"created", result.CreatedCount,
			"updated", result.UpdatedCount,
			"removed", result.RemovedCount,
			"duration_ms", result.DurationMs,
		)
	}
	return result, nil
}

type partialCounts struct {
	CreatedCount int
	UpdatedCount int
	RemovedCount int
}

func (s *CalendarService) syncSourceEvents(
	ctx context.Context,
	account entity.ConnectedAccount,
	source entity.CalendarSource,
	accessToken string,
) (partialCounts, error) {
	var out partialCounts
	now := s.now().UTC()

	provider, err := s.providerFor(account)
	if err != nil {
		return out, err
	}

	syncToken := ""
	if source.SyncCursor != nil {
		syncToken = strings.TrimSpace(*source.SyncCursor)
	}

	listed, listErr := provider.ListEvents(ctx, accessToken, source.ProviderCalendarID, calendarprovider.ListEventsOptions{
		SyncToken: syncToken,
		TimeMin:   now.Add(-constant.CalendarEventSyncLookback),
		TimeMax:   now.Add(constant.CalendarEventSyncLookahead),
	})
	if listErr != nil {
		var gone *calendarprovider.SyncCursorInvalidError
		if errors.As(listErr, &gone) {
			if s.log != nil {
				s.log.Warn(ctx, "calendar events sync token invalid; recovering with full sync",
					constant.LogAttrUserID, account.UserID.String(),
					"calendar_source_id", source.ID.String(),
					"provider", account.Provider,
				)
			}
			if _, clearErr := s.sources.ClearEventSyncCursor(ctx, source.ID, now); clearErr != nil {
				return out, clearErr
			}
			source.SyncCursor = nil
			listed, listErr = provider.ListEvents(ctx, accessToken, source.ProviderCalendarID, calendarprovider.ListEventsOptions{
				TimeMin: now.Add(-constant.CalendarEventSyncLookback),
				TimeMax: now.Add(constant.CalendarEventSyncLookahead),
			})
		}
	}
	if listErr != nil {
		var authErr *calendarprovider.AuthError
		if errors.As(listErr, &authErr) {
			return out, fmt.Errorf("%w: calendar provider events denied (%d): %s", apperr.ErrForbidden, authErr.Status, authErr.Body)
		}
		return out, fmt.Errorf("list calendar events from provider: %w", listErr)
	}

	txErr := s.tx.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		eventsRepo := s.events.WithTx(tx)
		keepIDs := make([]string, 0, len(listed.Events))
		for _, remote := range listed.Events {
			if remote.ID != "" {
				keepIDs = append(keepIDs, remote.ID)
			}
			if remote.Deleted || remote.Status == constant.CalendarEventStatusCancelled {
				_, delErr := eventsRepo.SoftDeleteByProviderEventID(ctx, source.ID, remote.ID, now)
				if delErr != nil && !errors.Is(delErr, apperr.ErrNotFound) {
					return delErr
				}
				if delErr == nil {
					out.RemovedCount++
				}
				continue
			}
			_, created, upsertErr := s.upsertEvent(ctx, eventsRepo, source, remote, now)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				out.CreatedCount++
			} else {
				out.UpdatedCount++
			}
		}
		// ReplaceAll providers (ICS) return the full feed truth — soft-delete missing UIDs.
		// Windowed OAuth sync must not do this; deletes arrive as cancelled/deleted rows.
		if listed.ReplaceAll {
			removed, delErr := eventsRepo.SoftDeleteMissing(ctx, source.ID, keepIDs, now)
			if delErr != nil {
				return delErr
			}
			out.RemovedCount += int(removed)
		}
		return nil
	})
	if txErr != nil {
		return out, txErr
	}

	var nextToken *string
	if listed.NextSyncToken != "" {
		tok := listed.NextSyncToken
		nextToken = &tok
	} else if source.SyncCursor != nil && listed.Incremental {
		nextToken = source.SyncCursor
	}
	if _, err := s.sources.UpdateEventSyncState(ctx, source.ID, nextToken, now, now); err != nil {
		return out, err
	}
	return out, nil
}

func (s *CalendarService) upsertEvent(
	ctx context.Context,
	repo repository.CalendarEventRepository,
	source entity.CalendarSource,
	remote calendarprovider.RemoteEvent,
	now time.Time,
) (entity.CalendarEvent, bool, error) {
	mapped, err := mapRemoteEventEntity(source, remote, now)
	if err != nil {
		return entity.CalendarEvent{}, false, err
	}

	existing, err := repo.GetBySourceAndProviderEvent(ctx, source.ID, remote.ID)
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		id, idErr := idgen.NewUUIDv7()
		if idErr != nil {
			return entity.CalendarEvent{}, false, idErr
		}
		mapped.ID = id
		mapped.PublicID = idgen.PublicID(constant.PublicIDPrefixCalendarEvent, id)
		mapped.CreatedAt = now
		if parentID, resolveErr := s.resolveRecurringParent(ctx, repo, source.ID, remote.RecurringEventID); resolveErr == nil {
			mapped.RecurringEventID = parentID
		}
		created, cErr := repo.Create(ctx, mapped)
		return created, true, cErr
	case err != nil:
		return entity.CalendarEvent{}, false, err
	default:
		mapped.ID = existing.ID
		mapped.PublicID = existing.PublicID
		mapped.CreatedAt = existing.CreatedAt
		mapped.RecurringEventID = existing.RecurringEventID
		if parentID, resolveErr := s.resolveRecurringParent(ctx, repo, source.ID, remote.RecurringEventID); resolveErr == nil {
			mapped.RecurringEventID = parentID
		}
		updated, uErr := repo.UpdateFromSync(ctx, mapped)
		return updated, false, uErr
	}
}

func (s *CalendarService) resolveRecurringParent(
	ctx context.Context,
	repo repository.CalendarEventRepository,
	sourceID uuid.UUID,
	providerRecurringID string,
) (*uuid.UUID, error) {
	providerRecurringID = strings.TrimSpace(providerRecurringID)
	if providerRecurringID == "" {
		return nil, nil
	}
	parent, err := repo.GetBySourceAndProviderEvent(ctx, sourceID, providerRecurringID)
	if err != nil {
		return nil, err
	}
	if parent.DeletedAt != nil {
		return nil, nil
	}
	id := parent.ID
	return &id, nil
}

func mapRemoteEventEntity(source entity.CalendarSource, remote calendarprovider.RemoteEvent, now time.Time) (entity.CalendarEvent, error) {
	status := normalizeEventStatus(remote.Status)
	attendees, err := json.Marshal(remote.Attendees)
	if err != nil {
		return entity.CalendarEvent{}, err
	}
	if len(remote.Attendees) == 0 {
		attendees = []byte("[]")
	}

	var organizer []byte
	if remote.OrganizerEmail != "" || remote.OrganizerDisplayName != "" {
		organizer, err = json.Marshal(map[string]any{
			"email":        remote.OrganizerEmail,
			"display_name": remote.OrganizerDisplayName,
			"self":         remote.OrganizerSelf,
		})
		if err != nil {
			return entity.CalendarEvent{}, err
		}
	}

	raw := remote.Raw
	if meetingURL := strings.TrimSpace(remote.OnlineMeetingURL); meetingURL != "" {
		if raw == nil {
			raw = map[string]any{}
		} else {
			cloned := make(map[string]any, len(raw)+1)
			for k, v := range raw {
				cloned[k] = v
			}
			raw = cloned
		}
		if _, exists := raw["online_meeting_url"]; !exists {
			raw["online_meeting_url"] = meetingURL
		}
	}

	payload, err := json.Marshal(raw)
	if err != nil {
		return entity.CalendarEvent{}, err
	}
	if len(payload) == 0 || string(payload) == "null" {
		payload = []byte("{}")
	}

	providerID := remote.ID
	event := entity.CalendarEvent{
		UserID:           source.UserID,
		CalendarSourceID: source.ID,
		Title:            remote.Title,
		StartsAt:         remote.StartsAt,
		EndsAt:           remote.EndsAt,
		IsAllDay:         remote.IsAllDay,
		Status:           status,
		AttendeesSummary: attendees,
		OrganizerSummary: organizer,
		ProviderEventID:  &providerID,
		ProviderPayload:  payload,
		Origin:           constant.CalendarEventOriginProviderSync,
		UpdatedAt:        now,
	}
	if d := strings.TrimSpace(remote.Description); d != "" {
		event.Description = &d
	}
	if l := strings.TrimSpace(remote.Location); l != "" {
		event.Location = &l
	} else if meeting := strings.TrimSpace(remote.OnlineMeetingURL); meeting != "" {
		event.Location = &meeting
	}
	if tz := strings.TrimSpace(remote.Timezone); tz != "" {
		event.Timezone = &tz
	}
	if v := normalizeVisibility(remote.Visibility); v != "" {
		event.Visibility = &v
	}
	if e := strings.TrimSpace(remote.ETag); e != "" {
		event.ProviderETag = &e
	}
	if !remote.UpdatedAt.IsZero() {
		u := remote.UpdatedAt
		event.ProviderUpdatedAt = &u
	}
	if rid := strings.TrimSpace(remote.RecurringEventID); rid != "" {
		event.ProviderRecurringEventID = &rid
	}
	if len(remote.Recurrence) > 0 {
		rule := strings.Join(remote.Recurrence, "\n")
		event.RecurrenceRule = &rule
	}
	return event, nil
}

func normalizeEventStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case constant.CalendarEventStatusTentative:
		return constant.CalendarEventStatusTentative
	case constant.CalendarEventStatusCancelled:
		return constant.CalendarEventStatusCancelled
	default:
		return constant.CalendarEventStatusConfirmed
	}
}

func normalizeVisibility(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "default", "private", "public":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}
