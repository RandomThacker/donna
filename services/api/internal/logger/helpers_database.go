package logger

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// DatabaseQuery logs a database operation at DEBUG, WARNing when slow or failed.
func (l *Logger) DatabaseQuery(ctx context.Context, op string, duration time.Duration, err error) {
	args := []any{
		constant.LogAttrQueryOp, op,
		constant.LogAttrDurationMS, duration.Milliseconds(),
	}
	if err != nil {
		args = append(args, constant.LogAttrError, err.Error())
		l.Error(ctx, "database query failed", args...)
		return
	}
	if duration >= constant.BudgetDBQuery {
		l.Warn(ctx, "database query exceeded budget", args...)
		return
	}
	l.Debug(ctx, "database query", args...)
}
