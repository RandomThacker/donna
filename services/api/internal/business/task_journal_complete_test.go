package business

import (
	"context"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type completeOccStub struct {
	byID map[uuid.UUID]entity.TaskOccurrenceWithTask
}

func (s *completeOccStub) Create(context.Context, entity.TaskOccurrence) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *completeOccStub) CountByUserDate(context.Context, uuid.UUID, time.Time) (int, error) {
	panic("unused")
}
func (s *completeOccStub) ListByUserDate(context.Context, uuid.UUID, time.Time) ([]entity.TaskOccurrenceWithTask, error) {
	panic("unused")
}
func (s *completeOccStub) GetByID(_ context.Context, id uuid.UUID) (entity.TaskOccurrenceWithTask, error) {
	return s.byID[id], nil
}
func (s *completeOccStub) ListIncompleteByUserDate(context.Context, uuid.UUID, time.Time) ([]entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *completeOccStub) MaxSortOrder(context.Context, uuid.UUID, time.Time) (int, error) {
	return 0, nil
}
func (s *completeOccStub) UpdateCompletion(context.Context, uuid.UUID, uuid.UUID, bool, *time.Time, time.Time) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *completeOccStub) CompleteIncompleteForTask(
	_ context.Context,
	taskID, _ uuid.UUID,
	completedAt, updatedAt time.Time,
) (int64, error) {
	var n int64
	for id, occ := range s.byID {
		if occ.TaskID != taskID || occ.Completed {
			continue
		}
		occ.Completed = true
		occ.CompletedAt = &completedAt
		occ.UpdatedAt = updatedAt
		s.byID[id] = occ
		n++
	}
	return n, nil
}
func (s *completeOccStub) UncompleteForTask(
	_ context.Context,
	taskID, _ uuid.UUID,
	updatedAt time.Time,
) (int64, error) {
	var n int64
	for id, occ := range s.byID {
		if occ.TaskID != taskID || !occ.Completed {
			continue
		}
		occ.Completed = false
		occ.CompletedAt = nil
		occ.UpdatedAt = updatedAt
		s.byID[id] = occ
		n++
	}
	return n, nil
}
func (s *completeOccStub) SyncIncompleteFromCompletedPeers(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *completeOccStub) UpdateSortOrder(context.Context, uuid.UUID, uuid.UUID, int, time.Time, time.Time) error {
	return nil
}
func (s *completeOccStub) BumpSortOrders(context.Context, uuid.UUID, time.Time, int, time.Time) error {
	panic("unused")
}
func (s *completeOccStub) UpdateDateAndSort(context.Context, uuid.UUID, uuid.UUID, time.Time, int, time.Time) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *completeOccStub) SummariesByDateRange(context.Context, uuid.UUID, time.Time, time.Time) ([]entity.TaskDaySummary, error) {
	panic("unused")
}
func (s *completeOccStub) ExistsForTaskDate(context.Context, uuid.UUID, time.Time) (bool, error) {
	panic("unused")
}
func (s *completeOccStub) DeleteIncompleteForTaskExcept(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *completeOccStub) DeleteCarryForwardAfter(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *completeOccStub) DeleteByTaskID(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *completeOccStub) WithTx(pgx.Tx) repository.TaskOccurrenceRepository { return s }

func TestUpdateOccurrenceUncompleteClearsAllPeers(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000601")
	taskID := uuid.MustParse("018f0000-0000-7000-8000-000000000602")
	originID := uuid.MustParse("018f0000-0000-7000-8000-000000000603")
	todayID := uuid.MustParse("018f0000-0000-7000-8000-000000000604")
	doneAt := time.Date(2026, 7, 31, 8, 0, 0, 0, time.UTC)

	origin := entity.TaskOccurrenceWithTask{
		TaskOccurrence: entity.TaskOccurrence{
			ID: originID, TaskID: taskID, UserID: userID,
			Date: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
			Completed: true, CompletedAt: &doneAt,
			Source: constant.TaskOccurrenceSourceManual,
		},
	}
	today := entity.TaskOccurrenceWithTask{
		TaskOccurrence: entity.TaskOccurrence{
			ID: todayID, TaskID: taskID, UserID: userID,
			Date: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
			Completed: true, CompletedAt: &doneAt,
			CarriedForward: true, Source: constant.TaskOccurrenceSourceCarryForward,
		},
	}

	occs := &completeOccStub{
		byID: map[uuid.UUID]entity.TaskOccurrenceWithTask{
			originID: origin,
			todayID:  today,
		},
	}
	svc := NewTaskJournalService(nil, occs, nil, stubTaskTags{}, stubJournalUsers{tz: "Asia/Kolkata"})
	svc.now = func() time.Time { return doneAt }

	out, err := svc.UpdateOccurrence(context.Background(), userID, todayID, false)
	if err != nil {
		t.Fatalf("UpdateOccurrence: %v", err)
	}
	if out.Completed {
		t.Fatal("clicked occurrence should be incomplete")
	}
	if occs.byID[originID].Completed {
		t.Fatal("peer occurrence should also be incomplete after uncheck")
	}
	if occs.byID[todayID].Completed {
		t.Fatal("today occurrence should be incomplete")
	}
}
