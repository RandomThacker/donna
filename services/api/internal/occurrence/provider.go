package occurrence

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Provider lists scheduling Occurrences for a user in [from, to).
// Implementations live in occurrence/provider; this interface lives here so
// OccurrenceService can depend on providers without an import cycle.
type Provider interface {
	ListOccurrences(
		ctx context.Context,
		userID uuid.UUID,
		from time.Time,
		to time.Time,
	) ([]Occurrence, error)
}
