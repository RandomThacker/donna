// Package calendarsyncmetrics holds in-process counters for calendar sync optimization.
// Prometheus export can wrap these later; Observability metrics pillar is still future.
package calendarsyncmetrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Names match the Phase 1 sprint contract.
const (
	NameSyncRequestedTotal  = "calendar_sync_requested_total"
	NameSyncSkippedTotal    = "calendar_sync_skipped_total"
	NameEventUpdatesTotal   = "calendar_event_updates_total"
	NameEventSkippedTotal   = "calendar_event_skipped_total"
	NameEventCreatedTotal   = "calendar_event_created_total"
	NameEventDeletedTotal   = "calendar_event_deleted_total"
	NameSyncDurationMS      = "calendar_sync_duration_ms"
)

// Registry is a process-wide set of calendar sync counters.
type Registry struct {
	syncRequested atomic.Int64
	syncSkipped   atomic.Int64
	eventUpdates  atomic.Int64
	eventSkipped  atomic.Int64
	eventCreated  atomic.Int64
	eventDeleted  atomic.Int64
	durationSumMS atomic.Int64
	durationCount atomic.Int64
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

// IncEventUpdates increments calendar_event_updates_total by n.
func (r *Registry) IncEventUpdates(n int) {
	if n > 0 {
		r.eventUpdates.Add(int64(n))
	}
}

// IncEventSkipped increments calendar_event_skipped_total by n.
func (r *Registry) IncEventSkipped(n int) {
	if n > 0 {
		r.eventSkipped.Add(int64(n))
	}
}

// IncEventCreated increments calendar_event_created_total by n.
func (r *Registry) IncEventCreated(n int) {
	if n > 0 {
		r.eventCreated.Add(int64(n))
	}
}

// IncEventDeleted increments calendar_event_deleted_total by n.
func (r *Registry) IncEventDeleted(n int) {
	if n > 0 {
		r.eventDeleted.Add(int64(n))
	}
}

// ObserveSyncDuration records calendar_sync_duration_ms.
func (r *Registry) ObserveSyncDuration(d time.Duration) {
	r.durationSumMS.Add(d.Milliseconds())
	r.durationCount.Add(1)
}

// Snapshot is a point-in-time view of counters.
type Snapshot struct {
	SyncRequestedTotal int64 `json:"calendar_sync_requested_total"`
	SyncSkippedTotal   int64 `json:"calendar_sync_skipped_total"`
	EventUpdatesTotal  int64 `json:"calendar_event_updates_total"`
	EventSkippedTotal  int64 `json:"calendar_event_skipped_total"`
	EventCreatedTotal  int64 `json:"calendar_event_created_total"`
	EventDeletedTotal  int64 `json:"calendar_event_deleted_total"`
	SyncDurationSumMS  int64 `json:"calendar_sync_duration_ms_sum"`
	SyncDurationCount  int64 `json:"calendar_sync_duration_ms_count"`
}

// Snapshot returns current counter values.
func (r *Registry) Snapshot() Snapshot {
	return Snapshot{
		SyncRequestedTotal: r.syncRequested.Load(),
		SyncSkippedTotal:   r.syncSkipped.Load(),
		EventUpdatesTotal:  r.eventUpdates.Load(),
		EventSkippedTotal:  r.eventSkipped.Load(),
		EventCreatedTotal:  r.eventCreated.Load(),
		EventDeletedTotal:  r.eventDeleted.Load(),
		SyncDurationSumMS:  r.durationSumMS.Load(),
		SyncDurationCount:  r.durationCount.Load(),
	}
}

// Reset clears counters (tests only).
func (r *Registry) Reset() {
	r.syncRequested.Store(0)
	r.syncSkipped.Store(0)
	r.eventUpdates.Store(0)
	r.eventSkipped.Store(0)
	r.eventCreated.Store(0)
	r.eventDeleted.Store(0)
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
