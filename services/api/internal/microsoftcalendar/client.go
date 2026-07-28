package microsoftcalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// RemoteCalendar is a Graph calendar list entry Donna cares about.
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

// ListOptions controls calendar list fetch (Microsoft has no calendar-list syncToken).
type ListOptions struct {
	SyncToken string // ignored; always full list
}

// ListResult is a pageable calendars response.
type ListResult struct {
	Calendars     []RemoteCalendar
	NextSyncToken string
	Incremental   bool
}

// Config configures the Microsoft Graph Calendar client.
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client calls Microsoft Graph calendar APIs.
type Client struct {
	cfg Config
}

// NewClient constructs a Graph Calendar API client.
func NewClient(cfg Config) *Client {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = constant.MicrosoftGraphAPIBaseURL
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

// ListCalendars fetches calendars for the authorized user (paginated via @odata.nextLink).
// Microsoft Graph has no incremental calendar-list sync; Incremental is always false
// and NextSyncToken is always empty.
func (c *Client) ListCalendars(ctx context.Context, accessToken string, _ ListOptions) (ListResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return ListResult{}, fmt.Errorf("access token is required")
	}

	var out ListResult
	nextURL := c.cfg.BaseURL + "/me/calendars"

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return ListResult{}, fmt.Errorf("calendars request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			return ListResult{}, fmt.Errorf("calendars: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return ListResult{}, fmt.Errorf("read calendars: %w", readErr)
		}
		if resp.StatusCode == http.StatusGone {
			return ListResult{}, &GoneError{Body: truncate(string(body), 500)}
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return ListResult{}, &AuthError{Status: resp.StatusCode, Body: truncate(string(body), 500)}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return ListResult{}, fmt.Errorf("calendars status %d: %s", resp.StatusCode, truncate(string(body), 500))
		}

		var raw struct {
			Value    []map[string]any `json:"value"`
			NextLink string           `json:"@odata.nextLink"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return ListResult{}, fmt.Errorf("decode calendars: %w", err)
		}
		for _, item := range raw.Value {
			cal, ok := mapRemoteCalendar(item)
			if !ok {
				continue
			}
			out.Calendars = append(out.Calendars, cal)
		}
		nextURL = strings.TrimSpace(raw.NextLink)
	}

	out.Incremental = false
	out.NextSyncToken = ""
	return out, nil
}

// AuthError indicates the access token lacks Calendar permission or is invalid.
type AuthError struct {
	Status int
	Body   string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("microsoft calendar auth failed (%d): %s", e.Status, e.Body)
}

// GoneError means the delta/sync cursor is invalid and a full sync is required.
type GoneError struct {
	Body string
}

func (e *GoneError) Error() string {
	return fmt.Sprintf("microsoft calendar sync cursor gone: %s", e.Body)
}

func mapRemoteCalendar(item map[string]any) (RemoteCalendar, bool) {
	id := stringField(item, "id")
	if id == "" {
		return RemoteCalendar{}, false
	}
	name := stringField(item, "name")
	if name == "" {
		name = id
	}
	writable := boolField(item, "canEdit")
	accessRole := "reader"
	if writable {
		accessRole = "writer"
	}
	if boolField(item, "isDefaultCalendar") {
		accessRole = "owner"
	}
	return RemoteCalendar{
		ID:         id,
		Name:       name,
		Primary:    boolField(item, "isDefaultCalendar"),
		Writable:   writable,
		AccessRole: accessRole,
		Color:      stringField(item, "hexColor"),
		ETag:       stringField(item, "changeKey"),
		Raw:        item,
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
