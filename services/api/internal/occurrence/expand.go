package occurrence

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/recurrence"
)

// expandSeriesTemplates expands series roots that still carry an RRULE and
// have no ParentID. Already-expanded instances (ParentID set) pass through.
//
// Providers may expand themselves; this stage stays so EXDATE / RECURRENCE-ID /
// overrides can land here later without growing ListUpcoming.
func expandSeriesTemplates(items []Occurrence, from, to time.Time) ([]Occurrence, error) {
	if len(items) == 0 {
		return nil, nil
	}

	out := make([]Occurrence, 0, len(items))
	for _, item := range items {
		if item.ParentID != nil || item.RecurrenceRule == nil {
			out = append(out, item)
			continue
		}

		rule, ok := recurrence.NormalizeRule(*item.RecurrenceRule)
		if !ok {
			out = append(out, item)
			continue
		}

		occs, err := recurrence.Expand(rule, item.StartAt, item.EndAt, item.Timezone, from, to)
		if err != nil {
			return nil, err
		}
		if len(occs) == 0 {
			continue
		}

		parentID := item.ID
		ruleCopy := rule
		for _, occ := range occs {
			child := item
			child.ParentID = &parentID
			child.StartAt = occ.Start
			child.EndAt = occ.End
			child.OccurrenceID = recurrence.ID(parentID, occ.Start)
			child.ID = child.OccurrenceID
			child.RecurrenceRule = &ruleCopy
			out = append(out, child)
		}
	}
	return out, nil
}
