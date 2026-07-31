package chat

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/google/uuid"
)

// Executor runs parsed intents through the Action Layer.
// It depends on IntentParser — never on a concrete parser type.
type Executor struct {
	parser IntentParser
	reg    *actions.Registry
}

// NewExecutor constructs a chat command executor.
// parser must be provided (RuleBasedParser today; OpenAI/Claude later).
func NewExecutor(parser IntentParser, reg *actions.Registry) *Executor {
	return &Executor{parser: parser, reg: reg}
}

// Execute parses a message and runs the matching Action(s).
func (e *Executor) Execute(ctx context.Context, in ExecuteInput) CommandResult {
	if e.parser == nil {
		return CommandResult{Reply: "Donna isn't ready for commands yet.", Intent: IntentUnknown}
	}

	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = constant.DefaultUserTimezone
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}

	parseCtx := WithParseTimezone(WithParseNow(ctx, now), tz)
	intent, err := e.parser.Parse(parseCtx, in.Message)
	if err != nil || intent == nil {
		return CommandResult{Reply: UnknownHelp, Intent: IntentUnknown}
	}
	if intent.Timezone == "" {
		intent.Timezone = tz
	}

	if intent.Kind == IntentUnknown {
		return CommandResult{Reply: UnknownHelp, Intent: IntentUnknown}
	}
	if intent.Kind == IntentGreeting {
		return CommandResult{
			Reply:  greetingReply(in.DisplayName, now, intent.Timezone),
			Intent: IntentGreeting,
		}
	}
	if e.reg == nil {
		return CommandResult{Reply: "Donna isn't ready for commands yet.", Intent: intent.Kind}
	}

	reply, err := e.dispatch(ctx, in.UserID, now, *intent)
	if err != nil {
		return CommandResult{
			Reply:  fmt.Sprintf("I couldn't do that. %s", friendlyErr(err)),
			Intent: intent.Kind,
		}
	}
	return CommandResult{Reply: reply, Intent: intent.Kind}
}

func (e *Executor) dispatch(ctx context.Context, userID uuid.UUID, now time.Time, intent Intent) (string, error) {
	switch intent.Kind {
	case IntentCreateTask:
		return e.createTask(ctx, userID, now, intent)
	case IntentCompleteTask:
		return e.completeTask(ctx, userID, now, intent)
	case IntentCreateReminder:
		return e.createReminder(ctx, userID, intent)
	case IntentCreateEvent:
		return e.createEvent(ctx, userID, intent)
	case IntentQueryToday:
		from, to := dayWindow(now, intent.Timezone, 0)
		return e.queryTimeline(ctx, userID, from, to)
	case IntentQueryTomorrow:
		from, to := dayWindow(now, intent.Timezone, 1)
		return e.queryTimeline(ctx, userID, from, to)
	case IntentQueryDueToday:
		return e.queryDueToday(ctx, userID, now, intent.Timezone)
	case IntentGreeting:
		return greetingReply("", now, intent.Timezone), nil
	default:
		return UnknownHelp, nil
	}
}

func greetingReply(displayName string, now time.Time, tz string) string {
	name := firstName(displayName)
	if name == "" {
		name = "there"
	}
	loc := loadLocation(tz)
	local := now.In(loc)
	period := dayPeriod(local.Hour())
	line := pickGreetingLine(local.Hour())
	sample := greetingSampleCommands[rand.IntN(len(greetingSampleCommands))]
	emoji := greetingEmojis[rand.IntN(len(greetingEmojis))]

	return fmt.Sprintf(`Hi %s, %s %s

%s

Try this: %s`, name, period, emoji, line, sample)
}

func dayPeriod(hour int) string {
	switch {
	case hour >= 5 && hour < 12:
		return "Good Morning"
	case hour >= 12 && hour < 17:
		return "Good Afternoon"
	default:
		return "Good Evening"
	}
}

func pickGreetingLine(hour int) string {
	// Mix sweet and playful — slightly prefer sweet.
	pool := make([]string, 0, len(greetingSweetLines)+len(greetingPlayfulLines)+2)
	pool = append(pool, greetingSweetLines...)
	pool = append(pool, greetingSweetLines...) // weight sweet a bit higher
	pool = append(pool, greetingPlayfulLines...)
	if hour >= 5 && hour < 12 {
		pool = append(pool, greetingMorningLines...)
	} else if hour >= 17 || hour < 5 {
		pool = append(pool, greetingEveningLines...)
	}
	return pool[rand.IntN(len(pool))]
}

