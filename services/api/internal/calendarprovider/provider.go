package calendarprovider

import (
	"context"
	"fmt"
	"time"
)

// Provider is the calendar integration port. Google and Microsoft implement this.
type Provider interface {
	Name() string
	ListCalendars(ctx context.Context, accessToken string, opts ListCalendarsOptions) (ListCalendarsResult, error)
	ListEvents(ctx context.Context, accessToken, calendarID string, opts ListEventsOptions) (ListEventsResult, error)
}

// ListCalendarsOptions controls calendar-list incremental sync.
type ListCalendarsOptions struct {
	SyncToken string
}

// ListCalendarsResult is a provider calendar list page set.
type ListCalendarsResult struct {
	Calendars     []RemoteCalendar
	NextSyncToken string
	Incremental   bool
}

// ListEventsOptions controls event list / delta sync.
type ListEventsOptions struct {
	SyncToken string
	TimeMin   time.Time
	TimeMax   time.Time
}

// ListEventsResult is a provider event list / delta page set.
type ListEventsResult struct {
	Events        []RemoteEvent
	NextSyncToken string
	Incremental   bool
	// ReplaceAll means the result is the full truth for this source within the
	// sync window — orchestration should soft-delete events not present in Events.
	ReplaceAll bool
}

// RemoteCalendar is a provider-neutral calendar feed.
type RemoteCalendar struct {
	ID          string
	Name        string
	Primary     bool
	Writable    bool
	AccessRole  string
	Color       string
	TimeZone    string
	ETag        string
	Description string
	Hidden      bool
	Selected    bool
	Deleted     bool
	Raw         map[string]any
}

// RemoteEvent is a provider-neutral calendar event.
type RemoteEvent struct {
	ID                   string
	ETag                 string
	Status               string
	Title                string
	Description          string
	Location             string
	OnlineMeetingURL     string
	StartsAt             time.Time
	EndsAt               time.Time
	IsAllDay             bool
	Timezone             string
	Visibility           string
	OrganizerEmail       string
	OrganizerDisplayName string
	OrganizerSelf        bool
	Attendees            []RemoteAttendee
	Recurrence           []string
	RecurringEventID     string
	UpdatedAt            time.Time
	Deleted              bool
	Raw                  map[string]any
}

// RemoteAttendee is a thin attendee projection.
type RemoteAttendee struct {
	Email          string `json:"email,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	ResponseStatus string `json:"response_status,omitempty"`
	Organizer      bool   `json:"organizer,omitempty"`
	Self           bool   `json:"self,omitempty"`
}

// TokenSet is a provider-neutral OAuth token payload for refresh.
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	Scope        string
}

// TokenRefresher refreshes access tokens for one OAuth provider.
type TokenRefresher interface {
	RefreshAccessToken(ctx context.Context, refreshToken string) (TokenSet, error)
}

// SyncCursorInvalidError means the provider sync/delta token must be cleared (HTTP 410).
type SyncCursorInvalidError struct {
	Body string
}

func (e *SyncCursorInvalidError) Error() string {
	if e == nil {
		return "calendar sync cursor invalid"
	}
	if e.Body == "" {
		return "calendar sync cursor invalid"
	}
	return fmt.Sprintf("calendar sync cursor invalid: %s", e.Body)
}

// AuthError means the provider denied the request (401/403).
type AuthError struct {
	Status int
	Body   string
}

func (e *AuthError) Error() string {
	if e == nil {
		return "calendar provider auth error"
	}
	return fmt.Sprintf("calendar provider auth error (%d): %s", e.Status, e.Body)
}
