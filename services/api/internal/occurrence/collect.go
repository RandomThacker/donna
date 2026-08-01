package occurrence

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// collectFromProviders gathers Occurrences from every provider for [from, to).
// It does not merge-sort or expand — later pipeline stages own those jobs.
func collectFromProviders(
	ctx context.Context,
	providers []Provider,
	userID uuid.UUID,
	from, to time.Time,
) ([]Occurrence, error) {
	if len(providers) == 0 {
		return nil, nil
	}

	out := make([]Occurrence, 0)
	for _, p := range providers {
		if p == nil {
			continue
		}
		chunk, err := p.ListOccurrences(ctx, userID, from, to)
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			continue
		}
		out = append(out, chunk...)
	}
	return out, nil
}
