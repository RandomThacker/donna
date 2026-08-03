// Package calendarsyncmetrics holds in-process counters for calendar sync optimization.
// Prometheus export can wrap these later; Observability metrics pillar is still future.
package calendarsyncmetrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Names match the Phase 1 sprint + reason-split observability contract.
const (
	NameSyncRequestedTotal     = "calendar_sync_requested_total"
	NameSyncSkippedTotal       = "calendar_sync_skipped_total"
	NameEventSkipETagTotal     = "calendar_event_skip_etag_total"
	NameEventSkipUpdatedAtTotal = "calendar_event_skip_updated_at_total"
	NameEventSkipHashTotal     = "calendar_event_skip_hash_total"
	NameEventUpdateTotal       = "calendar_event_update_total"
	NameEventCreateTotal       = "calendar_event_create_total"
	NameEventDeleteTotal       = "calendar_event_delete_total"
	NameSyncDurationMS         = "calendar_sync_duration_ms"

	// Decision reason labels (logs + metric routing).
	ReasonETag               = "etag"
	ReasonProviderUpdatedAt  = "provider_updated_at"
	ReasonContentHash        = "content_hash"
	ReasonETagChanged        = "etag_changed"
	ReasonResurrect          = "resurrect"
)

// Registry is a process-wide set of calendar sync counters.
type Registry struct {
	syncRequested      atomic.Int64
	syncSkipped        atomic.Int64
	eventSkipETag      atomic.Int64
	eventSkipUpdatedAt atomic.Int64
	eventSkipHash      atomic.Int64
	eventUpdate        atomic.Int64
	eventCreate        atomic.Int64
	eventDelete        atomic.Int64
	durationSumMS      atomic.Int64
	durationCount      atomic.Int64
}

// Global is the default process registry.
var Global = New()

// New constructs an empty metrics registry.
func New() *Registry {
	return &Registry{}
}

// IncSyncRequested increments calendar_sync_requested_total.
func (r *Registry) IncSyncRequested() { r.syncRequested.Add(1) }

// IncSyncSkipped increments calendar_sync_skipped_total.
func (r *Registry) IncSyncSkipped() { r.syncSkipped.Add(1) }

// IncEventSkipETag increments calendar_event_skip_etag_total.
func (r *Registry) IncEventSkipETag() { r.eventSkipETag.Add(1) }

// IncEventSkipUpdatedAt increments calendar_event_skip_updated_at_total.
func (r *Registry) IncEventSkipUpdatedAt() { r.eventSkipUpdatedAt.Add(1) }

// IncEventSkipHash increments calendar_event_skip_hash_total.
func (r *Registry) IncEventSkipHash() { r.eventSkipHash.Add(1) }

// IncEventUpdate increments calendar_event_update_total.
func (r *Registry) IncEventUpdate() { r.eventUpdate.Add(1) }

// IncEventCreate increments calendar_event_create_total by n.
func (r *Registry) IncEventCreate(n int) {
	if n > 0 {
		r.eventCreate.Add(int64(n))
	}
}

// IncEventDelete increments calendar_event_delete_total by n.
func (r *Registry) IncEventDelete(n int) {
	if n > 0 {
		r.eventDelete.Add(int64(n))
	}
}

// ObserveEventDecision increments the counter for a skip/update reason.
func (r *Registry) ObserveEventDecision(skipped bool, reason string) {
	if skipped {
		switch reason {
		case ReasonETag:
			r.IncEventSkipETag()
		case ReasonProviderUpdatedAt:
			r.IncEventSkipUpdatedAt()
		case ReasonContentHash:
			r.IncEventSkipHash()
		}
		return
	}
	r.IncEventUpdate()
}

// ObserveSyncDuration records calendar_sync_duration_ms.
func (r *Registry) ObserveSyncDuration(d time.Duration) {
	r.durationSumMS.Add(d.Milliseconds())
	r.durationCount.Add(1)
}

// Snapshot is a point-in-time view of counters.
type Snapshot struct {
	SyncRequestedTotal      int64 `json:"calendar_sync_requested_total"`
	SyncSkippedTotal        int64 `json:"calendar_sync_skipped_total"`
	EventSkipETagTotal      int64 `json:"calendar_event_skip_etag_total"`
	EventSkipUpdatedAtTotal int64 `json:"calendar_event_skip_updated_at_total"`
	EventSkipHashTotal      int64 `json:"calendar_event_skip_hash_total"`
	EventUpdateTotal        int64 `json:"calendar_event_update_total"`
	EventCreateTotal        int64 `json:"calendar_event_create_total"`
	EventDeleteTotal        int64 `json:"calendar_event_delete_total"`
	SyncDurationSumMS       int64 `json:"calendar_sync_duration_ms_sum"`
	SyncDurationCount       int64 `json:"calendar_sync_duration_ms_count"`
}

// Snapshot returns current counter values.
func (r *Registry) Snapshot() Snapshot {
	return Snapshot{
		SyncRequestedTotal:      r.syncRequested.Load(),
		SyncSkippedTotal:        r.syncSkipped.Load(),
		EventSkipETagTotal:      r.eventSkipETag.Load(),
		EventSkipUpdatedAtTotal: r.eventSkipUpdatedAt.Load(),
		EventSkipHashTotal:      r.eventSkipHash.Load(),
		EventUpdateTotal:        r.eventUpdate.Load(),
		EventCreateTotal:        r.eventCreate.Load(),
		EventDeleteTotal:        r.eventDelete.Load(),
		SyncDurationSumMS:       r.durationSumMS.Load(),
		SyncDurationCount:       r.durationCount.Load(),
	}
}

// Reset clears counters (tests only).
func (r *Registry) Reset() {
	r.syncRequested.Store(0)
	r.syncSkipped.Store(0)
	r.eventSkipETag.Store(0)
	r.eventSkipUpdatedAt.Store(0)
	r.eventSkipHash.Store(0)
	r.eventUpdate.Store(0)
	r.eventCreate.Store(0)
	r.eventDelete.Store(0)
	r.durationSumMS.Store(0)
	r.durationCount.Store(0)
}

// statusMu guards the last-known sync status map (in-process; future: Redis).
var (
	statusMu   sync.Mutex
	lastStatus = map[string]SyncStatus{}
)

// SyncStatus is a lightweight view of the latest sync admission outcome.
type SyncStatus struct {
	Source    string    `json:"source"`
	Reason    string    `json:"reason"`
	Skipped   bool      `json:"skipped"`
	Cooldown  bool      `json:"cooldown"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RecordStatus stores the latest status for a sync scope key (user or account id).
func RecordStatus(key string, status SyncStatus) {
	statusMu.Lock()
	defer statusMu.Unlock()
	lastStatus[key] = status
}

// StatusFor returns the last recorded status for key, if any.
func StatusFor(key string) (SyncStatus, bool) {
	statusMu.Lock()
	defer statusMu.Unlock()
	s, ok := lastStatus[key]
	return s, ok
}
