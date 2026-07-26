package icscalendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// FetchResult is an ICS HTTP download outcome.
type FetchResult struct {
	Body         []byte
	ETag         string
	LastModified string
	NotModified  bool
	StatusCode   int
}

// Client fetches ICS feeds with conditional request support.
type Client struct {
	http    *http.Client
	baseUA  string
}

// Config configures the ICS HTTP client.
type Config struct {
	Timeout time.Duration
	UA      string
}

// NewClient constructs an ICS feed client.
func NewClient(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.UA == "" {
		cfg.UA = constant.ICSHTTPUserAgent
	}
	return &Client{
		http:   &http.Client{Timeout: cfg.Timeout},
		baseUA: cfg.UA,
	}
}

// Fetch downloads an ICS feed, honoring If-None-Match / If-Modified-Since when provided.
func (c *Client) Fetch(ctx context.Context, feedURL, etag, lastModified string) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("ics request: %w", err)
	}
	req.Header.Set("User-Agent", c.baseUA)
	req.Header.Set("Accept", "text/calendar, application/calendar+json, text/plain, */*")
	if etag = strings.TrimSpace(etag); etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified = strings.TrimSpace(lastModified); lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return FetchResult{}, fmt.Errorf("ics fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return FetchResult{}, fmt.Errorf("ics read body: %w", err)
	}

	out := FetchResult{
		Body:         body,
		ETag:         strings.TrimSpace(resp.Header.Get("ETag")),
		LastModified: strings.TrimSpace(resp.Header.Get("Last-Modified")),
		StatusCode:   resp.StatusCode,
	}
	if resp.StatusCode == http.StatusNotModified {
		out.NotModified = true
		return out, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return out, &AuthError{Status: resp.StatusCode, Body: truncate(string(body), 200)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("ics fetch status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	if len(body) == 0 {
		return out, fmt.Errorf("ics feed body is empty")
	}
	return out, nil
}

// AuthError is returned for 401/403 ICS feed responses.
type AuthError struct {
	Status int
	Body   string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("ics feed auth error (%d): %s", e.Status, e.Body)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// EncodeSyncCursor packs ETag + Last-Modified into a single sync token.
func EncodeSyncCursor(etag, lastModified string) string {
	etag = strings.TrimSpace(etag)
	lastModified = strings.TrimSpace(lastModified)
	if etag == "" && lastModified == "" {
		return ""
	}
	return etag + "\x1e" + lastModified
}

// DecodeSyncCursor unpacks a sync token into ETag + Last-Modified.
func DecodeSyncCursor(token string) (etag, lastModified string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ""
	}
	parts := strings.SplitN(token, "\x1e", 2)
	etag = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		lastModified = strings.TrimSpace(parts[1])
	}
	return etag, lastModified
}
