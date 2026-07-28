package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RemoteCalendar is a provider calendar list entry Donna cares about.
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

// ListOptions controls calendarList pagination / incremental sync.
type ListOptions struct {
	SyncToken string
}

// ListResult is a pageable calendarList response.
type ListResult struct {
	Calendars     []RemoteCalendar
	NextSyncToken string
	Incremental   bool
}

// Config configures the Google Calendar API client.
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client calls Google Calendar API v3.
type Client struct {
	cfg Config
}

// NewClient constructs a Calendar API client.
func NewClient(cfg Config) *Client {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = "https://www.googleapis.com/calendar/v3"
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{cfg: cfg}
}

// ListCalendars fetches calendars for the authorized user (paginated).
// When opts.SyncToken is set, Google returns only changes (incremental).
// HTTP 410 means the token is invalid — callers should clear it and full-sync.
func (c *Client) ListCalendars(ctx context.Context, accessToken string, opts ListOptions) (ListResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return ListResult{}, fmt.Errorf("access token is required")
	}

	syncToken := strings.TrimSpace(opts.SyncToken)
	var (
		out       ListResult
		pageToken string
	)
	out.Incremental = syncToken != ""

	for {
		endpoint := c.cfg.BaseURL + "/users/me/calendarList"
		q := url.Values{}
		q.Set("maxResults", "250")
		if syncToken != "" {
			q.Set("syncToken", syncToken)
		} else {
			q.Set("showHidden", "true")
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
		if err != nil {
			return ListResult{}, fmt.Errorf("calendarList request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			return ListResult{}, fmt.Errorf("calendarList: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return ListResult{}, fmt.Errorf("read calendarList: %w", readErr)
		}
		if resp.StatusCode == http.StatusGone {
			return ListResult{}, &GoneError{Body: truncate(string(body), 500)}
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return ListResult{}, &AuthError{Status: resp.StatusCode, Body: truncate(string(body), 500)}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return ListResult{}, fmt.Errorf("calendarList status %d: %s", resp.StatusCode, truncate(string(body), 500))
		}

		var raw struct {
			Items         []map[string]any `json:"items"`
			NextPageToken string           `json:"nextPageToken"`
			NextSyncToken string           `json:"nextSyncToken"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return ListResult{}, fmt.Errorf("decode calendarList: %w", err)
		}

		for _, item := range raw.Items {
			cal, ok := mapRemoteCalendar(item)
			if !ok {
				continue
			}
			out.Calendars = append(out.Calendars, cal)
		}
		if raw.NextSyncToken != "" {
			out.NextSyncToken = raw.NextSyncToken
		}
		if raw.NextPageToken == "" {
			break
		}
		pageToken = raw.NextPageToken
	}

	return out, nil
}

// AuthError indicates the access token lacks Calendar permission or is invalid.
type AuthError struct {
	Status int
	Body   string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("google calendar auth failed (%d): %s", e.Status, e.Body)
}

// GoneError means the syncToken is invalid and a full sync is required.
type GoneError struct {
	Body string
}

func (e *GoneError) Error() string {
	return fmt.Sprintf("google calendar sync token gone: %s", e.Body)
}

func mapRemoteCalendar(item map[string]any) (RemoteCalendar, bool) {
	id := stringField(item, "id")
	if id == "" {
		return RemoteCalendar{}, false
	}
	name := stringField(item, "summary")
	if name == "" {
		name = id
	}
	accessRole := stringField(item, "accessRole")
	writable := accessRole == "owner" || accessRole == "writer"
	return RemoteCalendar{
		ID:          id,
		Name:        name,
		Primary:     boolField(item, "primary"),
		Writable:    writable,
		AccessRole:  accessRole,
		Color:       firstNonEmpty(stringField(item, "backgroundColor"), stringField(item, "colorId")),
		TimeZone:    stringField(item, "timeZone"),
		ETag:        stringField(item, "etag"),
		Description: stringField(item, "description"),
		Hidden:      boolField(item, "hidden"),
		Selected:    boolField(item, "selected"),
		Deleted:     boolField(item, "deleted"),
		Raw:         item,
	}, true
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func boolField(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