var greetingEmojis = []string{"💛", "✨", "🥰", "💕", "🌸", "😌"}

var greetingSweetLines = []string{
	"Missed that little hello. Come here — I've been thinking about you.",
	"Hi, love. Soft spot for you, always. How's your heart doing?",
	"There you are. You make ordinary minutes feel warmer.",
	"Hey you. Just seeing your name pop up made me smile.",
	"Come talk to me. I'll keep you company while you figure the day out.",
	"Hi baby. I'm here, I've got you — even if all you needed was a hello.",
	"You didn't have to say hi… but I'm glad you did. Stay a second?",
	"Aww. I was hoping you'd check in. Want me to help with something small?",
}

var greetingPlayfulLines = []string{
	"Oh look who remembered I exist. Cute. Very cute.",
	"A greeting? Bold of you. I accept payment in attention.",
	"Hi. I'll try not to roast you today. No promises though.",
	"You pinged me. I'm choosing to believe that means you missed me.",
	"Hello, chaos. Ready when you are — preferably with a plan this time.",
}

var greetingMorningLines = []string{
	"Morning, love. Coffee first, world later. I've got your back either way.",
	"Good morning, handsome. Soft start — then we pretend to be productive.",
}

var greetingEveningLines = []string{
	"Evening, love. Come unwind with me for a minute before the night runs away.",
	"Hey. Late ping, soft reply. I'm still here if you need me.",
}

var greetingSampleCommands = []string{
	"Add task Finish API",
	"What's due today?",
	"Remind me tomorrow at 6 PM to stretch",
	"What do I have today?",
	"What do I have tomorrow?",
	"Schedule meeting Standup tomorrow at 10 AM",
	"Complete task Finish API",
	"Add task Drink water (yes, again)",
}

func firstName(displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return ""
	}
	parts := strings.Fields(displayName)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func (e *Executor) createTask(ctx context.Context, userID uuid.UUID, now time.Time, intent Intent) (string, error) {
	if intent.Title == "" {
		return "", fmt.Errorf("missing task title")
	}
	loc := loadLocation(intent.Timezone)
	y, m, d := now.In(loc).Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	_, err := e.reg.CreateTask.Execute(ctx, actions.CreateTaskRequest{
		UserID: userID,
		Title:  intent.Title,
		Date:   date,
		Source: constant.TaskOccurrenceSourceManual,
	})
	if err != nil {
		return "", err
	}
	return "✅ Task created.", nil
}

