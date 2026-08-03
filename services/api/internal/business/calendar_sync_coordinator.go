package business

import (
	"context"
	"sync"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarsyncmetrics"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/google/uuid"
)

// CalendarSyncCoordinator is the sole admission point for calendar sync pipelines.
// Handlers and jobs request sync here; the coordinator enforces cooldown, dedupes
// in-flight work, records metrics/logs, and delegates execution to CalendarService.
// Distributed locking can replace the in-process maps later without changing callers.
type CalendarSyncCoordinator struct {
	svc     *CalendarService
	log     *logger.Logger
	metrics *calendarsyncmetrics.Registry

	mu              sync.Mutex
	inflightUser    map[uuid.UUID]*inflightSync
	inflightAccount map[uuid.UUID]*inflightSync
}

type inflightSync struct {
	done   chan struct{}
	result CalendarPipelineResult
	err    error
}

// NewCalendarSyncCoordinator constructs a coordinator.
func NewCalendarSyncCoordinator(svc *CalendarService, log *logger.Logger, metrics *calendarsyncmetrics.Registry) *CalendarSyncCoordinator {
	if metrics == nil {
		metrics = calendarsyncmetrics.Global
	}
	return &CalendarSyncCoordinator{
		svc:             svc,
		log:             log,
		metrics:         metrics,
		inflightUser:    map[uuid.UUID]*inflightSync{},
		inflightAccount: map[uuid.UUID]*inflightSync{},
	}
}

// Status returns the last recorded sync status for a user (in-process).
func (c *CalendarSyncCoordinator) Status(userID uuid.UUID) (calendarsyncmetrics.SyncStatus, bool) {
	return calendarsyncmetrics.StatusFor(userID.String())
}

// SyncUser runs a full pipeline for all of the user's syncable accounts.
// Manual / OAuth / ICS paths use this (or SyncAccount). No freshness cooldown.
func (c *CalendarSyncCoordinator) SyncUser(ctx context.Context, userID uuid.UUID, source, reason string) (CalendarPipelineResult, error) {
	if source == "" {
		source = constant.CalendarSyncTriggerManual
	}
	if reason == "" {
		reason = source
	}
	return c.runUser(ctx, userID, source, reason, false, 0)
}

// SyncAccount runs a full pipeline for one connected account (scheduler / ICS sync-now).
func (c *CalendarSyncCoordinator) SyncAccount(ctx context.Context, accountID uuid.UUID, source, reason string) (CalendarPipelineResult, error) {
	if source == "" {
		source = constant.CalendarSyncTriggerManual
	}
	if reason == "" {
		reason = source
	}
	return c.runAccount(ctx, accountID, source, reason)
}

// EnsureFresh runs the pipeline only when any syncable account's last success is older than maxAge.
func (c *CalendarSyncCoordinator) EnsureFresh(ctx context.Context, userID uuid.UUID, maxAge time.Duration) (CalendarPipelineResult, error) {
	if maxAge <= 0 {
		maxAge = constant.CalendarSyncStaleAfter
	}
	return c.runUser(ctx, userID, constant.CalendarSyncTriggerEnsure, "stale_check", true, maxAge)
}

func (c *CalendarSyncCoordinator) runUser(
	ctx context.Context,
	userID uuid.UUID,
	source, reason string,
	applyCooldown bool,
	maxAge time.Duration,
) (CalendarPipelineResult, error) {
	started := time.Now().UTC()
	c.metrics.IncSyncRequested()

	inflight, leader := c.beginUser(userID)
	if !leader {
		<-inflight.done
		c.logSyncRequested(ctx, source, reason, userID, uuid.Nil, true, false, time.Since(started), inflight.result)
		return inflight.result, inflight.err
	}
	defer c.finishUser(userID, inflight)

	if applyCooldown {
		skipped, result, err := c.tryCooldownSkip(ctx, userID, maxAge)
		if err != nil {
			inflight.err = err
			c.recordOutcome(userID.String(), source, reason, true, true, "", time.Since(started))
			c.logSyncRequested(ctx, source, reason, userID, uuid.Nil, true, true, time.Since(started), result)
			return result, err
		}
		if skipped {
			c.metrics.IncSyncSkipped()
			inflight.result = result
			c.recordOutcome(userID.String(), source, reason, true, true, result.SyncStatus, time.Since(started))
			c.logSyncRequested(ctx, source, reason, userID, uuid.Nil, true, true, time.Since(started), result)
			return result, nil
		}
	}

	result, err := c.svc.SyncPipeline(ctx, userID, source)
	inflight.result = result
	inflight.err = err
	c.observePipeline(result, err)
	c.recordOutcome(userID.String(), source, reason, false, false, result.Status, time.Since(started))
	c.logSyncRequested(ctx, source, reason, userID, uuid.Nil, false, false, time.Since(started), result)
	c.logSyncFinished(ctx, result)
	return result, err
}

