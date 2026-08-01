package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// AutomationCommandExecutionResult is one command within an execution.
type AutomationCommandExecutionResult struct {
	ID          uuid.UUID
	PublicID    string
	OrderIndex  int
	Command     string
	CommandType *string
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      string
	DurationMs  *int
	Response    *string
	Error       *string
}

// AutomationExecutionResult is a recorded automation run.
type AutomationExecutionResult struct {
	ID               uuid.UUID
	PublicID         string
	AutomationID     uuid.UUID
	AutomationName   *string
	UserID           uuid.UUID
	StartedAt        time.Time
	CompletedAt      *time.Time
	Status           string
	DurationMs       *int
	CommandsTotal    int
	CommandsSuccess  int
	CommandsFailed   int
	TriggerSource    string
	DeliveryChannels []string
	DeliveryStatus   *string
	Response         *string
	Error            *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Commands         []AutomationCommandExecutionResult
}

func automationExecutionFromEntity(e entity.AutomationExecution) AutomationExecutionResult {
	out := AutomationExecutionResult{
		ID:               e.ID,
		PublicID:         e.PublicID,
		AutomationID:     e.AutomationID,
		AutomationName:   e.AutomationName,
		UserID:           e.UserID,
		StartedAt:        e.StartedAt,
		CompletedAt:      e.CompletedAt,
		Status:           e.Status,
		DurationMs:       e.DurationMs,
		CommandsTotal:    e.CommandsTotal,
		CommandsSuccess:  e.CommandsSuccess,
		CommandsFailed:   e.CommandsFailed,
		TriggerSource:    e.TriggerSource,
		DeliveryChannels: e.DeliveryChannels,
		DeliveryStatus:   e.DeliveryStatus,
		Response:         e.Response,
		Error:            e.Error,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
	if len(e.Commands) > 0 {
		out.Commands = make([]AutomationCommandExecutionResult, 0, len(e.Commands))
		for _, c := range e.Commands {
			out.Commands = append(out.Commands, AutomationCommandExecutionResult{
				ID:          c.ID,
				PublicID:    c.PublicID,
				OrderIndex:  c.OrderIndex,
				Command:     c.Command,
				CommandType: c.CommandType,
				StartedAt:   c.StartedAt,
				CompletedAt: c.CompletedAt,
				Status:      c.Status,
				DurationMs:  c.DurationMs,
				Response:    c.Response,
				Error:       c.Error,
			})
		}
	}
	return out
}

// GetAutomationExecutionAction loads one execution with command breakdown.
type GetAutomationExecutionAction struct {
	execs AutomationExecutionQueryService
}

// NewGetAutomationExecutionAction constructs GetAutomationExecutionAction.
func NewGetAutomationExecutionAction(execs AutomationExecutionQueryService) *GetAutomationExecutionAction {
	return &GetAutomationExecutionAction{execs: execs}
}

// Execute returns execution detail.
func (a *GetAutomationExecutionAction) Execute(ctx context.Context, userID, executionID uuid.UUID) (AutomationExecutionResult, error) {
	if a.execs == nil {
		return AutomationExecutionResult{}, fmt.Errorf("%w: execution service is required", apperr.ErrInvalid)
	}
	exec, err := a.execs.GetExecution(ctx, userID, executionID)
	if err != nil {
		return AutomationExecutionResult{}, err
	}
	return automationExecutionFromEntity(exec), nil
}

// ListAutomationHistoryAction lists executions for one automation.
type ListAutomationHistoryAction struct {
	execs AutomationExecutionQueryService
}

// NewListAutomationHistoryAction constructs ListAutomationHistoryAction.
func NewListAutomationHistoryAction(execs AutomationExecutionQueryService) *ListAutomationHistoryAction {
	return &ListAutomationHistoryAction{execs: execs}
}

// Execute returns history for an automation.
func (a *ListAutomationHistoryAction) Execute(ctx context.Context, userID, automationID uuid.UUID, limit int) ([]AutomationExecutionResult, error) {
	if a.execs == nil {
		return nil, fmt.Errorf("%w: execution service is required", apperr.ErrInvalid)
	}
	rows, err := a.execs.ListHistoryForAutomation(ctx, userID, automationID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]AutomationExecutionResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, automationExecutionFromEntity(row))
	}
	return out, nil
}

// ListAllAutomationHistoryAction lists recent executions across automations.
type ListAllAutomationHistoryAction struct {
	execs AutomationExecutionQueryService
}

// NewListAllAutomationHistoryAction constructs ListAllAutomationHistoryAction.
func NewListAllAutomationHistoryAction(execs AutomationExecutionQueryService) *ListAllAutomationHistoryAction {
	return &ListAllAutomationHistoryAction{execs: execs}
}

// Execute returns global history for the user.
func (a *ListAllAutomationHistoryAction) Execute(ctx context.Context, userID uuid.UUID, limit int) ([]AutomationExecutionResult, error) {
	if a.execs == nil {
		return nil, fmt.Errorf("%w: execution service is required", apperr.ErrInvalid)
	}
	rows, err := a.execs.ListHistoryForUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]AutomationExecutionResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, automationExecutionFromEntity(row))
	}
	return out, nil
}

// GetAutomationAnalyticsAction returns aggregate execution analytics.
type GetAutomationAnalyticsAction struct {
	execs AutomationExecutionQueryService
}

// NewGetAutomationAnalyticsAction constructs GetAutomationAnalyticsAction.
func NewGetAutomationAnalyticsAction(execs AutomationExecutionQueryService) *GetAutomationAnalyticsAction {
	return &GetAutomationAnalyticsAction{execs: execs}
}

// Execute returns analytics.
func (a *GetAutomationAnalyticsAction) Execute(ctx context.Context, userID uuid.UUID) (business.AutomationAnalytics, error) {
	if a.execs == nil {
		return business.AutomationAnalytics{}, fmt.Errorf("%w: execution service is required", apperr.ErrInvalid)
	}
	return a.execs.Analytics(ctx, userID)
}
