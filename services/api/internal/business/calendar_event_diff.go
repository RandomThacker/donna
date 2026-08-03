package business

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarsyncmetrics"
	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// Event decision reasons — logged on every skip/update and used for metric routing.
const (
	eventReasonETag              = calendarsyncmetrics.ReasonETag
	eventReasonProviderUpdatedAt = calendarsyncmetrics.ReasonProviderUpdatedAt
	eventReasonContentHash       = calendarsyncmetrics.ReasonContentHash
	eventReasonETagChanged       = calendarsyncmetrics.ReasonETagChanged
	eventReasonResurrect         = calendarsyncmetrics.ReasonResurrect
)

// shouldSkipEventUpdate decides whether UpdateFromSync can be skipped.
//
// Rules (do not trust ETag alone — ICS publishers are unreliable):
//  1. Soft-deleted rows always update (resurrect).
//  2. ETag changed → update (reason=etag_changed).
//  3. ETag same → confirm with provider_updated_at when present;
//     same timestamps skip (reason=etag); disagreement falls through to content hash.
//  4. ETag missing → provider_updated_at (same → skip) else content hash.
//
// Hash is cheap relative to rewriting large JSONB rows.
func shouldSkipEventUpdate(existing, mapped entity.CalendarEvent) (skip bool, reason string) {
	if existing.DeletedAt != nil {
		return false, eventReasonResurrect
	}

	existingETag := strings.TrimSpace(derefString(existing.ProviderETag))
	mappedETag := strings.TrimSpace(derefString(mapped.ProviderETag))
	etagComparable := existingETag != "" && mappedETag != ""

	if etagComparable {
		if existingETag != mappedETag {
			return false, eventReasonETagChanged
		}
		// ETag matches — confirm with updated_at when available; never skip on ETag alone.
		if existing.ProviderUpdatedAt != nil && mapped.ProviderUpdatedAt != nil {
			if existing.ProviderUpdatedAt.UTC().Equal(mapped.ProviderUpdatedAt.UTC()) {
				return true, eventReasonETag
			}
			return skipByContentHash(existing, mapped)
		}
		return skipByContentHash(existing, mapped)
	}

	// ETag missing on either side → updated_at, then hash.
	if existing.ProviderUpdatedAt != nil && mapped.ProviderUpdatedAt != nil {
		if existing.ProviderUpdatedAt.UTC().Equal(mapped.ProviderUpdatedAt.UTC()) {
			return true, eventReasonProviderUpdatedAt
		}
		return skipByContentHash(existing, mapped)
	}

	return skipByContentHash(existing, mapped)
}

func skipByContentHash(existing, mapped entity.CalendarEvent) (skip bool, reason string) {
	if eventContentHash(existing) == eventContentHash(mapped) {
		return true, eventReasonContentHash
	}
	return false, eventReasonContentHash
}

// providerIdentityChanged reports whether etag or provider_updated_at changed.
// Used to decide whether provider_payload must be rewritten.
func providerIdentityChanged(existing, mapped entity.CalendarEvent) bool {
	if !stringPtrEqualTrimmed(existing.ProviderETag, mapped.ProviderETag) {
		return true
	}
	return !timePtrEqual(existing.ProviderUpdatedAt, mapped.ProviderUpdatedAt)
}

func eventContentHash(e entity.CalendarEvent) string {
	h := sha256.New()
	writeHashField(h, e.Title)
	writeHashField(h, derefString(e.Description))
	writeHashField(h, derefString(e.Location))
	writeHashField(h, e.StartsAt.UTC().Format(time.RFC3339Nano))
	writeHashField(h, e.EndsAt.UTC().Format(time.RFC3339Nano))
	writeHashField(h, fmt.Sprintf("%t", e.IsAllDay))
	writeHashField(h, e.Status)
	writeHashField(h, derefString(e.Visibility))
	writeHashField(h, derefString(e.Timezone))
	writeHashField(h, normalizeJSONBytes(e.OrganizerSummary))
	writeHashField(h, normalizeJSONBytes(e.AttendeesSummary))
	writeHashField(h, derefString(e.RecurrenceRule))
	return hex.EncodeToString(h.Sum(nil))
}

func writeHashField(h interface{ Write([]byte) (int, error) }, v string) {
	_, _ = h.Write([]byte(v))
	_, _ = h.Write([]byte{0})
}

func normalizeJSONBytes(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(canonical)
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func stringPtrEqualTrimmed(a, b *string) bool {
	return strings.TrimSpace(derefString(a)) == strings.TrimSpace(derefString(b))
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.UTC().Equal(b.UTC())
}