func (c *CalendarSyncCoordinator) runAccount(
	ctx context.Context,
	accountID uuid.UUID,
	source, reason string,
) (CalendarPipelineResult, error) {
	started := time.Now().UTC()
	c.metrics.IncSyncRequested()

	inflight, leader := c.beginAccount(accountID)
	if !leader {
		<-inflight.done
		c.logSyncRequested(ctx, source, reason, uuid.Nil, accountID, true, false, time.Since(started), inflight.result)
		return inflight.result, inflight.err
	}
	defer c.finishAccount(accountID, inflight)

	result, err := c.svc.SyncPipelineForAccount(ctx, accountID, source)
	inflight.result = result
	inflight.err = err
	c.observePipeline(result, err)
	c.recordOutcome(accountID.String(), source, reason, false, false, result.Status, time.Since(started))
	c.logSyncRequested(ctx, source, reason, resultUserID(result), accountID, false, false, time.Since(started), result)
	c.logSyncFinished(ctx, result)
	return result, err
}

func resultUserID(result CalendarPipelineResult) uuid.UUID {
	for _, src := range result.Sources {
		if src.UserID != uuid.Nil {
			return src.UserID
		}
	}
	return uuid.Nil
}

func (c *CalendarSyncCoordinator) tryCooldownSkip(
	ctx context.Context,
	userID uuid.UUID,
	maxAge time.Duration,
) (skipped bool, result CalendarPipelineResult, err error) {
	accounts, err := c.svc.listSyncableAccounts(ctx, userID)
	if err != nil {
		return false, CalendarPipelineResult{}, err
	}
	if len(accounts) == 0 {
		// Let SyncPipeline produce the same not-found error path.
		return false, CalendarPipelineResult{}, nil
	}

	now := time.Now().UTC()
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
	if !allFresh || referenceSyncedAt == nil {
		return false, CalendarPipelineResult{}, nil
	}

	sources, listErr := c.svc.sources.ListByUserID(ctx, userID)
	if listErr != nil {
		return false, CalendarPipelineResult{}, listErr
	}
	return true, CalendarPipelineResult{
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

func (c *CalendarSyncCoordinator) beginUser(userID uuid.UUID) (*inflightSync, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.inflightUser[userID]; ok {
		return existing, false
	}
	slot := &inflightSync{done: make(chan struct{})}
	c.inflightUser[userID] = slot
	return slot, true
}

func (c *CalendarSyncCoordinator) finishUser(userID uuid.UUID, slot *inflightSync) {
	c.mu.Lock()
	delete(c.inflightUser, userID)
	c.mu.Unlock()
	close(slot.done)
}

func (c *CalendarSyncCoordinator) beginAccount(accountID uuid.UUID) (*inflightSync, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.inflightAccount[accountID]; ok {
		return existing, false
	}
	slot := &inflightSync{done: make(chan struct{})}
	c.inflightAccount[accountID] = slot
	return slot, true
}

func (c *CalendarSyncCoordinator) finishAccount(accountID uuid.UUID, slot *inflightSync) {
	c.mu.Lock()
	delete(c.inflightAccount, accountID)
	c.mu.Unlock()
	close(slot.done)
}

func (c *CalendarSyncCoordinator) observePipeline(result CalendarPipelineResult, err error) {
	// Per-event create/update/skip/delete counters are recorded in syncSourceEvents.
	if result.DurationMs > 0 {
		c.metrics.ObserveSyncDuration(time.Duration(result.DurationMs) * time.Millisecond)
	}
	_ = err
	_ = result
}

func (c *CalendarSyncCoordinator) recordOutcome(
	key, source, reason string,
	skipped, cooldown bool,
	status string,
	_ time.Duration,
) {
	calendarsyncmetrics.RecordStatus(key, calendarsyncmetrics.SyncStatus{
		Source:    source,
		Reason:    reason,
		Skipped:   skipped,
		Cooldown:  cooldown,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	})
}

func (c *CalendarSyncCoordinator) logSyncRequested(
	ctx context.Context,
	source, reason string,
	userID, accountID uuid.UUID,
	skipped, cooldown bool,
	duration time.Duration,
	result CalendarPipelineResult,
) {
	if c.log == nil {
		return
	}
	attrs := []any{
		"source", source,
		"reason", reason,
		"skipped", skipped,
		"cooldown", cooldown,
		"duration_ms", duration.Milliseconds(),
	}
	if userID != uuid.Nil {
		attrs = append(attrs, constant.LogAttrUserID, userID.String())
	}
	if accountID != uuid.Nil {
		attrs = append(attrs, "connected_account_id", accountID.String())
	}
	if result.Status != "" {
		attrs = append(attrs, "status", result.Status)
	}
	c.log.Info(ctx, "calendar sync requested", attrs...)
}

func (c *CalendarSyncCoordinator) logSyncFinished(ctx context.Context, result CalendarPipelineResult) {
	if c.log == nil || result.Skipped {
		return
	}
	c.log.Info(ctx, "calendar sync completed",
		"source", result.Trigger,
		"status", result.Status,
		"events_scanned", result.EventsScanned,
		"events_created", result.EventsCreated,
		"events_updated", result.EventsUpdated,
		"events_skipped", result.EventsSkipped,
		"events_deleted", result.EventsDeleted,
		"duration_ms", result.DurationMs,
	)
}
