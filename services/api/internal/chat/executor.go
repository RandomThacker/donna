package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/google/uuid"
)

// Executor runs parsed intents through the Action Layer.
// It depends on IntentParser — never on a concrete parser type.
// Personality rendering is presentation-only and never mutates business outcomes.
type Executor struct {
	parser     IntentParser
	reg        *actions.Registry
	personality personality.Renderer
}

// NewExecutor constructs a chat command executor.
// parser must be provided (RuleBasedParser today; OpenAI/Claude later).
// renderer may be nil (canonical replies only).
func NewExecutor(parser IntentParser, reg *actions.Registry, renderer personality.Renderer) *Executor {
	return &Executor{parser: parser, reg: reg, personality: renderer}
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
		return e.finalize(ctx, in, CommandResult{Reply: UnknownHelp, Intent: IntentUnknown}, now, tz)
	}
	if intent.Timezone == "" {
		intent.Timezone = tz
	}

	if intent.Kind == IntentUnknown {
		return e.finalize(ctx, in, CommandResult{Reply: UnknownHelp, Intent: IntentUnknown}, now, tz)
	}
	if isGreetingIntent(intent.Kind) {
		return e.finalize(ctx, in, CommandResult{Reply: "", Intent: intent.Kind}, now, tz)
	}
	if e.reg == nil {
		return e.finalize(ctx, in, CommandResult{Reply: "Donna isn't ready for commands yet.", Intent: intent.Kind}, now, tz)
	}

	reply, err := e.dispatch(ctx, in.UserID, now, *intent, in.DryRun)
	if err != nil {
		return e.finalize(ctx, in, CommandResult{
			Reply:  friendlyErr(err),
			Intent: intent.Kind,
			Error:  friendlyErr(err),
		}, now, tz)
	}
	return e.finalize(ctx, in, CommandResult{Reply: reply, Intent: intent.Kind}, now, tz)
}

func (e *Executor) finalize(
	ctx context.Context,
	in ExecuteInput,
	result CommandResult,
	now time.Time,
	tz string,
) CommandResult {
	if in.SkipPersonality || e == nil || e.personality == nil {
		// Automations personalize the combined reply — do not invent a
		// display-name fallback greeting here (it becomes "Good evening, there.").
		if !in.SkipPersonality && isGreetingIntent(result.Intent) && strings.TrimSpace(result.Reply) == "" {
			result.Reply = fallbackGreeting(in.DisplayName, now, tz)
		}
		if result.Error != "" && strings.TrimSpace(result.Reply) != "" && !strings.HasPrefix(result.Reply, "I couldn't") {
			result.Reply = fmt.Sprintf("I couldn't do that. %s", result.Reply)
		}
		return result
	}
	kind := kindForIntent(result.Intent, result.Error != "")
	canonical := strings.TrimSpace(result.Reply)
	if kind == personality.KindError && canonical == "" {
		canonical = result.Error
	}
	out, err := e.personality.Render(ctx, personality.RenderInput{
		UserID:    in.UserID,
		Canonical: canonical,
		Kind:      kind,
		Now:       now,
		Timezone:  tz,
	})
	if err != nil || strings.TrimSpace(out.Text) == "" {
		if canonical == "" && isGreetingIntent(result.Intent) {
			result.Reply = fallbackGreeting(in.DisplayName, now, tz)
			return result
		}
		if canonical != "" {
			result.Reply = canonical
		}
		return result
	}
	result.Reply = out.Text
	return result
}

func isGreetingIntent(intent IntentKind) bool {
	switch intent {
	case IntentGreeting, IntentMorningGreeting, IntentEveningGreeting, IntentGoodNightGreeting:
		return true
	default:
		return false
	}
}

func kindForIntent(intent IntentKind, isError bool) personality.Kind {
	if isError {
		return personality.KindError
	}
	switch intent {
	case IntentGreeting:
		return personality.KindGreeting
	case IntentMorningGreeting:
		return personality.KindMorningGreeting
	case IntentEveningGreeting:
		return personality.KindEveningGreeting
	case IntentGoodNightGreeting:
		return personality.KindGoodNight
	case IntentCompleteTask:
		return personality.KindTaskComplete
	case IntentCreateTask, IntentCreateReminder, IntentCreateEvent:
		return personality.KindAcknowledgement
	case IntentUnknown:
		return personality.KindError
	default:
		return personality.KindChat
	}
}

func fallbackGreeting(displayName string, now time.Time, tz string) string {
	name := firstName(displayName)
	if name == "" {
		name = "there"
	}
	loc := loadLocation(tz)
	local := now.In(loc)
	switch {
	case local.Hour() >= 5 && local.Hour() < 12:
		return fmt.Sprintf("Good morning, %s.", name)
	case local.Hour() >= 12 && local.Hour() < 17:
		return fmt.Sprintf("Good afternoon, %s.", name)
	case local.Hour() >= 17 && local.Hour() < 21:
		return fmt.Sprintf("Good evening, %s.", name)
	default:
		return fmt.Sprintf("Hello, %s.", name)
	}
}

func (e *Executor) dispatch(ctx context.Context, userID uuid.UUID, now time.Time, intent Intent, dryRun bool) (string, error) {
	switch intent.Kind {
	case IntentCreateTask:
		if dryRun {
			return dryRunMutationReply("create task", intent.Title), nil
		}
		return e.createTask(ctx, userID, now, intent)
	case IntentCompleteTask:
		if dryRun {
			return dryRunMutationReply("complete task", intent.TargetTitle), nil
		}
		return e.completeTask(ctx, userID, now, intent)
	case IntentCreateReminder:
		if dryRun {
			return dryRunMutationReply("create reminder", intent.Title), nil
		}
		return e.createReminder(ctx, userID, intent)
	case IntentCreateEvent:
		if dryRun {
			return dryRunMutationReply("create event", intent.Title), nil
		}
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
		return "", nil
	default:
		return UnknownHelp, nil
	}
}

func dryRunMutationReply(action, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Sprintf("Preview: would %s (nothing was saved).", action)
	}
	return fmt.Sprintf("Preview: would %s %q (nothing was saved).", action, title)
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
