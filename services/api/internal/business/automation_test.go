package business

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestAutomationDueMinuteAndIdempotency(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 1, 9, 0, 12, 0, loc)
	auto := entity.Automation{
		Enabled:     true,
		TriggerType: constant.AutomationTriggerDaily,
		Timezone:    "Asia/Kolkata",
		TriggerTime: "09:00",
		Commands:    []entity.AutomationCommand{{Command: constant.AutomationCommandTasksDue, Variables: map[string]string{"priority": "all"}}},
	}
	if !AutomationDue(auto, now) {
		t.Fatal("expected due at matching minute")
	}
	if AutomationDue(auto, now.Add(time.Minute)) {
		t.Fatal("expected not due one minute later")
	}
	fired := now.UTC()
	auto.LastRunAt = &fired
	if AutomationDue(auto, now.Add(30*time.Second)) {
		t.Fatal("expected not due after same-day run")
	}
	nextDay := now.Add(24 * time.Hour)
	if !AutomationDue(auto, nextDay) {
		t.Fatal("expected due again next civil day")
	}
}

func TestAutomationDueDisabled(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	auto := entity.Automation{
		Enabled:     false,
		TriggerType: constant.AutomationTriggerDaily,
		Timezone:    "UTC",
		TriggerTime: "09:00",
		Commands:    []entity.AutomationCommand{{Command: constant.AutomationCommandGreeting}},
	}
	if AutomationDue(auto, now) {
		t.Fatal("disabled automation must not be due")
	}
}

