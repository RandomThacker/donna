package logger

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Scheduler event names.
const (
	SchedulerReminderScheduled = "scheduler.reminder_scheduled"
	SchedulerReminderSent      = "scheduler.reminder_sent"
	SchedulerReminderFailed    = "scheduler.reminder_failed"
)

// SchedulerEvent logs a scheduler business event at INFO (or ERROR when failed).
func (l *Logger) SchedulerEvent(ctx context.Context, event string, args ...any) {
	all := append([]any{constant.LogAttrEvent, event}, args...)
	if event == SchedulerReminderFailed {
		l.Error(ctx, "scheduler event", all...)
		return
	}
	l.Info(ctx, "scheduler event", all...)
}

// SchedulerJobCompleted logs job completion and WARNs when over the scheduler budget.
func (l *Logger) SchedulerJobCompleted(ctx context.Context, jobName string, duration time.Duration, args ...any) {
	all := append([]any{
		constant.LogAttrEvent, "scheduler.job_completed",
		constant.LogAttrJobID, jobName,
		constant.LogAttrDurationMS, duration.Milliseconds(),
	}, args...)
	if duration >= constant.BudgetSchedulerJob {
		l.Warn(ctx, "scheduler job exceeded budget", all...)
		return
	}
	l.Info(ctx, "scheduler job completed", all...)
}
