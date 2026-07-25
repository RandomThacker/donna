package model

import (
	"encoding/json"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// CalendarSourceResponse is the HTTP DTO for a calendar source.
type CalendarSourceResponse struct {
	ID                  string          `json:"id"`
	PublicID            string          `json:"public_id"`
	ConnectedAccountID  string          `json:"connected_account_id"`
	ProviderCalendarID  string          `json:"provider_calendar_id"`
	Name                string          `json:"name"`
	Color               *string         `json:"color,omitempty"`
	IsPrimaryOnProvider bool            `json:"is_primary_on_provider"`
	IsWritable          bool            `json:"is_writable"`
	AccessRole          *string         `json:"access_role,omitempty"`
	SyncEnabled         bool            `json:"sync_enabled"`
	SyncToken           *string         `json:"sync_token,omitempty"`
	LastSyncedAt        *string         `json:"last_synced_at,omitempty"`
	Timezone            *string         `json:"timezone,omitempty"`
	ProviderMetadata    json.RawMessage `json:"provider_metadata"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}

// CalendarSyncResponse is returned by POST /calendar/sync and /calendar/sync/ensure.
type CalendarSyncResponse struct {
	Sources      []CalendarSourceResponse `json:"sources"`
	CreatedCount int                      `json:"created_count"`
	UpdatedCount int                      `json:"updated_count"`
	RemovedCount int                      `json:"removed_count"`
	SyncedAt     string                   `json:"synced_at"`
	DurationMs   int                      `json:"duration_ms"`
	Incremental  bool                     `json:"incremental"`
	Skipped      bool                     `json:"skipped"`
	SyncStatus   string                   `json:"sync_status"`
}

// CalendarSourcesResponse is returned by GET /calendar/sources.
type CalendarSourcesResponse struct {
	Sources []CalendarSourceResponse     `json:"sources"`
	Sync    *CalendarAccountSyncResponse `json:"sync,omitempty"`
}

// CalendarAccountSyncResponse exposes connected-account sync observability (Donna DB).
type CalendarAccountSyncResponse struct {
	SyncStatus         string  `json:"sync_status"`
	LastSuccessfulSync *string `json:"last_successful_sync,omitempty"`
	LastFailedSync     *string `json:"last_failed_sync,omitempty"`
	SyncDurationMs     *int    `json:"sync_duration_ms,omitempty"`
	RecordsCreated     int     `json:"records_created"`
	RecordsUpdated     int     `json:"records_updated"`
	RecordsDeleted     int     `json:"records_deleted"`
}

// CalendarAccountSyncFromEntity maps connected-account sync fields to the transport model.
func CalendarAccountSyncFromEntity(account entity.ConnectedAccount) *CalendarAccountSyncResponse {
	if account.ID == uuid.Nil {
		return nil
	}
	resp := &CalendarAccountSyncResponse{
		SyncStatus:     account.CalendarSyncStatus,
		SyncDurationMs: account.LastSyncDurationMs,
		RecordsCreated: account.LastSyncCreatedCount,
		RecordsUpdated: account.LastSyncUpdatedCount,
		RecordsDeleted: account.LastSyncDeletedCount,
	}
	if account.LastSyncedAt != nil {
		v := account.LastSyncedAt.UTC().Format(time.RFC3339Nano)
		resp.LastSuccessfulSync = &v
	}
	if account.LastFailedSyncAt != nil {
		v := account.LastFailedSyncAt.UTC().Format(time.RFC3339Nano)
		resp.LastFailedSync = &v
	}
	return resp
}

// CalendarSourceFromEntity maps a domain source to the transport model.
func CalendarSourceFromEntity(s entity.CalendarSource) CalendarSourceResponse {
	meta := json.RawMessage(s.ProviderMetadata)
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	resp := CalendarSourceResponse{
		ID:                  s.ID.String(),
		PublicID:            s.PublicID,
		ConnectedAccountID:  s.ConnectedAccountID.String(),
		ProviderCalendarID:  s.ProviderCalendarID,
		Name:                s.Name,
		Color:               s.Color,
		IsPrimaryOnProvider: s.IsPrimaryOnProvider,
		IsWritable:          s.IsWritable,
		AccessRole:          s.AccessRole,
		SyncEnabled:         s.SyncEnabled,
		SyncToken:           s.SyncCursor,
		Timezone:            s.Timezone,
		ProviderMetadata:    meta,
		CreatedAt:           s.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           s.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if s.LastSyncedAt != nil {
		v := s.LastSyncedAt.UTC().Format(time.RFC3339Nano)
		resp.LastSyncedAt = &v
	}
	return resp
}

// CalendarSourcesFromEntities maps a slice of sources.
func CalendarSourcesFromEntities(sources []entity.CalendarSource) []CalendarSourceResponse {
	out := make([]CalendarSourceResponse, 0, len(sources))
	for _, source := range sources {
		out = append(out, CalendarSourceFromEntity(source))
	}
	return out
}
