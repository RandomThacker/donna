package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
)

// AutomationExecutionService records and queries automation run history.
type AutomationExecutionService struct {
	execs repository.AutomationExecutionRepository
	autos repository.AutomationRepository
	now   func() time.Time
}

// NewAutomationExecutionService constructs an AutomationExecutionService.
func NewAutomationExecutionService(
	execs repository.AutomationExecutionRepository,
	autos repository.AutomationRepository,
) *AutomationExecutionService {
	return &AutomationExecutionService{execs: execs, autos: autos, now: time.Now}
}

// BeginExecution creates a RUNNING execution row before commands run.
func (s *AutomationExecutionService) BeginExecution(
	ctx context.Context,
	auto entity.Automation,
	triggerSource string,
) (entity.AutomationExecution, error) {
	if auto.ID == uuid.Nil || auto.UserID == uuid.Nil {
		return entity.AutomationExecution{}, fmt.Errorf("%w: automation and user id are required", apperr.ErrValidation)
	}
	src := strings.TrimSpace(triggerSource)
	if src == "" {
		src = constant.AutomationTriggerSourceScheduler
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.AutomationExecution{}, err
	}
	now := s.now().UTC()
	pending := constant.AutomationDeliveryPending
	channels := append([]string{}, auto.DeliveryChannels...)
	if len(channels) == 0 {
		channels = []string{constant.AutomationDeliveryChat}
	}
	return s.execs.CreateExecution(ctx, entity.AutomationExecution{
		ID:               id,
		PublicID:         idgen.PublicID(constant.PublicIDPrefixAutomationExecution, id),
		AutomationID:     auto.ID,
		UserID:           auto.UserID,
		StartedAt:        now,
		Status:           constant.AutomationExecutionRunning,
		TriggerSource:    src,
		DeliveryChannels: channels,
		DeliveryStatus:   &pending,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

// RecordCommandInput captures one command result within an execution.
type RecordCommandInput struct {
	ExecutionID uuid.UUID
	OrderIndex  int
	Command     string
	CommandType string
	StartedAt   time.Time
	CompletedAt time.Time
	Status      string
	Response    string
	Error       string
}

// RecordCommand stores a per-command execution result.
func (s *AutomationExecutionService) RecordCommand(ctx context.Context, in RecordCommandInput) (entity.AutomationCommandExecution, error) {
	if in.ExecutionID == uuid.Nil {
		return entity.AutomationCommandExecution{}, fmt.Errorf("%w: execution id is required", apperr.ErrValidation)
	}
	cmd := strings.TrimSpace(in.Command)
	if cmd == "" {
		return entity.AutomationCommandExecution{}, fmt.Errorf("%w: command is required", apperr.ErrValidation)
	}
	status := strings.TrimSpace(in.Status)
	switch status {
	case constant.AutomationCommandSuccess, constant.AutomationCommandFailed, constant.AutomationCommandSkipped:
	default:
		return entity.AutomationCommandExecution{}, fmt.Errorf("%w: invalid command status", apperr.ErrValidation)
	}
	id, err := idgen.NewUUIDv7()
	if err != nil {
		return entity.AutomationCommandExecution{}, err
	}
	started := in.StartedAt.UTC()
	completed := in.CompletedAt.UTC()
	if completed.Before(started) {
		completed = started
	}
	dur := int(completed.Sub(started).Milliseconds())
	var cmdType *string
	if t := strings.TrimSpace(in.CommandType); t != "" {
		cmdType = &t
	}
	var resp *string
	if r := strings.TrimSpace(in.Response); r != "" {
		resp = &r
	}
	var errMsg *string
	if e := strings.TrimSpace(in.Error); e != "" {
		errMsg = &e
	}
	return s.execs.CreateCommandExecution(ctx, entity.AutomationCommandExecution{
		ID:          id,
		PublicID:    idgen.PublicID(constant.PublicIDPrefixAutomationCommandExecution, id),
		ExecutionID: in.ExecutionID,
		OrderIndex:  in.OrderIndex,
		Command:     cmd,
		CommandType: cmdType,
		StartedAt:   started,
		CompletedAt: &completed,
		Status:      status,
		DurationMs:  &dur,
		Response:    resp,
		Error:       errMsg,
		CreatedAt:   s.now().UTC(),
	})
}

// CompleteExecutionInput finalizes a RUNNING execution.
type CompleteExecutionInput struct {
	ExecutionID     uuid.UUID
	Status          string
	CommandsTotal   int
	CommandsSuccess int
	CommandsFailed  int
	DeliveryStatus  string
	Response        string
	Error           string
	StartedAt       time.Time
}

// CompleteExecution marks an execution terminal and stores metrics.
func (s *AutomationExecutionService) CompleteExecution(ctx context.Context, in CompleteExecutionInput) (entity.AutomationExecution, error) {
	if in.ExecutionID == uuid.Nil {
		return entity.AutomationExecution{}, fmt.Errorf("%w: execution id is required", apperr.ErrValidation)
	}
	status := strings.TrimSpace(in.Status)
	switch status {
	case constant.AutomationExecutionSuccess, constant.AutomationExecutionPartialSuccess,
		constant.AutomationExecutionFailed, constant.AutomationExecutionCancelled:
	default:
		return entity.AutomationExecution{}, fmt.Errorf("%w: invalid execution status", apperr.ErrValidation)
	}
	completed := s.now().UTC()
	started := in.StartedAt.UTC()
	if started.IsZero() {
		started = completed
	}
	dur := int(completed.Sub(started).Milliseconds())
	if dur < 0 {
		dur = 0
	}
	var resp *string
	if r := strings.TrimSpace(in.Response); r != "" {
		resp = &r
	}
	var errMsg *string
	if e := strings.TrimSpace(in.Error); e != "" {
		errMsg = &e
	}
	delivery := strings.TrimSpace(in.DeliveryStatus)
	if delivery == "" {
		delivery = constant.AutomationDeliverySkipped
	}
	return s.execs.CompleteExecution(ctx, in.ExecutionID, repository.AutomationExecutionCompleteFields{
		CompletedAt:     completed,
		Status:          status,
		DurationMs:      dur,
		CommandsTotal:   in.CommandsTotal,
		CommandsSuccess: in.CommandsSuccess,
		CommandsFailed:  in.CommandsFailed,
		DeliveryStatus:  delivery,
		Response:        resp,
		Error:           errMsg,
		UpdatedAt:       completed,
	})
}

// GetExecution returns an execution with command breakdown for the owner.
func (s *AutomationExecutionService) GetExecution(ctx context.Context, userID, executionID uuid.UUID) (entity.AutomationExecution, error) {
	if userID == uuid.Nil || executionID == uuid.Nil {
		return entity.AutomationExecution{}, fmt.Errorf("%w: user and execution id are required", apperr.ErrValidation)
	}
	exec, err := s.execs.GetExecutionByID(ctx, executionID, userID)
	if err != nil {
		return entity.AutomationExecution{}, err
	}
	cmds, err := s.execs.ListCommandExecutions(ctx, executionID)
	if err != nil {
		return entity.AutomationExecution{}, err
	}
	exec.Commands = cmds
	return exec, nil
}

// ListHistoryForAutomation returns recent executions for one automation.
func (s *AutomationExecutionService) ListHistoryForAutomation(
	ctx context.Context,
	userID, automationID uuid.UUID,
	limit int,
) ([]entity.AutomationExecution, error) {
	if userID == uuid.Nil || automationID == uuid.Nil {
		return nil, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	auto, err := s.autos.GetByID(ctx, automationID)
	if err != nil {
		return nil, err
	}
	if auto.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return s.execs.ListByAutomation(ctx, automationID, userID, limit)
}

// ListHistoryForUser returns recent executions across all automations.
func (s *AutomationExecutionService) ListHistoryForUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AutomationExecution, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.execs.ListByUser(ctx, userID, limit)
}

// MetricsForAutomations returns run metrics keyed by automation id.
func (s *AutomationExecutionService) MetricsForAutomations(
	ctx context.Context,
	userID uuid.UUID,
	automationIDs []uuid.UUID,
) (map[uuid.UUID]entity.AutomationRunMetrics, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return s.execs.MetricsByAutomations(ctx, userID, automationIDs)
}

// AutomationAnalytics is the API-facing analytics summary.
type AutomationAnalytics struct {
	TotalExecutions            int
	SuccessRate                float64
	FailureRate                float64
	AverageDurationMs          *float64
	AverageCommandsPerRun      *float64
	MostFrequentAutomationID   *uuid.UUID
	MostFrequentAutomationName *string
}

// Analytics computes aggregate stats for the user.
func (s *AutomationExecutionService) Analytics(ctx context.Context, userID uuid.UUID) (AutomationAnalytics, error) {
	if userID == uuid.Nil {
		return AutomationAnalytics{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	row, err := s.execs.Analytics(ctx, userID)
	if err != nil {
		return AutomationAnalytics{}, err
	}
	out := AutomationAnalytics{
		TotalExecutions:       row.TotalExecutions,
		AverageDurationMs:     row.AvgDurationMs,
		AverageCommandsPerRun: row.AvgCommands,
		MostFrequentAutomationID: row.TopAutomationID,
	}
	if row.TotalExecutions > 0 {
		out.SuccessRate = float64(row.Successful) / float64(row.TotalExecutions)
		out.FailureRate = float64(row.Failed) / float64(row.TotalExecutions)
	}
	if row.TopAutomationID != nil && *row.TopAutomationID != uuid.Nil {
		if auto, err := s.autos.GetByID(ctx, *row.TopAutomationID); err == nil {
			name := auto.Name
			out.MostFrequentAutomationName = &name
		}
	}
	return out, nil
}

// DeriveExecutionStatus maps command success/fail counts to an execution status.
func DeriveExecutionStatus(success, failed, skipped int) string {
	total := success + failed + skipped
	if total == 0 {
		return constant.AutomationExecutionFailed
	}
	if failed == 0 {
		return constant.AutomationExecutionSuccess
	}
	if success == 0 {
		return constant.AutomationExecutionFailed
	}
	return constant.AutomationExecutionPartialSuccess
}
