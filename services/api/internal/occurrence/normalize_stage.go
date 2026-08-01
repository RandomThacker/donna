package occurrence

// normalizeOccurrences trims fields and rejects invalid scheduling units.
func normalizeOccurrences(items []Occurrence) ([]Occurrence, error) {
	if len(items) == 0 {
		return nil, nil
	}

	out := make([]Occurrence, 0, len(items))
	for _, item := range items {
		normalized := item.Normalize()
		if err := normalized.Validate(); err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}
