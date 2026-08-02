package chat

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RuleBasedParser maps a small set of phrases to MVP intents.
// Later replace via IntentParser: OpenAIParser, ClaudeParser, GeminiParser.
type RuleBasedParser struct{}

// NewRuleBasedParser constructs the default Phase 3 parser.
func NewRuleBasedParser() *RuleBasedParser {
	return &RuleBasedParser{}
}

var (
	reCreateTask   = regexp.MustCompile(`(?i)^\s*(?:please\s+)?(?:add|create|make)?\s*(?:a\s+)?(?:new\s+)?task\s+(.+)$`)
	reTodo         = regexp.MustCompile(`(?i)^\s*(?:todo|to-do)\s+(.+)$`)
	reCompleteTask = regexp.MustCompile(`(?i)^\s*(?:please\s+)?(?:complete|finish|done|check off)\s+(?:the\s+)?(?:task\s+)?(.+)$`)
	reMarkDone     = regexp.MustCompile(`(?i)^\s*mark\s+(.+?)\s+(?:as\s+)?done\s*$`)

	reRemindMe       = regexp.MustCompile(`(?i)^\s*(?:please\s+)?remind\s+me\s+(.+)$`)
	reCreateReminder = regexp.MustCompile(`(?i)^\s*(?:please\s+)?(?:add|create)\s+(?:a\s+)?reminder\s+(?:to\s+)?(.+)$`)

	reCreateEvent = regexp.MustCompile(`(?i)^\s*(?:please\s+)?(?:schedule|add|create)\s+(?:an?\s+)?(?:event|meeting)\s+(.+)$`)
	reMeeting     = regexp.MustCompile(`(?i)^\s*(?:please\s+)?meeting\s+(.+)$`)

	reQueryToday = regexp.MustCompile(`(?i)^\s*(?:what(?:'s| is| do i have)?\s+(?:on\s+)?today|today(?:'s)?\s+(?:schedule|agenda|plan)|show\s+today)\s*[?.!]?\s*$`)
	reQueryTomorrow = regexp.MustCompile(`(?i)^\s*(?:what(?:'s| is| do i have)?\s+(?:on\s+)?tomorrow|tomorrow(?:'s)?\s+(?:schedule|agenda|plan)|show\s+tomorrow)\s*[?.!]?\s*$`)
	reQueryDueToday = regexp.MustCompile(`(?i)^\s*(?:what(?:'s| is)\s+due(?:\s+today)?|due\s+today|what\s+tasks?\s+(?:do i have\s+)?(?:today|due)|show\s+(?:my\s+)?(?:tasks|todos?)(?:\s+today)?)\s*[?.!]?\s*$`)
	reQueryBacklog  = regexp.MustCompile(`(?i)^\s*(?:` +
		`(?:show\s+(?:me\s+)?(?:my\s+)?)?backlog|` +
		`(?:what(?:'s| is)|show)\s+(?:my\s+)?(?:task\s+)?backlog|` +
		`task\s+backlog|` +
		`today(?:'s)?\s+backlog|` +
		`how(?:'s| is)\s+(?:my\s+)?backlog` +
		`)\s*[?.!]?\s*$`)

	reMorningGreeting = regexp.MustCompile(`(?i)^\s*(?:good\s+morning|morning\s+greeting)\s*[!.]*\s*$`)
	reEveningGreeting = regexp.MustCompile(`(?i)^\s*(?:good\s+evening|evening\s+greeting|how\s+was\s+(?:my\s+)?(?:the\s+)?day)\s*[!.]*\s*$`)
	reGoodNight       = regexp.MustCompile(`(?i)^\s*(?:good\s*night|goodnight|night\s+greeting)\s*[!.]*\s*$`)

	reGreeting = regexp.MustCompile(`(?i)^\s*(?:` +
		`hi+|hello|hey+|heya|hiya|howdy|yo|sup|hola|greetings|` +
		`good\s+(?:afternoon|day)|` +
		`what'?s\s+up|whats\s+up` +
		`)` +
		`(?:\s+(?:there|donna|folks|everyone|all))?` +
		`\s*[!.]*\s*$`)

	reEveryWeekday = regexp.MustCompile(`(?i)\bevery\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b`)
	reAtTime       = regexp.MustCompile(`(?i)\bat\s+(\d{1,2})(?::(\d{2}))?\s*(am|pm)?\b`)
	reToClause     = regexp.MustCompile(`(?i)\bto\s+(.+)$`)
)