func TestAutomationServiceCreateValidation(t *testing.T) {
	t.Parallel()
	repo := newMemAutomationRepo()
	svc := NewAutomationService(repo)
	svc.now = func() time.Time { return time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC) }
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000401")

	_, err := svc.Create(context.Background(), userID, CreateAutomationInput{
		Name:        "Empty",
		Timezone:    "UTC",
		TriggerTime: "08:00",
		Commands:    nil,
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected validation, got %v", err)
	}

	created, err := svc.Create(context.Background(), userID, CreateAutomationInput{
		TemplateID: strPtr("morning_brief"),
		Timezone:   "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Morning Brief" {
		t.Fatalf("name = %s", created.Name)
	}
	if len(created.Commands) < 2 {
		t.Fatalf("expected multi-command template, got %#v", created.Commands)
	}
	if created.TriggerType != constant.AutomationTriggerDaily {
		t.Fatalf("trigger = %s", created.TriggerType)
	}
	if created.NextRunAt == nil {
		t.Fatal("expected next_run_at")
	}
}

func TestAutomationSchedulerMultiCommandCombinedReply(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000402")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000403")
	now := time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	repo.byID[autoID] = entity.Automation{
		ID:               autoID,
		PublicID:         "atm_test",
		UserID:           userID,
		Name:             "Morning Brief",
		Enabled:          true,
		TriggerType:      constant.AutomationTriggerDaily,
		TriggerTime:      "07:00",
		Timezone:         "UTC",
		Commands:         []entity.AutomationCommand{
			{Command: constant.AutomationCommandTodaysAgenda, Variables: map[string]string{"range": "today"}},
			{Command: constant.AutomationCommandTasksDue, Variables: map[string]string{"priority": "all"}},
		},
		DeliveryChannels: []string{constant.AutomationDeliveryChat},
	}

	chatExec := &stubChatExec{replies: map[string]string{
		"What do I have today?": "Guitar at 6.",
		"What's due today?":     "Ship notifications.",
	}}
	notices := &stubNoticePoster{}
	sched := NewAutomationScheduler(repo, chatExec, notices, nil, nil)
	sched.now = func() time.Time { return now }

	sched.Tick(context.Background())
	if chatExec.calls != 2 {
		t.Fatalf("chat calls = %d", chatExec.calls)
	}
	if notices.posts != 1 {
		t.Fatalf("posts = %d", notices.posts)
	}
	if !strings.Contains(notices.lastContent, "Guitar at 6.") {
		t.Fatalf("combined missing first reply: %q", notices.lastContent)
	}
	if !strings.Contains(notices.lastContent, "Ship notifications.") {
		t.Fatalf("combined missing second reply: %q", notices.lastContent)
	}
	if notices.lastClientID != "automation:atm_test:2026-08-01" {
		t.Fatalf("client id = %s", notices.lastClientID)
	}
	auto := repo.byID[autoID]
	if auto.LastRunAt == nil {
		t.Fatal("expected last_run_at")
	}
	if auto.NextRunAt == nil {
		t.Fatal("expected next_run_at")
	}

	sched.Tick(context.Background())
	if notices.posts != 1 {
		t.Fatalf("second tick should not post again, posts = %d", notices.posts)
	}
	if chatExec.calls != 2 {
		t.Fatalf("second tick should not re-execute, calls = %d", chatExec.calls)
	}
}

func TestAutomationSchedulerSkipsDisabled(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000404")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000405")
	now := time.Date(2026, 8, 1, 7, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	repo.byID[autoID] = entity.Automation{
		ID: autoID, PublicID: "atm_off", UserID: userID, Name: "Off",
		Enabled: false, TriggerType: constant.AutomationTriggerDaily,
		TriggerTime: "07:00", Timezone: "UTC",
		Commands: []entity.AutomationCommand{{Command: constant.AutomationCommandGreeting}}, DeliveryChannels: []string{"chat"},
	}
	chatExec := &stubChatExec{replies: map[string]string{"Hi": "Hello"}}
	notices := &stubNoticePoster{}
	sched := NewAutomationScheduler(repo, chatExec, notices, nil, nil)
	sched.now = func() time.Time { return now }
	sched.Tick(context.Background())
	if notices.posts != 0 || chatExec.calls != 0 {
		t.Fatal("disabled automation must not run")
	}
}

func TestAutomationSchedulerRecordsExecutionHistory(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000406")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000407")
	now := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	repo.byID[autoID] = entity.Automation{
		ID: autoID, PublicID: "atm_hist", UserID: userID, Name: "Brief",
		Enabled: true, TriggerType: constant.AutomationTriggerDaily,
		TriggerTime: "08:00", Timezone: "UTC",
		Commands: []entity.AutomationCommand{
			{Command: constant.AutomationCommandTodaysAgenda, Variables: map[string]string{"range": "today"}},
			{Command: constant.AutomationCommandTasksDue, Variables: map[string]string{"priority": "all"}},
		},
		DeliveryChannels: []string{"chat"},
	}
	chatExec := &stubChatExec{
		replies: map[string]string{
			"What do I have today?": "Agenda.",
			"What's due today?":     "Tasks.",
		},
		intents: map[string]string{
			"What do I have today?": "QUERY_TODAY",
			"What's due today?":     "QUERY_DUE_TODAY",
		},
	}
	notices := &stubNoticePoster{}
	recorder := newMemExecutionRecorder()
	sched := NewAutomationScheduler(repo, chatExec, notices, recorder, nil)
	sched.now = func() time.Time { return now }
	sched.Tick(context.Background())

	if len(recorder.executions) != 1 {
		t.Fatalf("executions = %d", len(recorder.executions))
	}
	var exec entity.AutomationExecution
	for _, e := range recorder.executions {
		exec = e
	}
	if exec.Status != constant.AutomationExecutionSuccess {
		t.Fatalf("status = %s", exec.Status)
	}
	if exec.CommandsTotal != 2 || exec.CommandsSuccess != 2 {
		t.Fatalf("commands = %d/%d", exec.CommandsSuccess, exec.CommandsTotal)
	}
	if exec.DeliveryStatus == nil || *exec.DeliveryStatus != constant.AutomationDeliverySent {
		t.Fatalf("delivery = %v", exec.DeliveryStatus)
	}
	cmds := recorder.commands[exec.ID]
	if len(cmds) != 2 {
		t.Fatalf("command rows = %d", len(cmds))
	}
}

