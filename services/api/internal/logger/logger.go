package logger

import (
	"context"
	"log/slog"
)

// Logger is a module-scoped structured logger.
// Prefer context-aware methods so request/job fields attach automatically.
type Logger struct {
	log *slog.Logger
}

// Slog exposes the underlying slog.Logger for rare interop cases.
// Prefer Debug/Info/Warn/Error on Logger so context fields are applied.
func (l *Logger) Slog() *slog.Logger {
	if l == nil || l.log == nil {
		return slog.Default()
	}
	return l.log
}

// With returns a child logger with additional static attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{log: l.Slog().With(args...)}
}

// FromContext returns a slog.Logger enriched with Fields from ctx.
func (l *Logger) FromContext(ctx context.Context) *slog.Logger {
	attrs := FieldsFrom(ctx).Attrs()
	if len(attrs) == 0 {
		return l.Slog()
	}
	return l.Slog().With(attrs...)
}

// Debug logs at debug level with context fields.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.FromContext(ctx).Debug(msg, args...)
}

// Info logs at info level with context fields.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.FromContext(ctx).Info(msg, args...)
}

// Warn logs at warn level with context fields.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.FromContext(ctx).Warn(msg, args...)
}

// Error logs at error level with context fields.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.FromContext(ctx).Error(msg, args...)
}
