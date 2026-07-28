package scheduler_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/scheduler"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mockJobsRepo struct {
	due      []entity.SchedulerJob
	claimed  map[uuid.UUID]entity.SchedulerJob
	finished map[uuid.UUID]string
}

func (m *mockJobsRepo) Create(context.Context, entity.SchedulerJob) (entity.SchedulerJob, error) {
	return entity.SchedulerJob{}, errors.New("unused")
}
func (m *mockJobsRepo) ListDue(context.Context, []string, time.Time, int) ([]entity.SchedulerJob, error) {
	return m.due, nil
}
func (m *mockJobsRepo) GetPendingByTypeAndAccount(context.Context, string, uuid.UUID) (entity.SchedulerJob, error) {
	return entity.SchedulerJob{}, apperr.ErrNotFound
}
func (m *mockJobsRepo) Claim(_ context.Context, id uuid.UUID, startedAt time.Time) (entity.SchedulerJob, error) {
	for _, job := range m.due {
		if job.ID == id {
			job.Status = constant.SchedulerJobStatusRunning
			job.StartedAt = &startedAt
			if m.claimed == nil {
				m.claimed = map[uuid.UUID]entity.SchedulerJob{}
			}
			m.claimed[id] = job
			return job, nil
		}
	}
	return entity.SchedulerJob{}, apperr.ErrNotFound
}
func (m *mockJobsRepo) Finish(_ context.Context, id uuid.UUID, status string, _ *string, _ time.Time) (entity.SchedulerJob, error) {
	if m.finished == nil {
		m.finished = map[uuid.UUID]string{}
	}
	m.finished[id] = status
	job := m.claimed[id]
	job.Status = status
	return job, nil
}
func (m *mockJobsRepo) ReschedulePending(context.Context, uuid.UUID, time.Time, time.Time) (entity.SchedulerJob, error) {
	return entity.SchedulerJob{}, errors.New("unused")
}
func (m *mockJobsRepo) WithTx(pgx.Tx) repository.SchedulerJobRepository { return m }

type fakeJob struct {
	jobType     string
	runs        int
	scheduleN   int
	err         error
	scheduleErr error
}

func (f *fakeJob) Type() string { return f.jobType }
func (f *fakeJob) Run(context.Context, entity.SchedulerJob) error {
	f.runs++
	return f.err
}
func (f *fakeJob) ScheduleNext(context.Context, entity.SchedulerJob) error {
	f.scheduleN++
	return f.scheduleErr
}

func testSchedulerLog() *logger.Logger {
	return logger.NewFactory(logger.Options{
		Service:     "test",
		Environment: "test",
		Level:       "error",
		Output:      io.Discard,
	}).Module("scheduler")
}

func TestRunnerDispatchesByJobTypeAndReschedules(t *testing.T) {
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000701")
	jobID := uuid.MustParse("01900000-0000-7000-8000-000000000702")
	repo := &mockJobsRepo{due: []entity.SchedulerJob{{
		ID:                 jobID,
		JobType:            constant.SchedulerJobTypeCalendarSync,
		Status:             constant.SchedulerJobStatusPending,
		ConnectedAccountID: &accountID,
	}}}
	cal := &fakeJob{jobType: constant.SchedulerJobTypeCalendarSync}
	runner := scheduler.NewRunner(repo, testSchedulerLog(), []scheduler.Job{cal}, scheduler.Options{
		Now: func() time.Time { return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC) },
	})

	n, err := runner.ProcessDue(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || cal.runs != 1 || cal.scheduleN != 1 {
		t.Fatalf("n=%d runs=%d schedule=%d", n, cal.runs, cal.scheduleN)
	}
	if repo.finished[jobID] != constant.SchedulerJobStatusSucceeded {
		t.Fatalf("finished = %q", repo.finished[jobID])
	}
}

func TestRunnerRecordsFailureAndStillReschedules(t *testing.T) {
	accountID := uuid.MustParse("01900000-0000-7000-8000-000000000711")
	jobID := uuid.MustParse("01900000-0000-7000-8000-000000000712")
	repo := &mockJobsRepo{due: []entity.SchedulerJob{{
		ID:                 jobID,
		JobType:            constant.SchedulerJobTypeCalendarSync,
		Status:             constant.SchedulerJobStatusPending,
		ConnectedAccountID: &accountID,
	}}}
	cal := &fakeJob{jobType: constant.SchedulerJobTypeCalendarSync, err: errors.New("sync boom")}
	runner := scheduler.NewRunner(repo, testSchedulerLog(), []scheduler.Job{cal}, scheduler.Options{
		Now: func() time.Time { return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC) },
	})

	n, err := runner.ProcessDue(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || repo.finished[jobID] != constant.SchedulerJobStatusFailed || cal.scheduleN != 1 {
		t.Fatalf("n=%d finished=%q schedule=%d", n, repo.finished[jobID], cal.scheduleN)
	}
}

func TestValidateJobs(t *testing.T) {
	if err := scheduler.ValidateJobs(nil); err == nil {
		t.Fatal("expected error")
	}
	if err := scheduler.ValidateJobs([]scheduler.Job{
		&fakeJob{jobType: "a"},
		&fakeJob{jobType: "a"},
	}); err == nil {
		t.Fatal("expected duplicate error")
	}
}
