package business

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

// TimelineService merges TimelineProviders into one chronologically sorted feed.
type TimelineService struct {
	providers []TimelineProvider
}

// TimelineServiceDeps wires TimelineService dependencies.
type TimelineServiceDeps struct {
	Providers []TimelineProvider
}

// NewTimelineService constructs a TimelineService.
func NewTimelineService(deps TimelineServiceDeps) *TimelineService {
	providers := deps.Providers
	if providers == nil {
		providers = []TimelineProvider{}
	}
	return &TimelineService{providers: providers}
}

// ProviderCount returns how many non-nil TimelineProviders are configured.
func (s *TimelineService) ProviderCount() int {
	if s == nil {
		return 0
	}
	n := 0
	for _, p := range s.providers {
		if p != nil {
			n++
		}
	}
	return n
}

// List returns a chronologically sorted unified timeline for [from, to).
func (s *TimelineService) List(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	items, _, err := s.listWithStats(ctx, userID, from, to)
	return items, err
}

func (s *TimelineService) listWithStats(
	ctx context.Context,
	userID uuid.UUID,
	from, to time.Time,
) ([]entity.TimelineItem, FeedBuildStats, error) {
	stats := FeedBuildStats{FeedSource: FeedSourceTimeline}
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

	items := make([]entity.TimelineItem, 0)
	for _, provider := range s.providers {
		if provider == nil {
			continue
		}
		stats.ProvidersQueried++
		chunk, err := provider.Fetch(ctx, userID, from, to)
		if err != nil {
			return nil, stats, err
		}
		// Each provider issues one range query against its repository.
		stats.DatabaseQueries++
		stats.OccurrencesReturned += len(chunk)
		items = append(items, chunk...)
	}

	// Recurrence expansion lives inside TimelineProviders today, so the
	// service-level expansion/dedup stages are identity. Task 4's Occurrence
	// path fills these from real pipeline stages for before/after comparison.
	stats.AfterExpansion = stats.OccurrencesReturned
	stats.AfterDedup = stats.OccurrencesReturned

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartAt.Equal(items[j].StartAt) {
			return items[i].Title < items[j].Title
		}
		return items[i].StartAt.Before(items[j].StartAt)
	})
	return items, stats, nil
}
