package icscalendar

import (
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	ics "github.com/arran4/golang-ical"
)

// ParseCalendar parses RFC5545 ICS bytes into remote events.
func ParseCalendar(data []byte, feedName string) (calendarName string, events []calendarprovider.RemoteEvent, err error) {
	cal, err := ics.ParseCalendar(strings.NewReader(string(data)))
	if err != nil {
		return "", nil, err
	}
	if cal == nil {
		return "", nil, errEmptyCalendar
	}

	calendarName = strings.TrimSpace(feedName)
	for _, p := range cal.CalendarProperties {
		if strings.EqualFold(p.IANAToken, string(ics.PropertyName)) || strings.EqualFold(p.IANAToken, "X-WR-CALNAME") {
			if v := strings.TrimSpace(p.Value); v != "" {
				calendarName = v
				break
			}
		}
	}
	if calendarName == "" {
		calendarName = "ICS Calendar"
	}

	for _, event := range cal.Events() {
		if event == nil {
			continue
		}
		mapped, mapErr := mapVEVENT(event)
		if mapErr != nil || mapped.ID == "" {
			continue
		}
		events = append(events, mapped)
	}
	return calendarName, events, nil
}

func mapVEVENT(event *ics.VEvent) (calendarprovider.RemoteEvent, error) {
	uid := propValue(event, ics.ComponentPropertyUniqueId)
	if uid == "" {
		return calendarprovider.RemoteEvent{}, errEmptyCalendar
	}

	start, allDay, startTZ, err := parseEventTime(event, ics.ComponentPropertyDtStart)
	if err != nil {
		return calendarprovider.RemoteEvent{}, err
	}
	end, endAllDay, endTZ, endErr := parseEventTime(event, ics.ComponentPropertyDtEnd)
	if endErr != nil || end.IsZero() {
		if dur := propValue(event, ics.ComponentProperty(ics.PropertyDuration)); dur != "" {
			if d, derr := parseICSDuration(dur); derr == nil {
				end = start.Add(d)
			}
		}
		if end.IsZero() {
			if allDay {
				end = start.Add(24 * time.Hour)
			} else {
				end = start
			}
		}
	}
	if allDay || endAllDay {
		allDay = true
	}

	status := strings.ToLower(propValue(event, ics.ComponentPropertyStatus))
	switch status {
	case "cancelled", "canceled":
		status = constant.CalendarEventStatusCancelled
	case "tentative":
		status = "tentative"
	case "confirmed", "":
		status = constant.CalendarEventStatusConfirmed
	default:
		if status == "" {
			status = constant.CalendarEventStatusConfirmed
		}
	}

	visibility := strings.ToLower(propValue(event, ics.ComponentPropertyClass))
	switch visibility {
	case "private", "confidential":
		visibility = "private"
	case "public":
		visibility = "public"
	default:
		visibility = ""
	}

	tz := startTZ
	if tz == "" {
		tz = endTZ
	}

	organizerEmail, organizerName := "", ""
	if org := event.GetProperty(ics.ComponentPropertyOrganizer); org != nil {
		organizerEmail, _ = parseCalAddress(org.Value)
		if vals, ok := org.ICalParameters["CN"]; ok && len(vals) > 0 {
			organizerName = vals[0]
		}
	}
	attendees := parseAttendees(event)

	var recurrence []string
	for _, rrule := range event.Properties {
		if rrule.IANAToken == string(ics.PropertyRrule) && strings.TrimSpace(rrule.Value) != "" {
			recurrence = append(recurrence, "RRULE:"+strings.TrimSpace(rrule.Value))
		}
	}

	raw := map[string]any{}
	preserve := map[string]string{
		"url":           propValue(event, ics.ComponentPropertyUrl),
		"categories":    propValue(event, ics.ComponentPropertyCategories),
		"exdate":        joinProps(event, ics.ComponentPropertyExdate),
		"recurrence_id": propValue(event, ics.ComponentPropertyRecurrenceId),
		"created":       propValue(event, ics.ComponentPropertyCreated),
		"last_modified": propValue(event, ics.ComponentPropertyLastModified),
		"sequence":      propValue(event, ics.ComponentPropertySequence),
		"transp":        propValue(event, ics.ComponentPropertyTransp),
	}
	for k, v := range preserve {
		if strings.TrimSpace(v) != "" {
			raw[k] = v
		}
	}
	known := map[string]struct{}{
		string(ics.ComponentPropertyUniqueId): {}, string(ics.ComponentPropertyDtStart): {},
		string(ics.ComponentPropertyDtEnd): {}, string(ics.ComponentPropertyDuration): {},
		string(ics.ComponentPropertySummary): {}, string(ics.ComponentPropertyDescription): {},
		string(ics.ComponentPropertyLocation): {}, string(ics.ComponentPropertyOrganizer): {},
		string(ics.ComponentPropertyAttendee): {}, string(ics.ComponentPropertyStatus): {},
		string(ics.ComponentPropertyClass): {}, string(ics.PropertyRrule): {},
		string(ics.ComponentPropertyExdate): {}, string(ics.ComponentPropertyRecurrenceId): {},
		string(ics.ComponentPropertyUrl): {}, string(ics.ComponentPropertyCategories): {},
		string(ics.ComponentPropertyCreated): {}, string(ics.ComponentPropertyLastModified): {},
		string(ics.ComponentPropertySequence): {}, string(ics.ComponentPropertyTransp): {},
	}
	extras := map[string]string{}
	for _, p := range event.Properties {
		token := strings.ToUpper(strings.TrimSpace(p.IANAToken))
		if token == "" {
			continue
		}
		if _, ok := known[token]; ok {
			continue
		}
		if strings.TrimSpace(p.Value) == "" {
			continue
		}
		extras[token] = p.Value
	}
	if len(extras) > 0 {
		raw["unsupported_properties"] = extras
	}

	updatedAt := time.Now().UTC()
	if lm := propValue(event, ics.ComponentPropertyLastModified); lm != "" {
		if t, perr := parseICSTimeValue(lm, false, ""); perr == nil {
			updatedAt = t
		}
	}

	recurrenceIDProp := propValue(event, ics.ComponentPropertyRecurrenceId)
	title := propValue(event, ics.ComponentPropertySummary)
	if title == "" {
		title = "(No title)"
	}

	recurringParent := ""
	if recurrenceIDProp != "" {
		recurringParent = uid
		raw["recurrence_id"] = recurrenceIDProp
	}

	return calendarprovider.RemoteEvent{
		ID:                   uid,
		Status:               status,
		Title:                title,
		Description:          propValue(event, ics.ComponentPropertyDescription),
		Location:             propValue(event, ics.ComponentPropertyLocation),
		OnlineMeetingURL:     propValue(event, ics.ComponentPropertyUrl),
		StartsAt:             start.UTC(),
		EndsAt:               end.UTC(),
		IsAllDay:             allDay,
		Timezone:             tz,
		Visibility:           visibility,
		OrganizerEmail:       organizerEmail,
		OrganizerDisplayName: organizerName,
		Attendees:            attendees,
		Recurrence:           recurrence,
		RecurringEventID:     recurringParent,
		UpdatedAt:            updatedAt,
		Deleted:              status == constant.CalendarEventStatusCancelled,
		Raw:                  raw,
	}, nil
}

