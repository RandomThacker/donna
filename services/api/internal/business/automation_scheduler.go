package business

import (
	"context"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/google/uuid"
)

// AutomationLister lists enabled automations for the scheduler.
type AutomationLister interface {
	ListEnabled(ctx context.Context) ([]entity.Automation, error)
	MarkRun(ctx context.Context, autoID uuid.UUID, ranAt time.Time, nextRunAt *time.Time) (entity.Automation, error)
}

// ChatCommandInput is the runner's view of a chat execute call (avoids importing chat).
type ChatCommandInput struct {
	UserID          uuid.UUID
	Timezone        string
	Now             time.Time
	Message         string
	DisplayName     string
	DryRun          bool // preview: no mutation intents
	SkipPersonality bool // automation: personalize the combined reply instead
}

// ChatCommandResult is the runner's view of a chat execute result.
type ChatCommandResult struct {
	Reply  string
	Intent string
	Error  string // non-empty when the command failed at the action layer
}

// ChatCommandExecutor runs a natural-language command through the chat executor.
type ChatCommandExecutor interface {
	Execute(ctx context.Context, in ChatCommandInput) ChatCommandResult
}

// AssistantNoticePoster posts Donna messages into the primary chat thread.
type AssistantNoticePoster interface {
	PostAssistantNotice(ctx context.Context, userID uuid.UUID, content string, clientMessageID string) (entity.Message, bool, error)
}

// AutomationExecutionRecorder records run history without changing command execution.
type AutomationExecutionRecorder interface {
	BeginExecution(ctx context.Context, auto entity.Automation, triggerSource string) (entity.AutomationExecution, error)
	RecordCommand(ctx context.Context, in RecordCommandInput) (entity.AutomationCommandExecution, error)
	CompleteExecution(ctx context.Context, in CompleteExecutionInput) (entity.AutomationExecution, error)
}

// AutomationScheduler fires due automations once per local civil day.
// It only knows Trigger → Runner; execution lives in AutomationRunner.
type AutomationScheduler struct {
	autos    AutomationLister
	runner   *AutomationRunner
	log      *logger.Logger
	interval time.Duration
	now      func() time.Time
}

// NewAutomationScheduler constructs a minute ticker for automations.
func NewAutomationScheduler(
	autos AutomationLister,
	chatExec ChatCommandExecutor,
	notices AssistantNoticePoster,
	executions AutomationExecutionRecorder,
	log *logger.Logger,
) *AutomationScheduler {
	return &AutomationScheduler{
		autos:    autos,
		runner:   NewAutomationRunner(autos, chatExec, notices, executions, log),
		log:      log,
		interval: constant.AutomationSchedulerInterval,
		now:      time.Now,
	}
}

// Runner exposes the shared runner for manual run / preview actions.
func (s *AutomationScheduler) Runner() *AutomationRunner {
	if s == nil {
		return nil
	}
	return s.runner
}

// Run blocks until ctx is canceled, ticking every minute.
func (s *AutomationScheduler) Run(ctx context.Context) {
	if s.autos == nil || s.runner == nil || s.runner.chat == nil || s.runner.notices == nil {
		return
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.Tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Tick(ctx)
		}
	}
}

// Tick scans enabled automations once (exported for tests).
func (s *AutomationScheduler) Tick(ctx context.Context) {
	if s.autos == nil || s.runner == nil || s.runner.chat == nil || s.runner.notices == nil {
		return
	}
	now := s.now().UTC()
	s.runner.now = s.now
	autos, err := s.autos.ListEnabled(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "automation scheduler list failed", constant.LogAttrError, err)
		}
		return
	}
	for _, auto := range autos {
		if !AutomationDue(auto, now) {
			continue
		}
		if err := s.fire(ctx, auto, now); err != nil {
			if s.log != nil {
				s.log.Warn(ctx, "automation run failed",
					"automation_id", auto.ID.String(),
					"user_id", auto.UserID.String(),
					constant.LogAttrError, err,
				)
			}
		}
	}
}

func (s *AutomationScheduler) fire(ctx context.Context, auto entity.Automation, now time.Time) error {
	if !deliversChat(auto.DeliveryChannels) {
		if s.log != nil {
			s.log.Warn(ctx, "automation skipped — no chat delivery channel",
				"automation_id", auto.ID.String(),
			)
		}
		return nil
	}
	if len(auto.Commands) == 0 {
		if s.log != nil {
			s.log.Warn(ctx, "automation skipped — no commands",
				"automation_id", auto.ID.String(),
			)
		}
		return nil
	}

	_, err := s.runner.Run(ctx, auto, AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourceScheduler,
		RecordHistory:  true,
		DeliverToChat:  true,
		UpdateSchedule: true,
		DryRun:         false,
		Now:            now,
	})
	return err
}

func deliversChat(channels []string) bool {
	for _, ch := range channels {
		if strings.EqualFold(strings.TrimSpace(ch), constant.AutomationDeliveryChat) {
			return true
		}
	}
	// Empty channels default to chat for migrated/legacy rows.
	return len(channels) == 0
}
