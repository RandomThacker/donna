package business

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/google/uuid"
)

// CalendarSyncFailure records a per-calendar error during the events phase.
type CalendarSyncFailure struct {
	CalendarSourceID   string `json:"calendar_source_id"`
	ProviderCalendarID string `json:"provider_calendar_id"`
	Name               string `json:"name"`
	Stage              string `json:"stage"`
	Error              string `json:"error"`
}

// CalendarPipelineResult is the consolidated sources+events sync outcome.
type CalendarPipelineResult struct {
	RunID              uuid.UUID
	Trigger            string
	Status             string
	StartedAt          time.Time
	FinishedAt         time.Time
	DurationMs         int
	CalendarsProcessed int
	SourcesCreated     int
	SourcesUpdated     int
	SourcesDeleted     int
	EventsCreated      int
	EventsUpdated      int
	EventsDeleted      int
	Failures           []CalendarSyncFailure
	Sources            []entity.CalendarSource
	Incremental        bool
	Skipped            bool
	SyncStatus         string
}

// SyncPipeline runs the full calendar sync orchestration for a user.
// Steps: sync sources → sync events for each enabled calendar → aggregate + persist.
// Per-calendar event failures do not abort the run.
func (s *CalendarService) SyncPipeline(ctx context.Context, userID uuid.UUID, trigger string) (CalendarPipelineResult, error) {
	account, err := s.requireGoogleAccount(ctx, userID)
	if err != nil {
		return CalendarPipelineResult{}, err
	}
	return s.runPipeline(ctx, account, trigger)
}

// SyncPipelineForAccount runs the full pipeline for a connected account (scheduler).
func (s *CalendarService) SyncPipelineForAccount(ctx context.Context, accountID uuid.UUID, trigger string) (CalendarPipelineResult, error) {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return CalendarPipelineResult{}, err
	}
	return s.runPipeline(ctx, account, trigger)
}

func (s *CalendarService) runPipeline(ctx context.Context, account entity.ConnectedAccount, trigger string) (result CalendarPipelineResult, err error) {
	if trigger == "" {
		trigger = constant.CalendarSyncTriggerManual
	}
	started := s.now().UTC()
	result = CalendarPipelineResult{
		Trigger:    trigger,
		StartedAt:  started,
		Status:     constant.CalendarSyncRunStatusRunning,
		SyncStatus: constant.CalendarSyncStatusRunning,
		Failures:   make([]CalendarSyncFailure, 0),
	}

	run, runErr := s.beginSyncRun(ctx, account, trigger, started)
	if runErr != nil {
		if s.log != nil {
			s.log.Warn(ctx, "calendar sync run create failed", constant.LogAttrError, runErr)
		}
	} else if run.ID != uuid.Nil {
		result.RunID = run.ID
	}

	if s.log != nil {
		s.log.Info(ctx, "calendar sync pipeline started",
			constant.LogAttrUserID, account.UserID.String(),
			"trigger", trigger,
			"run_id", result.RunID.String(),
			"connected_account_id", account.ID.String(),
		)
	}

	sourcesResult, sourcesErr := s.syncAccount(ctx, account, false)
	if sourcesErr != nil {
		result.Status = constant.CalendarSyncRunStatusFailed
		result.SyncStatus = constant.CalendarSyncStatusFailed
		result.Failures = append(result.Failures, CalendarSyncFailure{
			Stage: "sources",
			Error: sourcesErr.Error(),
		})
		s.finishPipeline(ctx, account, &result, started)
		return result, sourcesErr
	}

	result.SourcesCreated = sourcesResult.CreatedCount
	result.SourcesUpdated = sourcesResult.UpdatedCount
	result.SourcesDeleted = sourcesResult.RemovedCount
	result.Incremental = sourcesResult.Incremental
	result.Sources = sourcesResult.Sources

	if s.log != nil {
		s.log.Info(ctx, "calendar sources phase complete",
			constant.LogAttrUserID, account.UserID.String(),
			"created", result.SourcesCreated,
			"updated", result.SourcesUpdated,
			"deleted", result.SourcesDeleted,
			"incremental", result.Incremental,
		)
	}

	// Re-list enabled sources after sources sync (captures added/removed calendars).
	liveSources, listErr := s.sources.ListByUserID(ctx, account.UserID)
	if listErr != nil {
		result.Status = constant.CalendarSyncRunStatusFailed
		result.SyncStatus = constant.CalendarSyncStatusFailed
		result.Failures = append(result.Failures, CalendarSyncFailure{
			Stage: "sources",
			Error: listErr.Error(),
		})
		s.finishPipeline(ctx, account, &result, started)
		return result, listErr
	}
	result.Sources = liveSources

	if s.events != nil {
		accessToken, tokenErr := s.resolveAccessToken(ctx, account)
		if tokenErr != nil {
			result.Status = constant.CalendarSyncRunStatusPartial
			result.SyncStatus = constant.CalendarSyncStatusFailed
			result.Failures = append(result.Failures, CalendarSyncFailure{
				Stage: "events",
				Error: tokenErr.Error(),
			})
			s.finishPipeline(ctx, account, &result, started)
			return result, nil
		}

		for _, source := range liveSources {
			if !source.SyncEnabled || source.DeletedAt != nil {
				continue
			}
			result.CalendarsProcessed++
			partial, syncErr := s.syncSourceEvents(ctx, account, source, accessToken)
			if syncErr != nil {
				failure := CalendarSyncFailure{
					CalendarSourceID:   source.ID.String(),
					ProviderCalendarID: source.ProviderCalendarID,
					Name:               source.Name,
					Stage:              "events",
					Error:              syncErr.Error(),
				}
				result.Failures = append(result.Failures, failure)
				if s.log != nil {
					s.log.Warn(ctx, "calendar events phase failed for source",
						constant.LogAttrUserID, account.UserID.String(),
						"calendar_source_id", source.ID.String(),
						"provider_calendar_id", source.ProviderCalendarID,
						"name", source.Name,
						constant.LogAttrError, syncErr,
					)
				}
				continue
			}
			result.EventsCreated += partial.CreatedCount
			result.EventsUpdated += partial.UpdatedCount
			result.EventsDeleted += partial.RemovedCount
			if s.log != nil {
				s.log.Info(ctx, "calendar events phase complete for source",
					constant.LogAttrUserID, account.UserID.String(),
					"calendar_source_id", source.ID.String(),
					"provider_calendar_id", source.ProviderCalendarID,
					"created", partial.CreatedCount,
					"updated", partial.UpdatedCount,
					"deleted", partial.RemovedCount,
				)
			}
		}
	}

	if len(result.Failures) == 0 {
		result.Status = constant.CalendarSyncRunStatusSucceeded
		result.SyncStatus = constant.CalendarSyncStatusSucceeded
	} else {
		result.Status = constant.CalendarSyncRunStatusPartial
		result.SyncStatus = constant.CalendarSyncStatusSucceeded
	}

	s.finishPipeline(ctx, account, &result, started)
	return result, nil
}

