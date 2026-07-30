package actions

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/google/uuid"
)

type taskDayAdapter struct {
	svc *business.TaskJournalService
}

func (a *taskDayAdapter) ListDayTaskOccurrences(
	ctx context.Context,
	userID uuid.UUID,
	date time.Time,
) ([]TaskOccurrenceResult, error) {
	view, err := a.svc.GetDay(ctx, userID, date)
	if err != nil {
		return nil, err
	}
	out := make([]TaskOccurrenceResult, 0, len(view.Occurrences))
	for _, o := range view.Occurrences {
		out = append(out, taskOccurrenceFromEntity(o))
	}
	return out, nil
}
