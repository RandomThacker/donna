package icscalendar

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

// NormalizeFeedURL canonicalizes webcal/http(s) ICS URLs for stable identity.
func NormalizeFeedURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errEmptyURL
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "webcal://"):
		raw = "https://" + raw[len("webcal://"):]
	case strings.HasPrefix(lower, "webcals://"):
		raw = "https://" + raw[len("webcals://"):]
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errUnsupportedScheme
	}
	if parsed.Host == "" {
		return "", errEmptyURL
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

// FeedAccountID returns a deterministic provider_account_id for an ICS feed URL.
func FeedAccountID(normalizedURL string) string {
	sum := sha256.Sum256([]byte(normalizedURL))
	return "ics_" + hex.EncodeToString(sum[:16])
}

// FeedCalendarID returns a deterministic provider_calendar_id for an ICS feed URL.
func FeedCalendarID(normalizedURL string) string {
	return FeedAccountID(normalizedURL)
}
