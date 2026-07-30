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

type rescheduleOccStub struct {
	byID       map[uuid.UUID]entity.TaskOccurrenceWithTask
	exists     map[string]bool
	deletedIDs []uuid.UUID
}

func (s *rescheduleOccStub) key(taskID uuid.UUID, date time.Time) string {
	return taskID.String() + "|" + civilDate(date).Format("2006-01-02")
}

func (s *rescheduleOccStub) Create(context.Context, entity.TaskOccurrence) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *rescheduleOccStub) CountByUserDate(context.Context, uuid.UUID, time.Time) (int, error) {
	panic("unused")
}
func (s *rescheduleOccStub) ListByUserDate(context.Context, uuid.UUID, time.Time) ([]entity.TaskOccurrenceWithTask, error) {
	panic("unused")
}
func (s *rescheduleOccStub) GetByID(_ context.Context, id uuid.UUID) (entity.TaskOccurrenceWithTask, error) {
	occ, ok := s.byID[id]
	if !ok {
		panic("missing occurrence")
	}
	return occ, nil
}
func (s *rescheduleOccStub) ListIncompleteByUserDate(context.Context, uuid.UUID, time.Time) ([]entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *rescheduleOccStub) MaxSortOrder(context.Context, uuid.UUID, time.Time) (int, error) {
	panic("unused")
}
func (s *rescheduleOccStub) UpdateCompletion(context.Context, uuid.UUID, uuid.UUID, bool, *time.Time, time.Time) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *rescheduleOccStub) CompleteIncompleteForTask(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (int64, error) {
	panic("unused")
}
func (s *rescheduleOccStub) UncompleteForTask(context.Context, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	panic("unused")
}
func (s *rescheduleOccStub) SyncIncompleteFromCompletedPeers(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *rescheduleOccStub) UpdateSortOrder(context.Context, uuid.UUID, uuid.UUID, int, time.Time, time.Time) error {
	panic("unused")
}
func (s *rescheduleOccStub) BumpSortOrders(context.Context, uuid.UUID, time.Time, int, time.Time) error {
	return nil
}
func (s *rescheduleOccStub) UpdateDateAndSort(
	_ context.Context,
	id, _ uuid.UUID,
	date time.Time,
	sortOrder int,
	updatedAt time.Time,
) (entity.TaskOccurrence, error) {
	occ := s.byID[id]
	delete(s.exists, s.key(occ.TaskID, occ.Date))
	occ.Date = civilDate(date)
	occ.SortOrder = sortOrder
	occ.UpdatedAt = updatedAt
	occ.Source = constant.TaskOccurrenceSourceManual
	occ.CarriedForward = false
	s.byID[id] = occ
	s.exists[s.key(occ.TaskID, occ.Date)] = true
	return occ.TaskOccurrence, nil
}
func (s *rescheduleOccStub) SummariesByDateRange(context.Context, uuid.UUID, time.Time, time.Time) ([]entity.TaskDaySummary, error) {
	panic("unused")
}
func (s *rescheduleOccStub) ExistsForTaskDate(_ context.Context, taskID uuid.UUID, date time.Time) (bool, error) {
	return s.exists[s.key(taskID, date)], nil
}
func (s *rescheduleOccStub) DeleteIncompleteForTaskExcept(
	_ context.Context,
	taskID, _ uuid.UUID,
	keepID uuid.UUID,
) (int64, error) {
	var n int64
	for id, occ := range s.byID {
		if occ.TaskID != taskID || occ.Completed || id == keepID {
			continue
		}
		delete(s.exists, s.key(occ.TaskID, occ.Date))
		delete(s.byID, id)
		s.deletedIDs = append(s.deletedIDs, id)
		n++
	}
	return n, nil
}
func (s *rescheduleOccStub) DeleteCarryForwardAfter(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *rescheduleOccStub) DeleteByTaskID(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *rescheduleOccStub) WithTx(pgx.Tx) repository.TaskOccurrenceRepository { return s }

type stubTaskTags struct{}

func (stubTaskTags) Create(context.Context, entity.TaskTag) (entity.TaskTag, error) {
	panic("unused")
}
func (stubTaskTags) ListByUser(context.Context, uuid.UUID) ([]entity.TaskTag, error) {
	panic("unused")
}
func (stubTaskTags) GetByID(context.Context, uuid.UUID) (entity.TaskTag, error) {
	panic("unused")
}
func (stubTaskTags) Update(context.Context, uuid.UUID, uuid.UUID, repository.TaskTagUpdateFields, time.Time) (entity.TaskTag, error) {
	panic("unused")
}
func (stubTaskTags) Delete(context.Context, uuid.UUID, uuid.UUID) error { panic("unused") }
func (stubTaskTags) ReplaceTaskTags(context.Context, uuid.UUID, []uuid.UUID, time.Time) error {
	panic("unused")
}
func (stubTaskTags) RemoveAllForTask(context.Context, uuid.UUID) error { panic("unused") }
func (stubTaskTags) ListByTaskIDs(context.Context, uuid.UUID, []uuid.UUID) (map[uuid.UUID][]entity.TaskTag, error) {
	return map[uuid.UUID][]entity.TaskTag{}, nil
}
func (stubTaskTags) WithTx(pgx.Tx) repository.TaskTagRepository { return stubTaskTags{} }

func TestRescheduleOccurrencePromotesCarryForwardAndClearsSiblings(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000501")
	taskID := uuid.MustParse("018f0000-0000-7000-8000-000000000502")
	originID := uuid.MustParse("018f0000-0000-7000-8000-000000000503")
	cfID := uuid.MustParse("018f0000-0000-7000-8000-000000000504")
	originDay := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	origin := entity.TaskOccurrenceWithTask{
		TaskOccurrence: entity.TaskOccurrence{
			ID: originID, TaskID: taskID, UserID: userID, Date: originDay,
			Completed: false, Source: constant.TaskOccurrenceSourceManual,
		},
	}
	cf := entity.TaskOccurrenceWithTask{
		TaskOccurrence: entity.TaskOccurrence{
			ID: cfID, TaskID: taskID, UserID: userID, Date: today,
			Completed: false, CarriedForward: true,
			Source: constant.TaskOccurrenceSourceCarryForward,
		},
	}

	occs := &rescheduleOccStub{
		byID: map[uuid.UUID]entity.TaskOccurrenceWithTask{
			originID: origin,
			cfID:     cf,
		},
		exists: map[string]bool{
			taskID.String() + "|2026-07-30": true,
			taskID.String() + "|2026-07-31": true,
		},
	}

	svc := NewTaskJournalService(nil, occs, nil, stubTaskTags{}, stubJournalUsers{tz: "Asia/Kolkata"})
	svc.now = func() time.Time { return time.Date(2026, 7, 31, 8, 0, 0, 0, time.UTC) }

	moved, err := svc.RescheduleOccurrence(context.Background(), userID, cfID, tomorrow)
	if err != nil {
		t.Fatalf("RescheduleOccurrence: %v", err)
	}
	if !civilDate(moved.Date).Equal(tomorrow) {
		t.Fatalf("date = %s, want %s", moved.Date.Format("2006-01-02"), "2026-08-01")
	}
	if moved.Source != constant.TaskOccurrenceSourceManual {
		t.Fatalf("source = %q, want manual", moved.Source)
	}
	if moved.CarriedForward {
		t.Fatal("carried_forward should be false after reschedule")
	}
	if _, ok := occs.byID[originID]; ok {
		t.Fatal("origin incomplete sibling should be deleted")
	}
	if len(occs.deletedIDs) != 1 || occs.deletedIDs[0] != originID {
		t.Fatalf("deletedIDs = %v, want [%s]", occs.deletedIDs, originID)
	}
	// Future CF purge only deletes source=carry_forward — promoted row must survive.
	if moved.Source == constant.TaskOccurrenceSourceCarryForward {
		t.Fatal("moved row must not remain carry_forward")
	}
}
