package occurrence

import "sort"

// sourcePriority establishes a stable secondary sort key across sources.
// Lower values sort first. Unknown sources sort last.
func sourcePriority(source OccurrenceSource) int {
	switch source {
	case SourceGoogle:
		return 0
	case SourceMicrosoftICS:
		return 1
	case SourceDonna:
		return 2
	default:
		return 99
	}
}

func parentIDKey(parentID *string) string {
	if parentID == nil {
		return ""
	}
	return *parentID
}

// sortOccurrences orders by StartAt, then source priority, then ParentID.
// SliceStable keeps equal keys in collector / provider order.
func sortOccurrences(items []Occurrence) []Occurrence {
	if len(items) < 2 {
		return items
	}

	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if !a.StartAt.Equal(b.StartAt) {
			return a.StartAt.Before(b.StartAt)
		}
		pa, pb := sourcePriority(a.Source), sourcePriority(b.Source)
		if pa != pb {
			return pa < pb
		}
		return parentIDKey(a.ParentID) < parentIDKey(b.ParentID)
	})
	return items
}
