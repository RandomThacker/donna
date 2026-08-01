// Package provider supplies scheduling-oriented OccurrenceProvider implementations.
//
// These are intentionally separate from TimelineProvider (presentation).
// Providers query repositories, expand recurrence, and return Occurrence values.
// They do not merge, sort, apply notification policy, or depend on Timeline/Notifications.
package provider

import (
	"github.com/RandomThacker/donna/services/api/internal/occurrence"
)

// OccurrenceProvider lists scheduling Occurrences for a user in [from, to).
// Alias of occurrence.Provider so implementations and the service share one contract.
type OccurrenceProvider = occurrence.Provider
