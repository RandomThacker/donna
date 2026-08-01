package business

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/RandomThacker/donna/services/api/internal/webpush"
	"github.com/google/uuid"
)

// AutomationRunOptions controls history, delivery, schedule markers, and dry-run.
type AutomationRunOptions struct {
	TriggerSource string
	RecordHistory bool
	// DeliverToChat, when true, delivers the combined reply to configured
	// channels (chat and/or web push). Preview sets this false.
	DeliverToChat  bool
	UpdateSchedule bool
	DryRun         bool
	Now            time.Time
	DisplayName    string
}

// AutomationCommandRunResult is one command outcome from a run/preview.
type AutomationCommandRunResult struct {
	OrderIndex  int
	Command     string
	CommandKey  string
	CommandType string
	Status      string
	DurationMs  int
	Response    string
	Error       string
}

// AutomationRunResult is the outcome of Manual / Scheduled / Preview execution.
type AutomationRunResult struct {
	Response        string
	Status          string
	DeliveryStatus  string
	CommandsTotal   int
	CommandsSuccess int
	CommandsFailed  int
	DurationMs      int
	TriggerSource   string
	Commands        []AutomationCommandRunResult
	Execution       *entity.AutomationExecution
}

// AutomationRunner executes automation commands through the existing chat path.
// Scheduling (due checks / tick) stays in AutomationScheduler.
type AutomationRunner struct {
	autos       AutomationLister
	chat        ChatCommandExecutor
	notices     AssistantNoticePoster
	executions  AutomationExecutionRecorder
	personality personality.Renderer
	pushSubs    *PushSubscriptionService
	pushSender  webpush.Sender
	log         *logger.Logger
	now         func() time.Time
}

// NewAutomationRunner constructs the shared runner used by scheduler, manual run, and preview.
func NewAutomationRunner(
	autos AutomationLister,
	chatExec ChatCommandExecutor,
	notices AssistantNoticePoster,
	executions AutomationExecutionRecorder,
	log *logger.Logger,
) *AutomationRunner {
	return &AutomationRunner{
		autos:      autos,
		chat:       chatExec,
		notices:    notices,
		executions: executions,
		log:        log,
		now:        time.Now,
	}
}

// SetPersonality attaches an optional personality renderer for combined replies.
func (r *AutomationRunner) SetPersonality(renderer personality.Renderer) {
	if r != nil {
		r.personality = renderer
	}
}

// SetWebPush attaches optional Web Push delivery (VAPID). Nil/unconfigured is fine.
func (r *AutomationRunner) SetWebPush(subs *PushSubscriptionService, sender webpush.Sender) {
	if r == nil {
		return
	}
	r.pushSubs = subs
	r.pushSender = sender
}

func (r *AutomationRunner) webPushConfigured() bool {
	return r != nil && r.pushSubs != nil && r.pushSender != nil && r.pushSender.Configured()
}