func (s *CalendarService) beginSyncRun(
	ctx context.Context,
	account entity.ConnectedAccount,
	trigger string,
	started time.Time,
) (entity.CalendarSyncRun, error) {
	if s.syncRuns == nil {
		return entity.CalendarSyncRun{}, nil
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.CalendarSyncRun{}, err
	}
	return s.syncRuns.Create(ctx, entity.CalendarSyncRun{
		ID:                 id,
		PublicID:           idgen.PublicID(constant.PublicIDPrefixCalendarSyncRun, id),
		UserID:             account.UserID,
		ConnectedAccountID: account.ID,
		Trigger:            trigger,
		Status:             constant.CalendarSyncRunStatusRunning,
		StartedAt:          started,
		Failures:           []byte("[]"),
		CreatedAt:          started,
	})
}

func (s *CalendarService) finishPipeline(
	ctx context.Context,
	account entity.ConnectedAccount,
	result *CalendarPipelineResult,
	started time.Time,
) {
	finished := s.now().UTC()
	result.FinishedAt = finished
	result.DurationMs = int(finished.Sub(started).Milliseconds())

	failuresJSON, _ := json.Marshal(result.Failures)
	if len(result.Failures) == 0 {
		failuresJSON = []byte("[]")
	}

	if s.syncRuns != nil && result.RunID != uuid.Nil {
		duration := result.DurationMs
		finishedAt := finished
		_, _ = s.syncRuns.Finish(ctx, entity.CalendarSyncRun{
			ID:                 result.RunID,
			Status:             result.Status,
			FinishedAt:         &finishedAt,
			DurationMs:         &duration,
			CalendarsProcessed: result.CalendarsProcessed,
			SourcesCreated:     result.SourcesCreated,
			SourcesUpdated:     result.SourcesUpdated,
			SourcesDeleted:     result.SourcesDeleted,
			EventsCreated:      result.EventsCreated,
			EventsUpdated:      result.EventsUpdated,
			EventsDeleted:      result.EventsDeleted,
			Failures:           failuresJSON,
		})
	}

	if s.log != nil {
		s.log.Info(ctx, "calendar sync pipeline finished",
			constant.LogAttrUserID, account.UserID.String(),
			"trigger", result.Trigger,
			"run_id", result.RunID.String(),
			"status", result.Status,
			"duration_ms", result.DurationMs,
			"calendars_processed", result.CalendarsProcessed,
			"sources_created", result.SourcesCreated,
			"sources_updated", result.SourcesUpdated,
			"sources_deleted", result.SourcesDeleted,
			"events_created", result.EventsCreated,
			"events_updated", result.EventsUpdated,
			"events_deleted", result.EventsDeleted,
			"failures", len(result.Failures),
		)
	}
}
