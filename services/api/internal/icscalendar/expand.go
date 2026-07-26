package icscalendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/teambition/rrule-go"
)

// ExpandRecurring materializes RRULE masters into concrete occurrences in [from, to].
// Exception overrides (RECURRENCE-ID) win over generated instances. EXDATE is skipped.
// Non-recurring events are kept as-is.
func ExpandRecurring(events []calendarprovider.RemoteEvent, from, to time.Time) []calendarprovider.RemoteEvent {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		now := time.Now().UTC()
		from = now.Add(-365 * 24 * time.Hour)
		to = now.Add(730 * 24 * time.Hour)
	}

	type overrideKey struct {
		uid string
		occ string
	}

	overrides := map[overrideKey]calendarprovider.RemoteEvent{}
	var masters []calendarprovider.RemoteEvent
	out := make([]calendarprovider.RemoteEvent, 0, len(events))

	for _, event := range events {
		uid := strings.TrimSpace(event.ID)
		if uid == "" {
			continue
		}
		recurrenceID := ""
		if raw, ok := event.Raw["recurrence_id"].(string); ok {
			recurrenceID = strings.TrimSpace(raw)
		}
		if recurrenceID != "" {
			occKey := occurrenceKey(parseLooseTime(recurrenceID, event.Timezone, event.IsAllDay))
			if occKey == "" {
				occKey = occurrenceKey(event.StartsAt)
			}
			fixed := event
			masterUID := uid
			if event.RecurringEventID != "" {
				masterUID = event.RecurringEventID
			}
			fixed.ID = instanceProviderID(masterUID, occKey)
			fixed.RecurringEventID = masterUID
			fixed.Recurrence = nil
			overrides[overrideKey{uid: masterUID, occ: occKey}] = fixed
			continue
		}
		if len(event.Recurrence) > 0 {
			masters = append(masters, event)
			continue
		}
		out = append(out, event)
	}

	usedOverrides := map[overrideKey]struct{}{}
	for _, master := range masters {
		uid := strings.TrimSpace(master.ID)
		exdates := parseExdates(master)
		duration := master.EndsAt.Sub(master.StartsAt)
		if duration < 0 {
			duration = 0
		}

		occurrences, err := expandMaster(master, from, to)
		if err != nil || len(occurrences) == 0 {
			// Fall back to the master DTSTART so the series is not dropped entirely.
			if !master.EndsAt.Before(from) && !master.StartsAt.After(to) {
				out = append(out, master)
			}
			continue
		}

		for _, occStart := range occurrences {
			occKey := occurrenceKey(occStart)
			if _, excluded := exdates[occKey]; excluded {
				continue
			}
			key := overrideKey{uid: uid, occ: occKey}
			if override, ok := overrides[key]; ok {
				out = append(out, override)
				usedOverrides[key] = struct{}{}
				continue
			}
			inst := master
			inst.ID = instanceProviderID(uid, occKey)
			inst.RecurringEventID = uid
			inst.Recurrence = nil
			inst.StartsAt = occStart.UTC()
			inst.EndsAt = occStart.Add(duration).UTC()
			if inst.Raw == nil {
				inst.Raw = map[string]any{}
			} else {
				copied := make(map[string]any, len(inst.Raw)+1)
				for k, v := range inst.Raw {
					copied[k] = v
				}
				inst.Raw = copied
			}
			inst.Raw["expanded_from_rrule"] = true
			out = append(out, inst)
		}
	}

	// Emit unused overrides (e.g. modified instances outside the rrule iterator edge cases).
	for key, override := range overrides {
		if _, used := usedOverrides[key]; used {
			continue
		}
		if override.EndsAt.Before(from) || override.StartsAt.After(to) {
			continue
		}
		out = append(out, override)
	}
	return out
}

func expandMaster(master calendarprovider.RemoteEvent, from, to time.Time) ([]time.Time, error) {
	ruleText := strings.TrimSpace(strings.Join(master.Recurrence, "\n"))
	if ruleText == "" {
		return nil, fmt.Errorf("empty rrule")
	}
	// rrule-go SetFromText expects DTSTART + RRULE lines.
	dtStart := master.StartsAt.UTC().Format("20060102T150405Z")
	if master.IsAllDay {
		dtStart = master.StartsAt.UTC().Format("20060102")
	}
	payload := "DTSTART:" + dtStart + "\n" + ruleText
	set, err := rrule.StrToRRuleSet(payload)
	if err != nil {
		// Some feeds only include RRULE without DTSTART prefix compatibility — try raw.
		set, err = rrule.StrToRRuleSet("DTSTART:" + dtStart + "\nRRULE:" + strings.TrimPrefix(strings.TrimPrefix(ruleText, "RRULE:"), "rrule:"))
		if err != nil {
			return nil, err
		}
	}
	return set.Between(from, to, true), nil
}

func parseExdates(event calendarprovider.RemoteEvent) map[string]struct{} {
	out := map[string]struct{}{}
	raw, _ := event.Raw["exdate"].(string)
	if strings.TrimSpace(raw) == "" {
		return out
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// EXDATE values are often floating local times; prefer event TZID.
		t := parseLooseTime(part, event.Timezone, event.IsAllDay)
		if t.IsZero() {
			continue
		}
		out[occurrenceKey(t)] = struct{}{}
		// Also accept the raw UTC key if the feed already used Zulu.
		if strings.HasSuffix(strings.ToUpper(part), "Z") {
			out[occurrenceKey(t)] = struct{}{}
		}
	}
	return out
}

func parseLooseTime(value, tzid string, allDay bool) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if t, err := parseICSTimeValue(value, allDay, tzid); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

func occurrenceKey(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("20060102T150405Z")
}

func instanceProviderID(uid, occKey string) string {
	return uid + "_" + occKey
}
