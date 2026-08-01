package chat_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/chat"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

func parseWith(t *testing.T, input, tz string, now time.Time) *chat.Intent {
	t.Helper()
	p := chat.NewRuleBasedParser()
	ctx := chat.WithParseTimezone(chat.WithParseNow(context.Background(), now), tz)
	intent, err := p.Parse(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	return intent
}

func TestRuleBasedParserCreateTask(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	intent := parseWith(t, "Add task Finish Timeline UI", "UTC", now)
	if intent.Kind != chat.IntentCreateTask {
		t.Fatalf("kind = %s", intent.Kind)
	}
	if intent.Title != "Finish Timeline UI" {
		t.Fatalf("title = %q", intent.Title)
	}
}

func TestRuleBasedParserCreateReminderWeekly(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, loc)
	intent := parseWith(t, "Remind me every Friday at 8 PM to call Mom", "Asia/Kolkata", now)
	if intent.Kind != chat.IntentCreateReminder {
		t.Fatalf("kind = %s", intent.Kind)
	}
	if intent.Title != "call Mom" {
		t.Fatalf("title = %q", intent.Title)
	}
	if intent.RecurrenceRule == nil || !strings.Contains(*intent.RecurrenceRule, "BYDAY=FR") {
		t.Fatalf("rule = %v", intent.RecurrenceRule)
	}
	local := intent.TriggerAt.In(loc)
	if local.Weekday() != time.Friday || local.Hour() != 20 {
		t.Fatalf("trigger = %v", local)
	}
}

func TestRuleBasedParserMVPQueries(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cases := map[string]chat.IntentKind{
		"What do I have tomorrow?": chat.IntentQueryTomorrow,
		"what's on today":          chat.IntentQueryToday,
		"What's due today?":        chat.IntentQueryDueToday,
		"due today":                chat.IntentQueryDueToday,
	}
	for input, want := range cases {
		got := parseWith(t, input, "UTC", now)
		if got.Kind != want {
			t.Fatalf("%q => %s want %s", input, got.Kind, want)
		}
	}
}

func TestRuleBasedParserGreeting(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	for _, input := range []string{
		"hi",
		"Hello!",
		"hey donna",
		"good morning",
		"What's up",
		"yo",
	} {
		got := parseWith(t, input, "UTC", now)
		if got.Kind != chat.IntentGreeting {
			t.Fatalf("%q => %s want GREETING", input, got.Kind)
		}
	}
}

func TestRuleBasedParserRejectsNonMVP(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	// Intentionally unsupported in MVP — should be UNKNOWN, not half-parsed.
	for _, input := range []string{
		"delete task Finish API",
		"notifications",
		"show this week",
		"tell me a joke",
	} {
		got := parseWith(t, input, "UTC", now)
		if got.Kind != chat.IntentUnknown {
			t.Fatalf("%q => %s want UNKNOWN", input, got.Kind)
		}
	}
}

func TestRuleBasedParserCompleteTask(t *testing.T) {
	t.Parallel()
	c := parseWith(t, "complete task Finish API", "UTC", time.Now().UTC())
	if c.Kind != chat.IntentCompleteTask || c.TargetTitle != "Finish API" {
		t.Fatalf("complete = %+v", c)
	}
}

func TestRuleBasedParserCreateEvent(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)
	intent := parseWith(t, "Schedule meeting Standup tomorrow at 10 AM", "UTC", now)
	if intent.Kind != chat.IntentCreateEvent {
		t.Fatalf("kind = %s", intent.Kind)
	}
	if !strings.Contains(strings.ToLower(intent.Title), "standup") {
		t.Fatalf("title = %q", intent.Title)
	}
	if intent.StartAt == nil || intent.StartAt.Day() != 31 || intent.StartAt.Hour() != 10 {
		t.Fatalf("start = %v", intent.StartAt)
	}
}

type stubParser struct{ intent *chat.Intent }

func (s stubParser) Parse(context.Context, string) (*chat.Intent, error) { return s.intent, nil }

func TestExecutorGreeting(t *testing.T) {
	t.Parallel()
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{Kind: chat.IntentGreeting, Timezone: "UTC"}}, nil, nil)
	morning := time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		DisplayName: "Aryan Thacker",
		Timezone:    "UTC",
		Now:         morning,
		Message:     "hi",
	})
	if out.Intent != chat.IntentGreeting {
		t.Fatalf("intent = %s", out.Intent)
	}
	if !strings.Contains(out.Reply, "Good morning, Aryan") {
		t.Fatalf("reply = %q", out.Reply)
	}
}

func TestExecutorUnknown(t *testing.T) {
	t.Parallel()
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{Kind: chat.IntentUnknown}}, &actions.Registry{}, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		UserID:  uuid.MustParse("018f0000-0000-7000-8000-000000000501"),
		Message: "???",
	})
	if out.Intent != chat.IntentUnknown || !strings.Contains(out.Reply, "couldn't understand") {
		t.Fatalf("out = %+v", out)
	}
}