func TestAutomationSchedulerPartialFailure(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000408")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000409")
	now := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	repo.byID[autoID] = entity.Automation{
		ID: autoID, PublicID: "atm_partial", UserID: userID, Name: "Mixed",
		Enabled: true, TriggerType: constant.AutomationTriggerDaily,
		TriggerTime: "08:00", Timezone: "UTC",
		Commands: []entity.AutomationCommand{
			{Command: constant.AutomationCommandChatMessage, Variables: map[string]string{"message": "ok cmd"}},
			{Command: constant.AutomationCommandChatMessage, Variables: map[string]string{"message": "bad cmd"}},
		},
		DeliveryChannels: []string{"chat"},
	}
	chatExec := &stubChatExec{
		replies: map[string]string{"ok cmd": "All good.", "bad cmd": "I couldn't do that."},
		errors:  map[string]string{"bad cmd": "timeout"},
	}
	notices := &stubNoticePoster{}
	recorder := newMemExecutionRecorder()
	sched := NewAutomationScheduler(repo, chatExec, notices, recorder, nil)
	sched.now = func() time.Time { return now }
	sched.Tick(context.Background())

	var exec entity.AutomationExecution
	for _, e := range recorder.executions {
		exec = e
	}
	if exec.Status != constant.AutomationExecutionPartialSuccess {
		t.Fatalf("status = %s", exec.Status)
	}
	if exec.CommandsSuccess != 1 || exec.CommandsFailed != 1 {
		t.Fatalf("success/fail = %d/%d", exec.CommandsSuccess, exec.CommandsFailed)
	}
}

func TestDeriveExecutionStatus(t *testing.T) {
	t.Parallel()
	if got := DeriveExecutionStatus(2, 0, 0); got != constant.AutomationExecutionSuccess {
		t.Fatalf("got %s", got)
	}
	if got := DeriveExecutionStatus(1, 1, 0); got != constant.AutomationExecutionPartialSuccess {
		t.Fatalf("got %s", got)
	}
	if got := DeriveExecutionStatus(0, 2, 0); got != constant.AutomationExecutionFailed {
		t.Fatalf("got %s", got)
	}
}

