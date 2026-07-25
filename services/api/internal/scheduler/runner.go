package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
)

// Runner is the reusable platform poller for scheduler_jobs.
// It claims due rows and dispatches to registered Job handlers by job_type.
type Runner struct {
	repo     repository.SchedulerJobRepository
	handlers map[string]Job
	types    []string
	log      *logger.Logger
	interval time.Duration
	limit    int
	now      func() time.Time
}

// Options configures the platform runner.
type Options struct {
	PollInterval time.Duration
	BatchLimit   int
	Now          func() time.Time
}

// NewRunner wires handlers into a single poll loop.
func NewRunner(repo repository.SchedulerJobRepository, log *logger.Logger, jobs []Job, opts Options) *Runner {
	if opts.PollInterval <= 0 {
		opts.PollInterval = 30 * time.Second
	}
	if opts.BatchLimit <= 0 {
		opts.BatchLimit = 10
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	handlers := make(map[string]Job, len(jobs))
	types := make([]string, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		t := job.Type()
		handlers[t] = job
		types = append(types, t)
	}
	return &Runner{
		repo:     repo,
		handlers: handlers,
		types:    types,
		log:      log,
		interval: opts.PollInterval,
		limit:    opts.BatchLimit,
		now:      opts.Now,
	}
}

// Run blocks until ctx is canceled.
func (r *Runner) Run(ctx context.Context) {
	if r.repo == nil || len(r.handlers) == 0 {
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Runner) tick(ctx context.Context) {
	started := r.now()
	n, err := r.ProcessDue(ctx, r.limit)
	if err != nil && r.log != nil {
		r.log.Error(ctx, "scheduler poll failed", constant.LogAttrError, err)
		return
	}
	if r.log != nil && n > 0 {
		r.log.SchedulerJobCompleted(ctx, "scheduler_runner", r.now().Sub(started), "processed", n)
	}
}

// ProcessDue claims and dispatches up to limit due jobs across registered types.
func (r *Runner) ProcessDue(ctx context.Context, limit int) (int, error) {
	if r.repo == nil || len(r.types) == 0 {
		return 0, nil
	}
	now := r.now().UTC()
	due, err := r.repo.ListDue(ctx, r.types, now, limit)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, pending := range due {
		handler, ok := r.handlers[pending.JobType]
		if !ok {
			continue
		}
		claimed, claimErr := r.repo.Claim(ctx, pending.ID, now)
		if claimErr != nil {
			continue
		}
		processed++

		runErr := handler.Run(ctx, claimed)
		finished := r.now().UTC()
		if runErr != nil {
			msg := runErr.Error()
			_, _ = r.repo.Finish(ctx, claimed.ID, constant.SchedulerJobStatusFailed, &msg, finished)
			if r.log != nil {
				r.log.Warn(ctx, "scheduler job failed",
					"job_type", claimed.JobType,
					"job_id", claimed.ID.String(),
					constant.LogAttrError, runErr,
				)
			}
		} else {
			_, _ = r.repo.Finish(ctx, claimed.ID, constant.SchedulerJobStatusSucceeded, nil, finished)
		}

		if recurring, ok := handler.(RecurringJob); ok {
			if nextErr := recurring.ScheduleNext(ctx, claimed); nextErr != nil && r.log != nil {
				r.log.Warn(ctx, "scheduler failed to enqueue next run",
					"job_type", claimed.JobType,
					constant.LogAttrError, nextErr,
				)
			}
		}
	}
	return processed, nil
}

// RegisteredTypes returns job types the runner will poll for.
func (r *Runner) RegisteredTypes() []string {
	out := make([]string, len(r.types))
	copy(out, r.types)
	return out
}

// Ensure at least one job is registered at construction for clarity in wiring logs.
func ValidateJobs(jobs []Job) error {
	if len(jobs) == 0 {
		return fmt.Errorf("scheduler: at least one Job handler is required")
	}
	seen := map[string]struct{}{}
	for _, job := range jobs {
		if job == nil {
			return fmt.Errorf("scheduler: nil Job handler")
		}
		t := job.Type()
		if t == "" {
			return fmt.Errorf("scheduler: Job.Type must be non-empty")
		}
		if _, ok := seen[t]; ok {
			return fmt.Errorf("scheduler: duplicate Job.Type %q", t)
		}
		seen[t] = struct{}{}
	}
	return nil
}