func TestExecutorNilParser(t *testing.T) {
	t.Parallel()
	ex := chat.NewExecutor(nil, &actions.Registry{}, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{Message: "hi"})
	if out.Intent != chat.IntentUnknown {
		t.Fatalf("intent = %s", out.Intent)
	}
}

type memTaskService struct {
	created bool
	title   string
}

func (m *memTaskService) CreateTask(_ context.Context, _ uuid.UUID, in business.CreateTaskInput) (entity.TaskOccurrenceWithTask, error) {
	m.created = true
	m.title = in.Title
	return entity.TaskOccurrenceWithTask{
		TaskOccurrence: entity.TaskOccurrence{ID: uuid.MustParse("018f0000-0000-7000-8000-000000000510")},
		Title:          in.Title,
	}, nil
}
func (m *memTaskService) UpdateTask(context.Context, uuid.UUID, uuid.UUID, business.UpdateTaskInput) (entity.Task, error) {
	return entity.Task{}, nil
}
func (m *memTaskService) UpdateOccurrence(context.Context, uuid.UUID, uuid.UUID, bool) (entity.TaskOccurrenceWithTask, error) {
	return entity.TaskOccurrenceWithTask{}, nil
}
func (m *memTaskService) DeleteTask(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *memTaskService) ListTaskTagsForTask(context.Context, uuid.UUID, uuid.UUID) ([]entity.TaskTag, error) {
	return nil, nil
}

func TestExecutorCreateTaskViaActions(t *testing.T) {
	t.Parallel()
	taskSvc := &memTaskService{}
	reg := &actions.Registry{CreateTask: actions.NewCreateTaskAction(taskSvc, nil)}
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{
		Kind: chat.IntentCreateTask, Title: "Finish Timeline UI", Timezone: "UTC",
	}}, reg, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000502"), Timezone: "UTC",
		Now: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), Message: "Add task Finish Timeline UI",
	})
	if !strings.Contains(out.Reply, "Task created") || !taskSvc.created {
		t.Fatalf("out=%+v svc=%+v", out, taskSvc)
	}
}

func TestExecutorDryRunSkipsMutation(t *testing.T) {
	t.Parallel()
	taskSvc := &memTaskService{}
	reg := &actions.Registry{CreateTask: actions.NewCreateTaskAction(taskSvc, nil)}
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{
		Kind: chat.IntentCreateTask, Title: "Finish Timeline UI", Timezone: "UTC",
	}}, reg, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000502"), Timezone: "UTC",
		Now: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), Message: "Add task Finish Timeline UI",
		DryRun: true,
	})
	if taskSvc.created {
		t.Fatal("dry-run must not create tasks")
	}
	if !strings.Contains(out.Reply, "Preview") || !strings.Contains(out.Reply, "nothing was saved") {
		t.Fatalf("reply = %q", out.Reply)
	}
}

type memTimelineService struct {
	items []entity.TimelineItem
}

func (m *memTimelineService) List(context.Context, uuid.UUID, time.Time, time.Time) ([]entity.TimelineItem, error) {
	return m.items, nil
}

func TestExecutorQueryTomorrow(t *testing.T) {
	t.Parallel()
	reg := &actions.Registry{
		QueryTimeline: actions.NewQueryTimelineAction(&memTimelineService{
			items: []entity.TimelineItem{
				{Title: "Standup", StartAt: time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC), Timezone: "UTC", Type: constant.TimelineTypeEvent},
				{Title: "Guitar Class", StartAt: time.Date(2026, 7, 31, 19, 0, 0, 0, time.UTC), Timezone: "UTC", Type: constant.TimelineTypeEvent},
			},
		}),
	}
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{Kind: chat.IntentQueryTomorrow, Timezone: "UTC"}}, reg, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000503"), Timezone: "UTC",
		Now: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), Message: "What do I have tomorrow?",
	})
	if !strings.Contains(out.Reply, "Standup") || !strings.Contains(out.Reply, "Guitar Class") {
		t.Fatalf("reply = %q", out.Reply)
	}
}

type memDayLister struct {
	occs []actions.TaskOccurrenceResult
}

func (m *memDayLister) ListDayTaskOccurrences(context.Context, uuid.UUID, time.Time) ([]actions.TaskOccurrenceResult, error) {
	return m.occs, nil
}

func TestExecutorQueryDueToday(t *testing.T) {
	t.Parallel()
	reg := &actions.Registry{
		ListDayTasks: actions.NewListDayTasksAction(&memDayLister{
			occs: []actions.TaskOccurrenceResult{
				{Title: "Finish API", Completed: false},
				{Title: "Done already", Completed: true},
			},
		}),
	}
	ex := chat.NewExecutor(stubParser{intent: &chat.Intent{Kind: chat.IntentQueryDueToday, Timezone: "UTC"}}, reg, nil)
	out := ex.Execute(context.Background(), chat.ExecuteInput{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000504"), Timezone: "UTC",
		Now: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), Message: "What's due today?",
	})
	if !strings.Contains(out.Reply, "Finish API") || strings.Contains(out.Reply, "Done already") {
		t.Fatalf("reply = %q", out.Reply)
	}
}