func TestResolveAutomationCommandVariables(t *testing.T) {
	t.Parallel()
	msg, label, err := ResolveAutomationCommand(entity.AutomationCommand{
		Command:   constant.AutomationCommandTodaysAgenda,
		Variables: map[string]string{"range": "tomorrow"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "What do I have tomorrow?" || label != "Tomorrow's Agenda" {
		t.Fatalf("got %q / %q", msg, label)
	}
}

func TestAutomationRunnerManualRecordsHistoryWithoutSchedule(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000410")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000411")
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	repo.byID[autoID] = entity.Automation{
		ID: autoID, PublicID: "atm_manual", UserID: userID, Name: "Brief",
		Enabled: true, TriggerType: constant.AutomationTriggerDaily,
		TriggerTime: "09:00", Timezone: "UTC",
		Commands: []entity.AutomationCommand{
			{Command: constant.AutomationCommandGreeting},
		},
		DeliveryChannels: []string{"chat"},
	}
	chatExec := &stubChatExec{replies: map[string]string{"Hi": "Hello there."}}
	notices := &stubNoticePoster{}
	recorder := newMemExecutionRecorder()
	runner := NewAutomationRunner(repo, chatExec, notices, recorder, nil)
	runner.now = func() time.Time { return now }

	result, err := runner.Run(context.Background(), repo.byID[autoID], AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourceManual,
		RecordHistory:  true,
		DeliverToChat:  true,
		UpdateSchedule: false,
		Now:            now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TriggerSource != constant.AutomationTriggerSourceManual {
		t.Fatalf("trigger = %s", result.TriggerSource)
	}
	if result.Execution == nil {
		t.Fatal("expected execution history")
	}
	if notices.posts != 1 {
		t.Fatalf("posts = %d", notices.posts)
	}
	if !strings.HasPrefix(notices.lastClientID, "automation:manual:") {
		t.Fatalf("client id = %s", notices.lastClientID)
	}
	auto := repo.byID[autoID]
	if auto.LastRunAt != nil {
		t.Fatal("manual run must not update last_run_at")
	}
	var exec entity.AutomationExecution
	for _, e := range recorder.executions {
		exec = e
	}
	if exec.TriggerSource != constant.AutomationTriggerSourceManual {
		t.Fatalf("stored trigger = %s", exec.TriggerSource)
	}
}

func TestAutomationRunnerPreviewSkipsHistoryAndDelivery(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000412")
	autoID := uuid.MustParse("018f0000-0000-7000-8000-000000000413")
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	repo := newMemAutomationRepo()
	auto := entity.Automation{
		ID: autoID, PublicID: "atm_preview", UserID: userID, Name: "Brief",
		Enabled: true, TriggerType: constant.AutomationTriggerDaily,
		TriggerTime: "09:00", Timezone: "UTC",
		Commands: []entity.AutomationCommand{
			{Command: constant.AutomationCommandTasksDue, Variables: map[string]string{"priority": "all"}},
		},
		DeliveryChannels: []string{"chat"},
	}
	repo.byID[autoID] = auto
	chatExec := &stubChatExec{replies: map[string]string{"What's due today?": "Nothing due today."}}
	notices := &stubNoticePoster{}
	recorder := newMemExecutionRecorder()
	runner := NewAutomationRunner(repo, chatExec, notices, recorder, nil)
	runner.now = func() time.Time { return now }

	result, err := runner.Run(context.Background(), auto, AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourcePreview,
		RecordHistory:  false,
		DeliverToChat:  false,
		UpdateSchedule: false,
		DryRun:         true,
		Now:            now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !chatExec.lastDryRun {
		t.Fatal("expected dry-run chat execution")
	}
	if notices.posts != 0 {
		t.Fatalf("preview must not deliver, posts = %d", notices.posts)
	}
	if len(recorder.executions) != 0 {
		t.Fatalf("preview must not record history, got %d", len(recorder.executions))
	}
	if result.Execution != nil {
		t.Fatal("preview must not return execution")
	}
	if result.DeliveryStatus != constant.AutomationDeliverySkipped {
		t.Fatalf("delivery = %s", result.DeliveryStatus)
	}
	if !strings.Contains(result.Response, "Nothing due today.") {
		t.Fatalf("response = %q", result.Response)
	}
}

func strPtr(s string) *string { return &s }

type memAutomationRepo struct {
	byID map[uuid.UUID]entity.Automation
}

func newMemAutomationRepo() *memAutomationRepo {
	return &memAutomationRepo{byID: map[uuid.UUID]entity.Automation{}}
}

func (m *memAutomationRepo) Create(_ context.Context, auto entity.Automation) (entity.Automation, error) {
	m.byID[auto.ID] = auto
	return auto, nil
}

func (m *memAutomationRepo) GetByID(_ context.Context, id uuid.UUID) (entity.Automation, error) {
	auto, ok := m.byID[id]
	if !ok || auto.DeletedAt != nil {
		return entity.Automation{}, apperr.ErrNotFound
	}
	return auto, nil
}

func (m *memAutomationRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]entity.Automation, error) {
	out := make([]entity.Automation, 0)
	for _, auto := range m.byID {
		if auto.UserID == userID && auto.DeletedAt == nil {
			out = append(out, auto)
		}
	}
	return out, nil
}

func (m *memAutomationRepo) ListEnabled(context.Context) ([]entity.Automation, error) {
	out := make([]entity.Automation, 0)
	for _, auto := range m.byID {
		if auto.DeletedAt == nil && auto.Enabled {
			out = append(out, auto)
		}
	}
	return out, nil
}

func (m *memAutomationRepo) Update(
	_ context.Context,
	id, userID uuid.UUID,
	fields repository.AutomationUpdateFields,
	updatedAt time.Time,
) (entity.Automation, error) {
	auto, ok := m.byID[id]
	if !ok || auto.DeletedAt != nil || auto.UserID != userID {
		return entity.Automation{}, apperr.ErrNotFound
	}
	if fields.Name != nil {
		auto.Name = *fields.Name
	}
	if fields.Description != nil {
		auto.Description = fields.Description
	}
	if fields.Enabled != nil {
		auto.Enabled = *fields.Enabled
	}
	if fields.TriggerType != nil {
		auto.TriggerType = *fields.TriggerType
	}
	if fields.TriggerTime != nil {
		auto.TriggerTime = *fields.TriggerTime
	}
	if fields.TriggerDaysSet {
		auto.TriggerDays = fields.TriggerDays
		if auto.TriggerDays == nil {
			auto.TriggerDays = []string{}
		}
	}
	if fields.Timezone != nil {
		auto.Timezone = *fields.Timezone
	}
	if fields.Commands != nil {
		auto.Commands = fields.Commands
	}
	if fields.DeliveryChannels != nil {
		auto.DeliveryChannels = fields.DeliveryChannels
	}
	if fields.NextRunAt != nil {
		auto.NextRunAt = fields.NextRunAt
	}
	auto.UpdatedAt = updatedAt
	m.byID[id] = auto
	return auto, nil
}

func (m *memAutomationRepo) MarkRun(_ context.Context, id uuid.UUID, ranAt time.Time, nextRunAt *time.Time) (entity.Automation, error) {
	auto, ok := m.byID[id]
	if !ok || auto.DeletedAt != nil {
		return entity.Automation{}, apperr.ErrNotFound
	}
	auto.LastRunAt = &ranAt
	auto.NextRunAt = nextRunAt
	auto.UpdatedAt = ranAt
	m.byID[id] = auto
	return auto, nil
}

func (m *memAutomationRepo) SoftDelete(_ context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	auto, ok := m.byID[id]
	if !ok || auto.DeletedAt != nil || auto.UserID != userID {
		return apperr.ErrNotFound
	}
	auto.DeletedAt = &deletedAt
	auto.UpdatedAt = deletedAt
	m.byID[id] = auto
	return nil
}

func (m *memAutomationRepo) WithTx(pgx.Tx) repository.AutomationRepository { return m }

type stubChatExec struct {
	replies    map[string]string
	intents    map[string]string
	errors     map[string]string
	calls      int
	lastDryRun bool
}

func (s *stubChatExec) Execute(_ context.Context, in ChatCommandInput) ChatCommandResult {
	s.calls++
	s.lastDryRun = in.DryRun
	out := ChatCommandResult{Reply: "ok"}
	if s.replies != nil {
		if reply, ok := s.replies[in.Message]; ok {
			out.Reply = reply
		}
	}
	if s.intents != nil {
		out.Intent = s.intents[in.Message]
	}
	if s.errors != nil {
		out.Error = s.errors[in.Message]
	}
	return out
}

type stubNoticePoster struct {
	posts        int
	lastClientID string
	lastContent  string
}

func (s *stubNoticePoster) PostAssistantNotice(
	_ context.Context,
	_ uuid.UUID,
	content string,
	clientMessageID string,
) (entity.Message, bool, error) {
	s.posts++
	s.lastClientID = clientMessageID
	s.lastContent = content
	return entity.Message{}, true, nil
}

type memExecutionRecorder struct {
	executions map[uuid.UUID]entity.AutomationExecution
	commands   map[uuid.UUID][]entity.AutomationCommandExecution
}

func newMemExecutionRecorder() *memExecutionRecorder {
	return &memExecutionRecorder{
		executions: map[uuid.UUID]entity.AutomationExecution{},
		commands:   map[uuid.UUID][]entity.AutomationCommandExecution{},
	}
}

func (m *memExecutionRecorder) BeginExecution(
	_ context.Context,
	auto entity.Automation,
	triggerSource string,
) (entity.AutomationExecution, error) {
	id := uuid.New()
	now := time.Now().UTC()
	pending := constant.AutomationDeliveryPending
	exec := entity.AutomationExecution{
		ID:               id,
		PublicID:         "aex_test",
		AutomationID:     auto.ID,
		UserID:           auto.UserID,
		StartedAt:        now,
		Status:           constant.AutomationExecutionRunning,
		TriggerSource:    triggerSource,
		DeliveryChannels: auto.DeliveryChannels,
		DeliveryStatus:   &pending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.executions[id] = exec
	return exec, nil
}

func (m *memExecutionRecorder) RecordCommand(
	_ context.Context,
	in RecordCommandInput,
) (entity.AutomationCommandExecution, error) {
	cmd := entity.AutomationCommandExecution{
		ID:          uuid.New(),
		ExecutionID: in.ExecutionID,
		OrderIndex:  in.OrderIndex,
		Command:     in.Command,
		Status:      in.Status,
		Response:    strPtr(in.Response),
		Error:       strPtr(in.Error),
	}
	if in.CommandType != "" {
		cmd.CommandType = &in.CommandType
	}
	m.commands[in.ExecutionID] = append(m.commands[in.ExecutionID], cmd)
	return cmd, nil
}

func (m *memExecutionRecorder) CompleteExecution(
	_ context.Context,
	in CompleteExecutionInput,
) (entity.AutomationExecution, error) {
	exec, ok := m.executions[in.ExecutionID]
	if !ok {
		return entity.AutomationExecution{}, apperr.ErrNotFound
	}
	now := time.Now().UTC()
	exec.CompletedAt = &now
	exec.Status = in.Status
	exec.CommandsTotal = in.CommandsTotal
	exec.CommandsSuccess = in.CommandsSuccess
	exec.CommandsFailed = in.CommandsFailed
	ds := in.DeliveryStatus
	exec.DeliveryStatus = &ds
	if in.Response != "" {
		exec.Response = &in.Response
	}
	if in.Error != "" {
		exec.Error = &in.Error
	}
	dur := int(now.Sub(in.StartedAt).Milliseconds())
	exec.DurationMs = &dur
	exec.UpdatedAt = now
	m.executions[in.ExecutionID] = exec
	return exec, nil
}

func TestAutomationDueWeeklyWeekday(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	// Saturday 2026-08-01 09:00 IST
	now := time.Date(2026, 8, 1, 9, 0, 0, 0, loc)
	auto := entity.Automation{
		Enabled:     true,
		TriggerType: constant.AutomationTriggerWeekly,
		Timezone:    "Asia/Kolkata",
		TriggerTime: "09:00",
		TriggerDays: []string{"MO", "WE", "FR"},
		Commands:    []entity.AutomationCommand{{Command: constant.AutomationCommandGreeting}},
	}
	if AutomationDue(auto, now) {
		t.Fatal("Saturday should not match MO/WE/FR")
	}
	auto.TriggerDays = []string{"SA"}
	if !AutomationDue(auto, now) {
		t.Fatal("expected due on Saturday")
	}
}

func TestNextAutomationRunAtWeeklySkipsNonMatchingDays(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatal(err)
	}
	// Friday evening after the trigger — next should be Monday 09:00
	now := time.Date(2026, 7, 31, 10, 0, 0, 0, loc)
	next := NextAutomationRunAt(now, "UTC", constant.AutomationTriggerWeekly, "09:00", []string{"MO", "WE"})
	got := next.In(loc)
	if got.Weekday() != time.Monday || got.Format("15:04") != "09:00" {
		t.Fatalf("got %v, want Monday 09:00", got)
	}
}
