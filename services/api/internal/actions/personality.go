package actions

import (
	"context"
	"fmt"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/personality"
	"github.com/google/uuid"
)

// PersonalityServicePort is the service port for personality actions.
type PersonalityServicePort interface {
	Get(ctx context.Context, userID uuid.UUID) (personality.Profile, error)
	Update(ctx context.Context, userID uuid.UUID, in business.PersonalityUpdateInput) (personality.Profile, error)
	ListDefinitions() ([]personality.Definition, error)
	Preview(ctx context.Context, userID uuid.UUID, renderer personality.Renderer, override *personality.Profile, timezone string) (map[string]string, error)
}

// GetPersonalityAction returns the current user's personality profile.
type GetPersonalityAction struct {
	svc PersonalityServicePort
}

// NewGetPersonalityAction constructs GetPersonalityAction.
func NewGetPersonalityAction(svc PersonalityServicePort) *GetPersonalityAction {
	return &GetPersonalityAction{svc: svc}
}

// Execute loads the profile.
func (a *GetPersonalityAction) Execute(ctx context.Context, userID uuid.UUID) (personality.Profile, error) {
	if a.svc == nil {
		return personality.Profile{}, fmt.Errorf("%w: personality service is required", apperr.ErrInvalid)
	}
	if userID == uuid.Nil {
		return personality.Profile{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return a.svc.Get(ctx, userID)
}

// UpdatePersonalityAction patches personality preferences.
type UpdatePersonalityAction struct {
	svc PersonalityServicePort
}

// NewUpdatePersonalityAction constructs UpdatePersonalityAction.
func NewUpdatePersonalityAction(svc PersonalityServicePort) *UpdatePersonalityAction {
	return &UpdatePersonalityAction{svc: svc}
}

// Execute updates the profile.
func (a *UpdatePersonalityAction) Execute(
	ctx context.Context,
	userID uuid.UUID,
	in business.PersonalityUpdateInput,
) (personality.Profile, error) {
	if a.svc == nil {
		return personality.Profile{}, fmt.Errorf("%w: personality service is required", apperr.ErrInvalid)
	}
	if userID == uuid.Nil {
		return personality.Profile{}, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return a.svc.Update(ctx, userID, in)
}

// ListPersonalityCatalogAction lists built-in personalities.
type ListPersonalityCatalogAction struct {
	svc PersonalityServicePort
}

// NewListPersonalityCatalogAction constructs ListPersonalityCatalogAction.
func NewListPersonalityCatalogAction(svc PersonalityServicePort) *ListPersonalityCatalogAction {
	return &ListPersonalityCatalogAction{svc: svc}
}

// Execute returns catalog definitions.
func (a *ListPersonalityCatalogAction) Execute() ([]personality.Definition, error) {
	if a.svc == nil {
		return nil, fmt.Errorf("%w: personality service is required", apperr.ErrInvalid)
	}
	return a.svc.ListDefinitions()
}

// PreviewPersonalityAction renders live preview samples.
type PreviewPersonalityAction struct {
	svc      PersonalityServicePort
	renderer personality.Renderer
}

// NewPreviewPersonalityAction constructs PreviewPersonalityAction.
func NewPreviewPersonalityAction(svc PersonalityServicePort, renderer personality.Renderer) *PreviewPersonalityAction {
	return &PreviewPersonalityAction{svc: svc, renderer: renderer}
}

// PreviewPersonalityRequest is the preview input.
type PreviewPersonalityRequest struct {
	UserID    uuid.UUID
	Timezone  string
	Override  *personality.Profile
}

// Execute returns sample rendered strings keyed by scenario.
func (a *PreviewPersonalityAction) Execute(ctx context.Context, req PreviewPersonalityRequest) (map[string]string, error) {
	if a.svc == nil || a.renderer == nil {
		return nil, fmt.Errorf("%w: personality preview is not configured", apperr.ErrInvalid)
	}
	if req.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	return a.svc.Preview(ctx, req.UserID, a.renderer, req.Override, req.Timezone)
}
