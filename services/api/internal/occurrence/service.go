package occurrence

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/google/uuid"
)

// Service builds upcoming scheduling Occurrences from Providers.
//
// It is the scheduling counterpart to TimelineService (presentation).
// It does not apply notification policy, generate notifications, or map to UI.
type Service interface {
	ListUpcoming(
		ctx context.Context,
		userID uuid.UUID,
		from time.Time,
		to time.Time,
	) ([]Occurrence, error)
}

// ServiceDeps wires OccurrenceService dependencies.
type ServiceDeps struct {
	Providers []Provider
}

type service struct {
	providers []Provider
}

// NewService constructs an OccurrenceService.
func NewService(deps ServiceDeps) Service {
	providers := deps.Providers
	if providers == nil {
		providers = []Provider{}
	}
	return &service{providers: providers}
}

// ProviderCount returns how many non-nil providers are configured.
func ProviderCount(svc Service) int {
	concrete, ok := svc.(*service)
	if !ok || concrete == nil {
		return 0
	}
	n := 0
	for _, p := range concrete.providers {
		if p != nil {
			n++
		}
	}
	return n
}

// ListStats captures pipeline stage sizes for scheduler instrumentation.
type ListStats struct {
	ProvidersQueried    int
	OccurrencesReturned int
	AfterExpansion      int
	AfterDedup          int
	DatabaseQueries     int
}

// ListUpcoming collects, expands, normalizes, sorts, and deduplicates
// Occurrences in [from, to).
func (s *service) ListUpcoming(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]Occurrence, error) {
	items, _, err := s.listUpcomingWithStats(ctx, userID, from, to)
	return items, err
}

// ListUpcomingWithStats is ListUpcoming plus pipeline counters for tick metrics.
// Available on the concrete service via type assert when the scheduler cuts over.
func ListUpcomingWithStats(
	svc Service,
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]Occurrence, ListStats, error) {
	if concrete, ok := svc.(*service); ok {
		return concrete.listUpcomingWithStats(ctx, userID, from, to)
	}
	items, err := svc.ListUpcoming(ctx, userID, from, to)
	return items, ListStats{}, err
}

func (s *service) listUpcomingWithStats(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]Occurrence, ListStats, error) {
	stats := ListStats{}
	if userID == uuid.Nil {
		return nil, stats, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		return nil, stats, fmt.Errorf("%w: from and to are required", apperr.ErrValidation)
	}
	if !to.After(from) {
		return nil, stats, fmt.Errorf("%w: to must be after from", apperr.ErrValidation)
	}
	from = from.UTC()
	to = to.UTC()

	for _, p := range s.providers {
		if p != nil {
			stats.ProvidersQueried++
		}
	}

	collected, err := collectFromProviders(ctx, s.providers, userID, from, to)
	if err != nil {
		return nil, stats, err
	}
	stats.OccurrencesReturned = len(collected)
	stats.DatabaseQueries = stats.ProvidersQueried

	expanded, err := expandSeriesTemplates(collected, from, to)
	if err != nil {
		return nil, stats, err
	}
	stats.AfterExpansion = len(expanded)

	normalized, err := normalizeOccurrences(expanded)
	if err != nil {
		return nil, stats, err
	}

	sorted := sortOccurrences(normalized)
	out := dedupeOccurrences(sorted)
	stats.AfterDedup = len(out)
	return out, stats, nil
}