// Run executes an automation according to opts. Does not change due/idempotency logic.
func (r *AutomationRunner) Run(ctx context.Context, auto entity.Automation, opts AutomationRunOptions) (AutomationRunResult, error) {
	if r == nil || r.chat == nil {
		return AutomationRunResult{}, fmt.Errorf("automation runner is not configured")
	}
	if len(auto.Commands) == 0 {
		return AutomationRunResult{}, fmt.Errorf("automation has no commands")
	}

	now := opts.Now
	if now.IsZero() {
		now = r.now().UTC()
	} else {
		now = now.UTC()
	}
	startedWall := r.now().UTC()

	trigger := strings.TrimSpace(opts.TriggerSource)
	if trigger == "" {
		trigger = constant.AutomationTriggerSourceScheduler
	}

	var exec entity.AutomationExecution
	record := opts.RecordHistory && r.executions != nil
	if record {
		started, beginErr := r.executions.BeginExecution(ctx, auto, trigger)
		if beginErr != nil {
			if r.log != nil {
				r.log.Warn(ctx, "automation execution begin failed",
					"automation_id", auto.ID.String(),
					constant.LogAttrError, beginErr,
				)
			}
			record = false
		} else {
			exec = started
		}
	}

	parts := make([]string, 0, len(auto.Commands))
	cmdResults := make([]AutomationCommandRunResult, 0, len(auto.Commands))
	success, failed, skipped := 0, 0, 0
	greetingOnly := false
	order := 0

	for _, structured := range auto.Commands {
		message, label, resolveErr := ResolveAutomationCommand(structured)
		cmdStart := r.now().UTC()
		var result ChatCommandResult
		status := constant.AutomationCommandSuccess
		errText := ""
		reply := ""
		isGreeting := strings.EqualFold(strings.TrimSpace(structured.Command), constant.AutomationCommandGreeting)

		if resolveErr != nil {
			status = constant.AutomationCommandFailed
			errText = resolveErr.Error()
			failed++
		} else {
			result = r.chat.Execute(ctx, ChatCommandInput{
				UserID:          auto.UserID,
				Timezone:        auto.Timezone,
				Now:             now,
				Message:         message,
				DisplayName:     opts.DisplayName,
				DryRun:          opts.DryRun,
				SkipPersonality: true,
			})
			reply = strings.TrimSpace(result.Reply)
			errText = strings.TrimSpace(result.Error)
			if errText != "" {
				status = constant.AutomationCommandFailed
				failed++
			} else if reply == "" && isGreeting {
				// Greeting body comes from the personality renderer on the combined reply.
				success++
				greetingOnly = greetingOnly || len(parts) == 0
			} else if reply == "" {
				status = constant.AutomationCommandSkipped
				skipped++
			} else {
				success++
				parts = append(parts, reply)
				greetingOnly = false
			}
		}
		cmdEnd := r.now().UTC()
		dur := int(cmdEnd.Sub(cmdStart).Milliseconds())
		if dur < 0 {
			dur = 0
		}

		displayCmd := label
		if displayCmd == "" {
			displayCmd = structured.Command
		}
		cmdResults = append(cmdResults, AutomationCommandRunResult{
			OrderIndex:  order,
			Command:     displayCmd,
			CommandKey:  structured.Command,
			CommandType: result.Intent,
			Status:      status,
			DurationMs:  dur,
			Response:    reply,
			Error:       errText,
		})

		if record && exec.ID != uuid.Nil {
			_, _ = r.executions.RecordCommand(ctx, RecordCommandInput{
				ExecutionID: exec.ID,
				OrderIndex:  order,
				Command:     displayCmd,
				CommandType: result.Intent,
				StartedAt:   cmdStart,
				CompletedAt: cmdEnd,
				Status:      status,
				Response:    reply,
				Error:       errText,
			})
		}
		order++
	}

	combined := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if r.personality != nil {
		kind := personality.KindAutomation
		switch {
		case greetingOnly && combined == "":
			kind = personality.KindGreeting
		case strings.Contains(strings.ToLower(auto.Name), "morning"):
			kind = personality.KindMorningBrief
		}
		if combined == "" && kind != personality.KindGreeting {
			combined = "I'm here — nothing to report just yet."
		}
		if rendered, err := r.personality.Render(ctx, personality.RenderInput{
			UserID:    auto.UserID,
			Canonical: combined,
			Kind:      kind,
			Now:       now,
			Timezone:  auto.Timezone,
		}); err == nil && strings.TrimSpace(rendered.Text) != "" {
			combined = strings.TrimSpace(rendered.Text)
		}
	} else if combined == "" {
		combined = "I'm here — nothing to report just yet."
	}

	deliveryStatus := constant.AutomationDeliverySkipped
	execStatus := DeriveExecutionStatus(success, failed, skipped)
	var deliveryErrText string
	var deliveryErr error

	if opts.DeliverToChat {
		chatWanted := deliversChat(auto.DeliveryChannels)
		pushWanted := deliversPush(auto.DeliveryChannels)
		anySent := false

		if chatWanted {
			if r.notices == nil {
				deliveryErr = fmt.Errorf("assistant notice poster is not configured")
				deliveryStatus = constant.AutomationDeliveryFailed
				execStatus = constant.AutomationExecutionFailed
				deliveryErrText = deliveryErr.Error()
			} else {
				loc, err := time.LoadLocation(strings.TrimSpace(auto.Timezone))
				if err != nil || loc == nil {
					loc = time.UTC
				}
				localNow := now.In(loc)
				clientID := ClientMessageIDForAutomationRun(auto.PublicID, localNow)
				if trigger == constant.AutomationTriggerSourceManual && exec.PublicID != "" {
					clientID = ClientMessageIDForManualRun(exec.PublicID)
				}
				_, _, deliveryErr = r.notices.PostAssistantNotice(ctx, auto.UserID, combined, clientID)
				if deliveryErr != nil {
					deliveryStatus = constant.AutomationDeliveryFailed
					execStatus = constant.AutomationExecutionFailed
					deliveryErrText = deliveryErr.Error()
				} else {
					anySent = true
					deliveryStatus = constant.AutomationDeliverySent
				}
			}
		}

		if pushWanted && deliveryErr == nil {
			if r.deliverWebPush(ctx, auto, combined) {
				anySent = true
				deliveryStatus = constant.AutomationDeliverySent
			} else if !anySent {
				deliveryStatus = constant.AutomationDeliveryFailed
				deliveryErrText = "web push delivery failed or no device subscribed"
			}
		}

		if !chatWanted && !pushWanted {
			deliveryStatus = constant.AutomationDeliverySkipped
		}
	}

	total := success + failed + skipped
	wallEnd := r.now().UTC()
	durationMs := int(wallEnd.Sub(startedWall).Milliseconds())
	if durationMs < 0 {
		durationMs = 0
	}

	out := AutomationRunResult{
		Response:        combined,
		Status:          execStatus,
		DeliveryStatus:  deliveryStatus,
		CommandsTotal:   total,
		CommandsSuccess: success,
		CommandsFailed:  failed,
		DurationMs:      durationMs,
		TriggerSource:   trigger,
		Commands:        cmdResults,
	}

	if record && exec.ID != uuid.Nil {
		completed, completeErr := r.executions.CompleteExecution(ctx, CompleteExecutionInput{
			ExecutionID:     exec.ID,
			Status:          execStatus,
			CommandsTotal:   total,
			CommandsSuccess: success,
			CommandsFailed:  failed,
			DeliveryStatus:  deliveryStatus,
			Response:        combined,
			Error:           deliveryErrText,
			StartedAt:       exec.StartedAt,
		})
		if completeErr == nil {
			out.Execution = &completed
		} else {
			out.Execution = &exec
		}
	}

	if deliveryErr != nil {
		return out, deliveryErr
	}

	if opts.UpdateSchedule && r.autos != nil {
		next := NextAutomationRunAt(now.Add(time.Minute), auto.Timezone, auto.TriggerType, auto.TriggerTime, auto.TriggerDays)
		if _, err := r.autos.MarkRun(ctx, auto.ID, now, &next); err != nil {
			return out, err
		}
	}
	return out, nil
}

