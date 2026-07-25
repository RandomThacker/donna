package logger

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// WorkerEvent logs a background worker lifecycle/business event at INFO.
func (l *Logger) WorkerEvent(ctx context.Context, workerName, event string, args ...any) {
	all := append([]any{
		constant.LogAttrWorkerName, workerName,
		constant.LogAttrEvent, event,
	}, args...)
	l.Info(ctx, "worker event", all...)
}

// WorkerError logs a background worker failure at ERROR.
func (l *Logger) WorkerError(ctx context.Context, workerName, event string, err error, args ...any) {
	all := append([]any{
		constant.LogAttrWorkerName, workerName,
		constant.LogAttrEvent, event,
		constant.LogAttrError, err.Error(),
	}, args...)
	l.Error(ctx, "worker error", all...)
}
