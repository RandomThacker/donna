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

// CalendarSyncResponse is the consolidated pipeline response for POST /calendar/sync.
type CalendarSyncResponse struct {
	RunID              string                        `json:"run_id,omitempty"`
	Trigger            string                        `json:"trigger"`
	Status             string                        `json:"status"`
	StartedAt          string                        `json:"started_at"`
	FinishedAt         string                        `json:"finished_at"`
	DurationMs         int                           `json:"duration_ms"`
	CalendarsProcessed int                           `json:"calendars_processed"`
	SourcesCreated     int                           `json:"sources_created"`
	SourcesUpdated     int                           `json:"sources_updated"`
	SourcesDeleted     int                           `json:"sources_deleted"`
	EventsCreated      int                           `json:"events_created"`
	EventsUpdated      int                           `json:"events_updated"`
	EventsDeleted      int                           `json:"events_deleted"`
	Failures           []CalendarSyncFailureResponse `json:"failures"`
	Sources            []CalendarSourceResponse      `json:"sources"`
	Incremental        bool                          `json:"incremental"`
	Skipped            bool                          `json:"skipped"`
	SyncStatus         string                        `json:"sync_status"`
}

// CalendarSyncFailureResponse is a per-calendar failure in a sync run.
type CalendarSyncFailureResponse struct {
	CalendarSourceID   string `json:"calendar_source_id,omitempty"`
	ProviderCalendarID string `json:"provider_calendar_id,omitempty"`
	Name               string `json:"name,omitempty"`
	Stage              string `json:"stage"`
	Error              string `json:"error"`
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

// CalendarEventResponse is the HTTP DTO for a calendar event.
type CalendarEventResponse struct {
	ID                       string          `json:"id"`
	PublicID                 string          `json:"public_id"`
	CalendarSourceID         string          `json:"calendar_source_id"`
	Title                    string          `json:"title"`
	Description              *string         `json:"description,omitempty"`
	Location                 *string         `json:"location,omitempty"`
	StartTime                string          `json:"start_time"`
	EndTime                  string          `json:"end_time"`
	Timezone                 *string         `json:"timezone,omitempty"`
	AllDay                   bool            `json:"all_day"`
	Status                   string          `json:"status"`
	Visibility               *string         `json:"visibility,omitempty"`
	Organizer                json.RawMessage `json:"organizer,omitempty"`
	Attendees                json.RawMessage `json:"attendees"`
	RecurringEventID         *string         `json:"recurring_event_id,omitempty"`
	ProviderRecurringEventID *string         `json:"provider_recurring_event_id,omitempty"`
	ProviderEventID          *string         `json:"provider_event_id,omitempty"`
	ProviderUpdatedAt        *string         `json:"provider_updated_at,omitempty"`
	Origin                   string          `json:"origin"`
	CreatedAt                string          `json:"created_at"`
	UpdatedAt                string          `json:"updated_at"`
}

// CalendarEventsResponse is returned by GET /calendar/events.
type CalendarEventsResponse struct {
	Events []CalendarEventResponse `json:"events"`
	From   string                  `json:"from"`
	To     string                  `json:"to"`
}

// CalendarEventSyncResponse is returned by POST /calendar/events/sync.
type CalendarEventSyncResponse struct {
	Events       []CalendarEventResponse `json:"events"`
	CreatedCount int                     `json:"created_count"`
	UpdatedCount int                     `json:"updated_count"`
	RemovedCount int                     `json:"removed_count"`
	SyncedAt     string                  `json:"synced_at"`
	DurationMs   int                     `json:"duration_ms"`
	SourceCount  int                     `json:"source_count"`
}

// CalendarEventFromEntity maps a domain event to the transport model.
func CalendarEventFromEntity(e entity.CalendarEvent) CalendarEventResponse {
	attendees := json.RawMessage(e.AttendeesSummary)
	if len(attendees) == 0 {
		attendees = json.RawMessage(`[]`)
	}
	resp := CalendarEventResponse{
		ID:                       e.ID.String(),
		PublicID:                 e.PublicID,
		CalendarSourceID:         e.CalendarSourceID.String(),
		Title:                    e.Title,
		Description:              e.Description,
		Location:                 e.Location,
		StartTime:                e.StartsAt.UTC().Format(time.RFC3339Nano),
		EndTime:                  e.EndsAt.UTC().Format(time.RFC3339Nano),
		Timezone:                 e.Timezone,
		AllDay:                   e.IsAllDay,
		Status:                   e.Status,
		Visibility:               e.Visibility,
		Attendees:                attendees,
		ProviderRecurringEventID: e.ProviderRecurringEventID,
		ProviderEventID:          e.ProviderEventID,
		Origin:                   e.Origin,
		CreatedAt:                e.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:                e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if len(e.OrganizerSummary) > 0 {
		resp.Organizer = json.RawMessage(e.OrganizerSummary)
	}
	if e.RecurringEventID != nil {
		v := e.RecurringEventID.String()
		resp.RecurringEventID = &v
	}
	if e.ProviderUpdatedAt != nil {
		v := e.ProviderUpdatedAt.UTC().Format(time.RFC3339Nano)
		resp.ProviderUpdatedAt = &v
	}
	return resp
}

// CalendarEventsFromEntities maps a slice of events.
func CalendarEventsFromEntities(events []entity.CalendarEvent) []CalendarEventResponse {
	out := make([]CalendarEventResponse, 0, len(events))
	for _, event := range events {
		out = append(out, CalendarEventFromEntity(event))
	}
	return out
}