func propValue(event *ics.VEvent, key ics.ComponentProperty) string {
	p := event.GetProperty(key)
	if p == nil {
		return ""
	}
	return strings.TrimSpace(p.Value)
}

func joinProps(event *ics.VEvent, key ics.ComponentProperty) string {
	var parts []string
	for _, p := range event.Properties {
		if p.IANAToken == string(key) && strings.TrimSpace(p.Value) != "" {
			parts = append(parts, strings.TrimSpace(p.Value))
		}
	}
	return strings.Join(parts, ",")
}

func parseEventTime(event *ics.VEvent, key ics.ComponentProperty) (time.Time, bool, string, error) {
	p := event.GetProperty(key)
	if p == nil {
		return time.Time{}, false, "", errEmptyCalendar
	}
	value := strings.TrimSpace(p.Value)
	params := p.ICalParameters
	allDay := false
	tzid := ""
	if vals, ok := params["VALUE"]; ok {
		for _, v := range vals {
			if strings.EqualFold(v, "DATE") {
				allDay = true
			}
		}
	}
	if vals, ok := params["TZID"]; ok && len(vals) > 0 {
		tzid = vals[0]
	}
	if !allDay && len(value) == 8 && !strings.Contains(value, "T") {
		allDay = true
	}
	t, err := parseICSTimeValue(value, allDay, tzid)
	return t, allDay, tzid, err
}

