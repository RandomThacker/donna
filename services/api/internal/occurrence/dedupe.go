package occurrence

// dedupeOccurrences drops later duplicates with the same OccurrenceID.
// Call after sortOccurrences so the chronologically / priority-first row wins.
func dedupeOccurrences(items []Occurrence) []Occurrence {
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]Occurrence, 0, len(items))
	for _, item := range items {
		id := item.OccurrenceID
		if id == "" {
			out = append(out, item)
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, item)
	}
	return out
}