// Parse implements IntentParser.
func (p *RuleBasedParser) Parse(ctx context.Context, input string) (*Intent, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return &Intent{Kind: IntentUnknown, Raw: input, Confidence: 0}, nil
	}

	tz := parseTimezoneFrom(ctx)
	loc := loadLocation(tz)
	now := parseNowFrom(ctx).In(loc)
	lower := strings.ToLower(raw)

	switch {
	case reMorningGreeting.MatchString(raw):
		return &Intent{Kind: IntentMorningGreeting, Raw: raw, Timezone: loc.String(), Confidence: 0.99}, nil
	case reEveningGreeting.MatchString(raw):
		return &Intent{Kind: IntentEveningGreeting, Raw: raw, Timezone: loc.String(), Confidence: 0.99}, nil
	case reGoodNight.MatchString(raw):
		return &Intent{Kind: IntentGoodNightGreeting, Raw: raw, Timezone: loc.String(), Confidence: 0.99}, nil
	case reGreeting.MatchString(raw):
		return &Intent{Kind: IntentGreeting, Raw: raw, Timezone: loc.String(), Confidence: 0.99}, nil
	case reQueryBacklog.MatchString(raw):
		return &Intent{Kind: IntentQueryBacklog, Raw: raw, Timezone: loc.String(), Confidence: 0.95}, nil
	case reQueryDueToday.MatchString(raw):
		return &Intent{Kind: IntentQueryDueToday, Raw: raw, Timezone: loc.String(), Confidence: 0.95}, nil
	case reQueryToday.MatchString(raw):
		return &Intent{Kind: IntentQueryToday, Raw: raw, Timezone: loc.String(), Confidence: 0.95}, nil
	case reQueryTomorrow.MatchString(raw):
		return &Intent{Kind: IntentQueryTomorrow, Raw: raw, Timezone: loc.String(), Confidence: 0.95}, nil
	}

	if m := reCompleteTask.FindStringSubmatch(raw); m != nil {
		done := true
		return &Intent{
			Kind: IntentCompleteTask, Raw: raw, TargetTitle: cleanTitle(m[1]),
			Completed: &done, Timezone: loc.String(), Confidence: 0.9,
		}, nil
	}
	if m := reMarkDone.FindStringSubmatch(raw); m != nil {
		done := true
		return &Intent{
			Kind: IntentCompleteTask, Raw: raw, TargetTitle: cleanTitle(m[1]),
			Completed: &done, Timezone: loc.String(), Confidence: 0.9,
		}, nil
	}
	if m := reCreateTask.FindStringSubmatch(raw); m != nil {
		return &Intent{Kind: IntentCreateTask, Raw: raw, Title: cleanTitle(m[1]), Timezone: loc.String(), Confidence: 0.95}, nil
	}
	if m := reTodo.FindStringSubmatch(raw); m != nil {
		return &Intent{Kind: IntentCreateTask, Raw: raw, Title: cleanTitle(m[1]), Timezone: loc.String(), Confidence: 0.9}, nil
	}

	if m := reRemindMe.FindStringSubmatch(raw); m != nil {
		return parseCreateReminder(raw, m[1], now, loc), nil
	}
	if m := reCreateReminder.FindStringSubmatch(raw); m != nil {
		return parseCreateReminder(raw, m[1], now, loc), nil
	}

	if m := reCreateEvent.FindStringSubmatch(raw); m != nil {
		return parseCreateEvent(raw, m[1], now, loc), nil
	}
	if m := reMeeting.FindStringSubmatch(raw); m != nil {
		return parseCreateEvent(raw, m[1], now, loc), nil
	}

	if strings.Contains(lower, "tomorrow") && (strings.Contains(lower, "have") || strings.Contains(lower, "schedule") || strings.Contains(lower, "agenda")) {
		return &Intent{Kind: IntentQueryTomorrow, Raw: raw, Timezone: loc.String(), Confidence: 0.7}, nil
	}
	if strings.Contains(lower, "today") && (strings.Contains(lower, "have") || strings.Contains(lower, "schedule") || strings.Contains(lower, "agenda")) {
		return &Intent{Kind: IntentQueryToday, Raw: raw, Timezone: loc.String(), Confidence: 0.7}, nil
	}
	if strings.Contains(lower, "backlog") {
		return &Intent{Kind: IntentQueryBacklog, Raw: raw, Timezone: loc.String(), Confidence: 0.7}, nil
	}
	if strings.Contains(lower, "due") && strings.Contains(lower, "today") {
		return &Intent{Kind: IntentQueryDueToday, Raw: raw, Timezone: loc.String(), Confidence: 0.7}, nil
	}

	return &Intent{Kind: IntentUnknown, Raw: raw, Confidence: 0}, nil
}