func parseICSTimeValue(value string, allDay bool, tzid string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if allDay || (len(value) == 8 && !strings.Contains(value, "T")) {
		t, err := time.ParseInLocation("20060102", value, time.UTC)
		return t, err
	}
	if strings.HasSuffix(value, "Z") {
		return time.Parse("20060102T150405Z", value)
	}
	loc := time.UTC
	if tzid != "" {
		loc = loadLocation(tzid)
	}
	if t, err := time.ParseInLocation("20060102T150405", value, loc); err == nil {
		return t, nil
	}
	return time.ParseInLocation("20060102T150405", value, time.UTC)
}

func parseICSDuration(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errEmptyCalendar
	}
	negative := false
	if strings.HasPrefix(raw, "-") {
		negative = true
		raw = raw[1:]
	}
	if !strings.HasPrefix(raw, "P") {
		return 0, errEmptyCalendar
	}
	raw = raw[1:]
	var total time.Duration
	inTime := false
	num := ""
	flush := func(unit byte) {
		if num == "" {
			return
		}
		var n int
		for _, c := range num {
			n = n*10 + int(c-'0')
		}
		num = ""
		d := time.Duration(n)
		switch unit {
		case 'W':
			total += d * 7 * 24 * time.Hour
		case 'D':
			total += d * 24 * time.Hour
		case 'H':
			total += d * time.Hour
		case 'M':
			if inTime {
				total += d * time.Minute
			} else {
				total += d * 30 * 24 * time.Hour // rough month fallback
			}
		case 'S':
			total += d * time.Second
		}
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c >= '0' && c <= '9':
			num += string(c)
		case c == 'T':
			flush(0)
			inTime = true
		case c == 'W' || c == 'D' || c == 'H' || c == 'M' || c == 'S':
			flush(c)
		}
	}
	if negative {
		total = -total
	}
	return total, nil
}

func parseCalAddress(raw string) (email, name string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	// CN may live in params; golang-ical puts CN on the property params when present.
	if idx := strings.Index(strings.ToUpper(raw), "MAILTO:"); idx >= 0 {
		email = strings.TrimSpace(raw[idx+7:])
	} else if strings.Contains(raw, "@") {
		email = raw
	}
	return email, name
}

func parseAttendees(event *ics.VEvent) []calendarprovider.RemoteAttendee {
	out := make([]calendarprovider.RemoteAttendee, 0)
	for _, p := range event.Properties {
		if p.IANAToken != string(ics.PropertyAttendee) {
			continue
		}
		email, _ := parseCalAddress(p.Value)
		name := ""
		if vals, ok := p.ICalParameters["CN"]; ok && len(vals) > 0 {
			name = vals[0]
		}
		status := ""
		if vals, ok := p.ICalParameters["PARTSTAT"]; ok && len(vals) > 0 {
			status = strings.ToLower(vals[0])
		}
		roleOrganizer := false
		if vals, ok := p.ICalParameters["ROLE"]; ok {
			for _, v := range vals {
				if strings.EqualFold(v, "CHAIR") {
					roleOrganizer = true
				}
			}
		}
		out = append(out, calendarprovider.RemoteAttendee{
			Email:          email,
			DisplayName:    name,
			ResponseStatus: status,
			Organizer:      roleOrganizer,
		})
	}
	return out
}
