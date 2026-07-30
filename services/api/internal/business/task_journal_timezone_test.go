package business

import (
	"context"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type stubJournalUsers struct {
	tz string
}

func (s stubJournalUsers) Create(context.Context, entity.User) (entity.User, error) {
	panic("unused")
}
func (s stubJournalUsers) GetByID(_ context.Context, id uuid.UUID) (entity.User, error) {
	return entity.User{ID: id, Timezone: s.tz}, nil
}
func (s stubJournalUsers) GetByEmail(context.Context, string) (entity.User, error) {
	return entity.User{}, apperr.ErrNotFound
}
func (s stubJournalUsers) Update(context.Context, uuid.UUID, repository.UserUpdateFields, time.Time) (entity.User, error) {
	panic("unused")
}
func (s stubJournalUsers) SoftDelete(context.Context, uuid.UUID, string, time.Time) error {
	panic("unused")
}
func (s stubJournalUsers) TouchLastLogin(context.Context, uuid.UUID, time.Time) (entity.User, error) {
	panic("unused")
}
func (s stubJournalUsers) ListActiveIDs(context.Context) ([]uuid.UUID, error) {
	panic("unused")
}
func (s stubJournalUsers) WithTx(pgx.Tx) repository.UserRepository { return s }

type stubOccurrences struct {
	incomplete map[string][]entity.TaskOccurrence // date key YYYY-MM-DD
	created    []entity.TaskOccurrence
	exists     map[string]bool // taskID|date
}

func (s *stubOccurrences) key(t time.Time) string { return civilDate(t).Format("2006-01-02") }

func (s *stubOccurrences) Create(_ context.Context, occ entity.TaskOccurrence) (entity.TaskOccurrence, error) {
	s.created = append(s.created, occ)
	s.exists[occ.TaskID.String()+"|"+s.key(occ.Date)] = true
	return occ, nil
}
func (s *stubOccurrences) CountByUserDate(context.Context, uuid.UUID, time.Time) (int, error) {
	panic("unused")
}
func (s *stubOccurrences) ListByUserDate(context.Context, uuid.UUID, time.Time) ([]entity.TaskOccurrenceWithTask, error) {
	panic("unused")
}
func (s *stubOccurrences) GetByID(context.Context, uuid.UUID) (entity.TaskOccurrenceWithTask, error) {
	panic("unused")
}
func (s *stubOccurrences) ListIncompleteByUserDate(_ context.Context, _ uuid.UUID, date time.Time) ([]entity.TaskOccurrence, error) {
	return append([]entity.TaskOccurrence(nil), s.incomplete[s.key(date)]...), nil
}
func (s *stubOccurrences) MaxSortOrder(context.Context, uuid.UUID, time.Time) (int, error) {
	return -1, nil
}
func (s *stubOccurrences) UpdateCompletion(context.Context, uuid.UUID, uuid.UUID, bool, *time.Time, time.Time) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *stubOccurrences) CompleteIncompleteForTask(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (int64, error) {
	panic("unused")
}
func (s *stubOccurrences) SyncIncompleteFromCompletedPeers(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *stubOccurrences) UpdateSortOrder(context.Context, uuid.UUID, uuid.UUID, int, time.Time, time.Time) error {
	panic("unused")
}
func (s *stubOccurrences) BumpSortOrders(context.Context, uuid.UUID, time.Time, int, time.Time) error {
	panic("unused")
}
func (s *stubOccurrences) UpdateDateAndSort(context.Context, uuid.UUID, uuid.UUID, time.Time, int, time.Time) (entity.TaskOccurrence, error) {
	panic("unused")
}
func (s *stubOccurrences) SummariesByDateRange(context.Context, uuid.UUID, time.Time, time.Time) ([]entity.TaskDaySummary, error) {
	panic("unused")
}
func (s *stubOccurrences) ExistsForTaskDate(_ context.Context, taskID uuid.UUID, date time.Time) (bool, error) {
	return s.exists[taskID.String()+"|"+s.key(date)], nil
}
func (s *stubOccurrences) DeleteCarryForwardAfter(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}
func (s *stubOccurrences) DeleteByTaskID(context.Context, uuid.UUID, uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *stubOccurrences) WithTx(pgx.Tx) repository.TaskOccurrenceRepository { return s }

func TestEnsureDayUsesUserTimezoneNotUTC(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000401")
	taskID := uuid.MustParse("018f0000-0000-7000-8000-000000000402")
	prev := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	target := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	// 00:18 IST on Jul 31 == 18:48 UTC on Jul 30 — UTC "today" is still Jul 30.
	now := time.Date(2026, 7, 30, 18, 48, 0, 0, time.UTC)
	occs := &stubOccurrences{
		incomplete: map[string][]entity.TaskOccurrence{
			"2026-07-30": {{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000403"),
				TaskID: taskID, UserID: userID, Date: prev, Completed: false,
			}},
		},
		exists: map[string]bool{},
	}
	svc := NewTaskJournalService(nil, occs, nil, nil, stubJournalUsers{tz: "Asia/Kolkata"})
	svc.now = func() time.Time { return now }

	if err := svc.EnsureDay(context.Background(), userID, target); err != nil {
		t.Fatal(err)
	}
	if len(occs.created) != 1 {
		t.Fatalf("expected carry-forward clone, got %d", len(occs.created))
	}
	got := occs.created[0]
	if !got.CarriedForward || got.Source != constant.TaskOccurrenceSourceCarryForward {
		t.Fatalf("clone = %+v", got)
	}
	if !civilDate(got.Date).Equal(target) {
		t.Fatalf("clone date = %v", got.Date)
	}
}

func TestEnsureDayStillBlocksTrueFuture(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000404")
	occs := &stubOccurrences{incomplete: map[string][]entity.TaskOccurrence{}, exists: map[string]bool{}}
	svc := NewTaskJournalService(nil, occs, nil, nil, stubJournalUsers{tz: "Asia/Kolkata"})
	// Local today Jul 31 IST.
	svc.now = func() time.Time { return time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC) }

	future := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if err := svc.EnsureDay(context.Background(), userID, future); err != nil {
		t.Fatal(err)
	}
	if len(occs.created) != 0 {
		t.Fatalf("future day should stay empty, got %d", len(occs.created))
	}
}

func TestCivilDateInIST(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	// Still Jul 30 UTC, but already Jul 31 in IST.
	now := time.Date(2026, 7, 30, 18, 48, 0, 0, time.UTC)
	got := civilDateIn(now, loc)
	want := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