func parseCreateReminder(raw, body string, now time.Time, loc *time.Location) *Intent {
	body = strings.TrimSpace(body)
	title := body
	var rule *string
	trigger := now.Add(time.Hour)

	if m := reEveryWeekday.FindStringSubmatch(body); m != nil {
		byday := weekdayToBYDAY(m[1])
		r := fmt.Sprintf("FREQ=WEEKLY;BYDAY=%s", byday)
		rule = &r
		trigger = nextWeekdayAt(now, loc, m[1], extractClock(body, 20, 0))
		if tm := reToClause.FindStringSubmatch(body); tm != nil {
			title = cleanTitle(tm[1])
		} else {
			title = cleanTitle(stripTemporalPhrases(body))
		}
	} else {
		trigger = resolveDateTime(body, now, loc, 9, 0)
		if tm := reToClause.FindStringSubmatch(body); tm != nil {
			title = cleanTitle(tm[1])
		} else {
			title = cleanTitle(stripTemporalPhrases(body))
		}
	}
	if title == "" {
		title = "Reminder"
	}
	return &Intent{
		Kind: IntentCreateReminder, Raw: raw, Title: title,
		TriggerAt: &trigger, Timezone: loc.String(), RecurrenceRule: rule, Confidence: 0.9,
	}
}

func parseCreateEvent(raw, body string, now time.Time, loc *time.Location) *Intent {
	body = strings.TrimSpace(body)
	start := resolveDateTime(body, now, loc, 10, 0)
	end := start.Add(time.Hour)
	title := cleanTitle(stripTemporalPhrases(body))
	if title == "" {
		title = "Event"
	}
	return &Intent{
		Kind: IntentCreateEvent, Raw: raw, Title: title,
		StartAt: &start, EndAt: &end, Timezone: loc.String(), Confidence: 0.9,
	}
}

func resolveDateTime(s string, now time.Time, loc *time.Location, defaultHour, defaultMin int) time.Time {
	c := extractClock(s, defaultHour, defaultMin)
	hour, min := c.hour, c.min
	day := now
	lower := strings.ToLower(s)
	switch {
	case strings.Contains(lower, "day after tomorrow"):
		day = now.AddDate(0, 0, 2)
	case strings.Contains(lower, "tomorrow"):
		day = now.AddDate(0, 0, 1)
	case strings.Contains(lower, "today"):
		day = now
	default:
		if m := reEveryWeekday.FindStringSubmatch(s); m != nil {
			return nextWeekdayAt(now, loc, m[1], clock{hour, min})
		}
		for i, name := range weekdays {
			if strings.Contains(lower, name) {
				return nextWeekdayAt(now, loc, weekdays[i], clock{hour, min})
			}
		}
	}
	y, m, d := day.Date()
	return time.Date(y, m, d, hour, min, 0, 0, loc)
}

type clock struct{ hour, min int }

func extractClock(s string, defaultHour, defaultMin int) clock {
	m := reAtTime.FindStringSubmatch(s)
	if m == nil {
		return clock{defaultHour, defaultMin}
	}
	hour := atoiDefault(m[1], defaultHour)
	min := 0
	if m[2] != "" {
		min = atoiDefault(m[2], 0)
	}
	ampm := strings.ToLower(m[3])
	if ampm == "pm" && hour < 12 {
		hour += 12
	}
	if ampm == "am" && hour == 12 {
		hour = 0
	}
	return clock{hour, min}
}

func nextWeekdayAt(now time.Time, loc *time.Location, weekday string, c clock) time.Time {
	want := weekdayNameToTime(weekday)
	n := now.In(loc)
	for i := 0; i < 8; i++ {
		candidate := n.AddDate(0, 0, i)
		if candidate.Weekday() != want {
			continue
		}
		y, m, d := candidate.Date()
		t := time.Date(y, m, d, c.hour, c.min, 0, 0, loc)
		if i == 0 && !t.After(n) {
			continue
		}
		return t
	}
	y, m, d := n.AddDate(0, 0, 7).Date()
	return time.Date(y, m, d, c.hour, c.min, 0, 0, loc)
}

var weekdays = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

func weekdayToBYDAY(name string) string {
	switch strings.ToLower(name) {
	case "monday":
		return "MO"
	case "tuesday":
		return "TU"
	case "wednesday":
		return "WE"
	case "thursday":
		return "TH"
	case "friday":
		return "FR"
	case "saturday":
		return "SA"
	case "sunday":
		return "SU"
	default:
		return "MO"
	}
}

func weekdayNameToTime(name string) time.Weekday {
	switch strings.ToLower(name) {
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	case "sunday":
		return time.Sunday
	default:
		return time.Monday
	}
}

func stripTemporalPhrases(s string) string {
	out := s
	patterns := []*regexp.Regexp{
		reEveryWeekday,
		reAtTime,
		regexp.MustCompile(`(?i)\b(?:today|tomorrow|day after tomorrow)\b`),
		regexp.MustCompile(`(?i)\bon\s+(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b`),
		regexp.MustCompile(`(?i)\bto\s+`),
	}
	for _, p := range patterns {
		out = p.ReplaceAllString(out, " ")
	}
	return cleanTitle(out)
}

func cleanTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".,!?;:\"'")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func loadLocation(tz string) *time.Location {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		tz = "Asia/Kolkata"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.FixedZone("IST", 5*3600+30*60)
	}
	return loc
}

func atoiDefault(s string, def int) int {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