// ClientMessageIDForManualRun builds a unique delivery id per manual execution.
func ClientMessageIDForManualRun(executionPublicID string) string {
	return fmt.Sprintf("automation:manual:%s", strings.TrimSpace(executionPublicID))
}

func (r *AutomationRunner) deliverWebPush(ctx context.Context, auto entity.Automation, body string) bool {
	if !r.webPushConfigured() {
		if r.log != nil {
			r.log.Info(ctx, "automation web push skipped — not configured",
				"automation_id", auto.ID.String(),
			)
		}
		return false
	}
	subs, err := r.pushSubs.List(ctx, auto.UserID)
	if err != nil {
		if r.log != nil {
			r.log.Warn(ctx, "automation web push list failed",
				"automation_id", auto.ID.String(),
				constant.LogAttrError, err,
			)
		}
		return false
	}
	if len(subs) == 0 {
		if r.log != nil {
			r.log.Info(ctx, "automation web push skipped — no device subscriptions",
				"automation_id", auto.ID.String(),
				"user_id", auto.UserID.String(),
			)
		}
		return false
	}

	title := strings.TrimSpace(auto.Name)
	if title == "" {
		title = "Donna"
	}
	payload := webpush.Payload{
		Title:    title,
		Body:     truncatePushBody(body),
		DeepLink: constant.NotificationDeepLinkPath,
	}

	var delivered int
	for _, sub := range subs {
		result, sendErr := r.pushSender.Send(ctx, sub, payload)
		if sendErr != nil {
			if r.log != nil {
				r.log.Warn(ctx, "automation web push send failed",
					"automation_id", auto.ID.String(),
					"endpoint", sub.Endpoint,
					"status_code", result.StatusCode,
					constant.LogAttrError, sendErr,
				)
			}
			if result.Gone {
				_ = r.pushSubs.Unsubscribe(ctx, auto.UserID, sub.Endpoint)
			}
			continue
		}
		delivered++
	}
	if r.log != nil {
		r.log.Info(ctx, "automation web push fan-out",
			"automation_id", auto.ID.String(),
			"subscriptions", len(subs),
			"delivered", delivered,
		)
	}
	return delivered > 0
}

const automationPushBodyMaxRunes = 240

func truncatePushBody(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "Donna has an update for you."
	}
	if utf8.RuneCountInString(trimmed) <= automationPushBodyMaxRunes {
		return trimmed
	}
	runes := []rune(trimmed)
	return strings.TrimSpace(string(runes[:automationPushBodyMaxRunes-1])) + "…"
}