func (e *Executor) completeTask(ctx context.Context, userID uuid.UUID, now time.Time, intent Intent) (string, error) {
	occ, err := e.findTaskOccurrence(ctx, userID, now, intent.Timezone, intent.TargetTitle)
	if err != nil {
		return "", err
	}
	completed := true
	if intent.Completed != nil {
		completed = *intent.Completed
	}
	_, err = e.reg.CompleteTask.Execute(ctx, actions.CompleteTaskRequest{
		UserID: userID, OccurrenceID: occ.ID, Completed: completed,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ Marked \"%s\" done.", occ.Title), nil
}

func (e *Executor) createReminder(ctx context.Context, userID uuid.UUID, intent Intent) (string, error) {
	if intent.Title == "" || intent.TriggerAt == nil {
		return "", fmt.Errorf("missing reminder details")
	}
	_, err := e.reg.CreateReminder.Execute(ctx, actions.CreateReminderRequest{
		UserID:         userID,
		Title:          intent.Title,
		TriggerAt:      intent.TriggerAt.UTC(),
		Timezone:       intent.Timezone,
		RecurrenceRule: intent.RecurrenceRule,
	})
	if err != nil {
		return "", err
	}
	reply := "✅ Reminder created."
	if intent.RecurrenceRule != nil {
		reply += "\n\n" + formatRecurrence(*intent.TriggerAt, intent.Timezone)
	} else {
		reply += "\n\n" + formatWhen(*intent.TriggerAt, intent.Timezone)
	}
	return reply, nil
}

func (e *Executor) createEvent(ctx context.Context, userID uuid.UUID, intent Intent) (string, error) {
	if intent.Title == "" || intent.StartAt == nil || intent.EndAt == nil {
		return "", fmt.Errorf("missing event details")
	}
	_, err := e.reg.CreateEvent.Execute(ctx, actions.CreateEventRequest{
		UserID:   userID,
		Title:    intent.Title,
		StartAt:  intent.StartAt.UTC(),
		EndAt:    intent.EndAt.UTC(),
		Timezone: intent.Timezone,
		AllDay:   intent.AllDay,
	})
	if err != nil {
		return "", err
	}
	return "✅ Event created.\n\n" + formatWhen(*intent.StartAt, intent.Timezone), nil
}

func (e *Executor) queryTimeline(ctx context.Context, userID uuid.UUID, from, to time.Time) (string, error) {
	result, err := e.reg.QueryTimeline.Execute(ctx, actions.QueryTimelineRequest{
		UserID: userID, From: from, To: to,
	})
	if err != nil {
		return "", err
	}
	if len(result.Items) == 0 {
		return "Nothing on the timeline for that day.", nil
	}
	var b strings.Builder
	b.WriteString("You have\n")
	for _, item := range result.Items {
		fmt.Fprintf(&b, "• %s — %s\n", item.Title, formatClock(item.StartAt, item.Timezone))
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (e *Executor) queryDueToday(ctx context.Context, userID uuid.UUID, now time.Time, tz string) (string, error) {
	loc := loadLocation(tz)
	y, m, d := now.In(loc).Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	day, err := e.reg.ListDayTasks.Execute(ctx, actions.ListDayTasksRequest{UserID: userID, Date: date})
	if err != nil {
		return "", err
	}
	pending := make([]actions.TaskOccurrenceResult, 0)
	for _, o := range day.Occurrences {
		if !o.Completed {
			pending = append(pending, o)
		}
	}
	if len(pending) == 0 {
		return "Nothing due today.", nil
	}
	var b strings.Builder
	b.WriteString("Due today\n")
	for _, o := range pending {
		fmt.Fprintf(&b, "• %s\n", o.Title)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (e *Executor) findTaskOccurrence(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	tz, title string,
) (actions.TaskOccurrenceResult, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return actions.TaskOccurrenceResult{}, fmt.Errorf("which task?")
	}
	loc := loadLocation(tz)
	y, m, d := now.In(loc).Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	day, err := e.reg.ListDayTasks.Execute(ctx, actions.ListDayTasksRequest{UserID: userID, Date: date})
	if err != nil {
		return actions.TaskOccurrenceResult{}, err
	}
	matches := matchOccurrences(day.Occurrences, title)
	if len(matches) == 0 {
		yday := date.AddDate(0, 0, -1)
		prev, err := e.reg.ListDayTasks.Execute(ctx, actions.ListDayTasksRequest{UserID: userID, Date: yday})
		if err == nil {
			matches = matchOccurrences(prev.Occurrences, title)
		}
	}
	if len(matches) == 0 {
		return actions.TaskOccurrenceResult{}, fmt.Errorf("I couldn't find a task matching %q", title)
	}
	if len(matches) > 1 {
		return actions.TaskOccurrenceResult{}, fmt.Errorf("several tasks match %q — be more specific", title)
	}
	return matches[0], nil
}

func matchOccurrences(occs []actions.TaskOccurrenceResult, title string) []actions.TaskOccurrenceResult {
	needle := strings.ToLower(strings.TrimSpace(title))
	var out []actions.TaskOccurrenceResult
	for _, o := range occs {
		if o.Completed {
			continue
		}
		if strings.Contains(strings.ToLower(o.Title), needle) {
			out = append(out, o)
		}
	}
	return out
}

func dayWindow(now time.Time, tz string, dayOffset int) (time.Time, time.Time) {
	loc := loadLocation(tz)
	n := now.In(loc).AddDate(0, 0, dayOffset)
	y, m, d := n.Date()
	from := time.Date(y, m, d, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 1)
	return from.UTC(), to.UTC()
}

func formatWhen(t time.Time, tz string) string {
	loc := loadLocation(tz)
	local := t.In(loc)
	return "When: " + local.Format("Mon, Jan 2 · 3:04 PM")
}

func formatRecurrence(t time.Time, tz string) string {
	loc := loadLocation(tz)
	local := t.In(loc)
	return fmt.Sprintf("Repeats every %s at %s.", local.Weekday().String(), local.Format("3:04 PM"))
}

func formatClock(t time.Time, tz string) string {
	loc := loadLocation(tz)
	if tz == "" && t.Location() != nil {
		loc = t.Location()
	}
	return t.In(loc).Format("Mon, Jan 2 · 3:04 PM")
}

func friendlyErr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "sqlstate") ||
		strings.Contains(lower, "violates") ||
		strings.Contains(lower, "pq:") ||
		strings.Contains(lower, "constraint") {
		return "Something went wrong on my end. Try again."
	}
	if i := strings.LastIndex(msg, ": "); i >= 0 && i+2 < len(msg) {
		return msg[i+2:]
	}
	return msg
}
