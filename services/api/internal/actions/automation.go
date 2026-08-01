package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/automationcatalog"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// AutomationResult is the domain DTO returned by automation actions.
type AutomationResult struct {
	ID               uuid.UUID
	PublicID         string
	UserID           uuid.UUID
	Name             string
	Description      *string
	Enabled          bool
	TriggerType      string
	TriggerTime      string
	TriggerDays      []string
	Timezone         string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
	TemplateID       *string
	LastRunAt        *time.Time
	NextRunAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	// Execution metrics (optional; populated on list).
	LastStatus          *string
	SuccessRate         *float64
	AverageDurationMs   *float64
	LastCommandsTotal   *int
	LastCommandsSuccess *int
	TotalExecutions     int
}

func automationFromEntity(e entity.Automation) AutomationResult {
	return AutomationResult{
		ID:               e.ID,
		PublicID:         e.PublicID,
		UserID:           e.UserID,
		Name:             e.Name,
		Description:      e.Description,
		Enabled:          e.Enabled,
		TriggerType:      e.TriggerType,
		TriggerTime:      e.TriggerTime,
		TriggerDays:      e.TriggerDays,
		Timezone:         e.Timezone,
		Commands:         e.Commands,
		DeliveryChannels: e.DeliveryChannels,
		TemplateID:       e.TemplateID,
		LastRunAt:        e.LastRunAt,
		NextRunAt:        e.NextRunAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

// CreateAutomationRequest is the input for CreateAutomationAction.
type CreateAutomationRequest struct {
	UserID           uuid.UUID
	Name             string
	Description      *string
	Enabled          *bool
	TriggerType      string
	TriggerTime      string
	TriggerDays      []string
	Timezone         string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
	TemplateID       *string
}

// CreateAutomationAction creates an automation.
type CreateAutomationAction struct {
	autos AutomationService
}

// NewCreateAutomationAction constructs CreateAutomationAction.
func NewCreateAutomationAction(autos AutomationService) *CreateAutomationAction {
	return &CreateAutomationAction{autos: autos}
}

// Execute runs the create workflow.
func (a *CreateAutomationAction) Execute(ctx context.Context, req CreateAutomationRequest) (AutomationResult, error) {
	if a.autos == nil {
		return AutomationResult{}, fmt.Errorf("%w: automation service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return AutomationResult{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	created, err := a.autos.Create(ctx, req.UserID, business.CreateAutomationInput{
		Name:             req.Name,
		Description:      req.Description,
		Enabled:          req.Enabled,
		TriggerType:      req.TriggerType,
		TriggerTime:      req.TriggerTime,
		TriggerDays:      req.TriggerDays,
		Timezone:         req.Timezone,
		Commands:         req.Commands,
		DeliveryChannels: req.DeliveryChannels,
		TemplateID:       req.TemplateID,
	})
	if err != nil {
		return AutomationResult{}, err
	}
	return automationFromEntity(created), nil
}

// UpdateAutomationRequest is the input for UpdateAutomationAction.
type UpdateAutomationRequest struct {
	UserID           uuid.UUID
	AutomationID     uuid.UUID
	Name             *string
	Description      *string
	Enabled          *bool
	TriggerType      *string
	TriggerTime      *string
	TriggerDays      []string
	TriggerDaysSet   bool
	Timezone         *string
	Commands         []entity.AutomationCommand
	DeliveryChannels []string
}

// UpdateAutomationAction patches an automation.
type UpdateAutomationAction struct {
	autos AutomationService
}

// NewUpdateAutomationAction constructs UpdateAutomationAction.
func NewUpdateAutomationAction(autos AutomationService) *UpdateAutomationAction {
	return &UpdateAutomationAction{autos: autos}
}

// Execute runs the update workflow.
func (a *UpdateAutomationAction) Execute(ctx context.Context, req UpdateAutomationRequest) (AutomationResult, error) {
	if a.autos == nil {
		return AutomationResult{}, fmt.Errorf("%w: automation service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.AutomationID == uuid.Nil {
		return AutomationResult{}, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	updated, err := a.autos.Update(ctx, req.UserID, req.AutomationID, business.UpdateAutomationInput{
		Name:             req.Name,
		Description:      req.Description,
		Enabled:          req.Enabled,
		TriggerType:      req.TriggerType,
		TriggerTime:      req.TriggerTime,
		TriggerDays:      req.TriggerDays,
		TriggerDaysSet:   req.TriggerDaysSet,
		Timezone:         req.Timezone,
		Commands:         req.Commands,
		DeliveryChannels: req.DeliveryChannels,
	})
	if err != nil {
		return AutomationResult{}, err
	}
	return automationFromEntity(updated), nil
}

// DeleteAutomationRequest is the input for DeleteAutomationAction.
type DeleteAutomationRequest struct {
	UserID       uuid.UUID
	AutomationID uuid.UUID
}

// DeleteAutomationAction soft-deletes an automation.
type DeleteAutomationAction struct {
	autos AutomationService
}

// NewDeleteAutomationAction constructs DeleteAutomationAction.
func NewDeleteAutomationAction(autos AutomationService) *DeleteAutomationAction {
	return &DeleteAutomationAction{autos: autos}
}

// Execute runs the delete workflow.
func (a *DeleteAutomationAction) Execute(ctx context.Context, req DeleteAutomationRequest) error {
	if a.autos == nil {
		return fmt.Errorf("%w: automation service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.AutomationID == uuid.Nil {
		return fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	return a.autos.Delete(ctx, req.UserID, req.AutomationID)
}

// ListAutomationsRequest is the input for ListAutomationsAction.
type ListAutomationsRequest struct {
	UserID uuid.UUID
}

// ListAutomationsAction lists automations for a user.
type ListAutomationsAction struct {
	autos   AutomationService
	metrics AutomationExecutionQueryService
}

// NewListAutomationsAction constructs ListAutomationsAction.
func NewListAutomationsAction(autos AutomationService, metrics AutomationExecutionQueryService) *ListAutomationsAction {
	return &ListAutomationsAction{autos: autos, metrics: metrics}
}

// Execute runs the list workflow.
func (a *ListAutomationsAction) Execute(ctx context.Context, req ListAutomationsRequest) ([]AutomationResult, error) {
	if a.autos == nil {
		return nil, fmt.Errorf("%w: automation service is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	autos, err := a.autos.List(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(autos))
	for _, auto := range autos {
		ids = append(ids, auto.ID)
	}
	var metrics map[uuid.UUID]entity.AutomationRunMetrics
	if a.metrics != nil && len(ids) > 0 {
		metrics, _ = a.metrics.MetricsForAutomations(ctx, req.UserID, ids)
	}
	out := make([]AutomationResult, 0, len(autos))
	for _, auto := range autos {
		result := automationFromEntity(auto)
		if m, ok := metrics[auto.ID]; ok {
			result.LastStatus = m.LastStatus
			result.SuccessRate = m.SuccessRate
			result.AverageDurationMs = m.AverageDurationMs
			result.LastCommandsTotal = m.LastCommandsTotal
			result.LastCommandsSuccess = m.LastCommandsSuccess
			result.TotalExecutions = m.TotalExecutions
		}
		out = append(out, result)
	}
	return out, nil
}

// ListAutomationTemplatesAction lists Intent Catalog templates.
type ListAutomationTemplatesAction struct {
	autos AutomationService
}

// NewListAutomationTemplatesAction constructs ListAutomationTemplatesAction.
func NewListAutomationTemplatesAction(autos AutomationService) *ListAutomationTemplatesAction {
	return &ListAutomationTemplatesAction{autos: autos}
}

// Execute returns catalog templates.
func (a *ListAutomationTemplatesAction) Execute() ([]automationcatalog.Template, error) {
	if a.autos == nil {
		return nil, fmt.Errorf("%w: automation service is required", apperr.ErrInvalid)
	}
	return a.autos.ListTemplates()
}

// RunAutomationRequest is POST /automations/:id/run.
type RunAutomationRequest struct {
	UserID       uuid.UUID
	AutomationID uuid.UUID
	DisplayName  string
}

// PreviewAutomationRequest is POST /automations/:id/preview.
type PreviewAutomationRequest struct {
	UserID       uuid.UUID
	AutomationID uuid.UUID
	DisplayName  string
}

// AutomationRunnerPort executes automations (manual / preview).
type AutomationRunnerPort interface {
	Run(ctx context.Context, auto entity.Automation, opts business.AutomationRunOptions) (business.AutomationRunResult, error)
}

// AutomationOwnerLookup loads an owned automation template.
type AutomationOwnerLookup interface {
	GetOwned(ctx context.Context, userID, autoID uuid.UUID) (entity.Automation, error)
}

// RunAutomationAction runs an automation immediately (history + chat delivery).
type RunAutomationAction struct {
	autos  AutomationOwnerLookup
	runner AutomationRunnerPort
}

// NewRunAutomationAction constructs RunAutomationAction.
func NewRunAutomationAction(autos AutomationOwnerLookup, runner AutomationRunnerPort) *RunAutomationAction {
	return &RunAutomationAction{autos: autos, runner: runner}
}

// Execute runs the automation with trigger_source=manual.
func (a *RunAutomationAction) Execute(ctx context.Context, req RunAutomationRequest) (business.AutomationRunResult, error) {
	if a.autos == nil || a.runner == nil {
		return business.AutomationRunResult{}, fmt.Errorf("%w: automation runner is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.AutomationID == uuid.Nil {
		return business.AutomationRunResult{}, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	auto, err := a.autos.GetOwned(ctx, req.UserID, req.AutomationID)
	if err != nil {
		return business.AutomationRunResult{}, err
	}
	return a.runner.Run(ctx, auto, business.AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourceManual,
		RecordHistory:  true,
		DeliverToChat:  true,
		UpdateSchedule: false,
		DryRun:         false,
		DisplayName:    req.DisplayName,
	})
}

// PreviewAutomationAction dry-runs an automation (no history, delivery, or mutations).
type PreviewAutomationAction struct {
	autos  AutomationOwnerLookup
	runner AutomationRunnerPort
}

// NewPreviewAutomationAction constructs PreviewAutomationAction.
func NewPreviewAutomationAction(autos AutomationOwnerLookup, runner AutomationRunnerPort) *PreviewAutomationAction {
	return &PreviewAutomationAction{autos: autos, runner: runner}
}

// Execute previews the automation with dry-run chat execution.
func (a *PreviewAutomationAction) Execute(ctx context.Context, req PreviewAutomationRequest) (business.AutomationRunResult, error) {
	if a.autos == nil || a.runner == nil {
		return business.AutomationRunResult{}, fmt.Errorf("%w: automation runner is required", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil || req.AutomationID == uuid.Nil {
		return business.AutomationRunResult{}, fmt.Errorf("%w: user and automation id are required", apperr.ErrValidation)
	}
	auto, err := a.autos.GetOwned(ctx, req.UserID, req.AutomationID)
	if err != nil {
		return business.AutomationRunResult{}, err
	}
	return a.runner.Run(ctx, auto, business.AutomationRunOptions{
		TriggerSource:  constant.AutomationTriggerSourcePreview,
		RecordHistory:  false,
		DeliverToChat:  false,
		UpdateSchedule: false,
		DryRun:         true,
		DisplayName:    req.DisplayName,
	})
}
