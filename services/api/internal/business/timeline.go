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

// List returns a chronologically sorted unified timeline for [from, to).
func (s *TimelineService) List(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", apperr.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("%w: from and to are required", apperr.ErrValidation)
	}
	if !to.After(from) {
		return nil, fmt.Errorf("%w: to must be after from", apperr.ErrValidation)
	}
	from = from.UTC()
	to = to.UTC()

	items := make([]entity.TimelineItem, 0)
	for _, provider := range s.providers {
		if provider == nil {
			continue
		}
		chunk, err := provider.Fetch(ctx, userID, from, to)
		if err != nil {
			return nil, err
		}
		items = append(items, chunk...)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartAt.Equal(items[j].StartAt) {
			return items[i].Title < items[j].Title
		}
		return items[i].StartAt.Before(items[j].StartAt)
	})
	return items, nil
}
