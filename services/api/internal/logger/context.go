package logger

import (
	"context"
)

// contextKey is an unexported type for context values.
type contextKey struct{}

// Fields holds optional correlation data for a request or job.
// Only non-empty fields are emitted to logs.
type Fields struct {
	RequestID         string
	UserID            string
	ConversationID    string
	SessionID         string
	CalendarSourceID  string
	ConnectionID      string
	JobID             string
	SchedulerID       string
	TraceID           string
}

// WithFields merges fields into ctx. Non-empty values in next overwrite existing ones.
func WithFields(ctx context.Context, next Fields) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	cur := FieldsFrom(ctx)
	merged := cur.merge(next)
	return context.WithValue(ctx, contextKey{}, merged)
}

// FieldsFrom returns correlation fields stored on ctx, or zero Fields.
func FieldsFrom(ctx context.Context) Fields {
	if ctx == nil {
		return Fields{}
	}
	if v, ok := ctx.Value(contextKey{}).(Fields); ok {
		return v
	}
	return Fields{}
}

func (f Fields) merge(next Fields) Fields {
	if next.RequestID != "" {
		f.RequestID = next.RequestID
	}
	if next.UserID != "" {
		f.UserID = next.UserID
	}
	if next.ConversationID != "" {
		f.ConversationID = next.ConversationID
	}
	if next.SessionID != "" {
		f.SessionID = next.SessionID
	}
	if next.CalendarSourceID != "" {
		f.CalendarSourceID = next.CalendarSourceID
	}
	if next.ConnectionID != "" {
		f.ConnectionID = next.ConnectionID
	}
	if next.JobID != "" {
		f.JobID = next.JobID
	}
	if next.SchedulerID != "" {
		f.SchedulerID = next.SchedulerID
	}
	if next.TraceID != "" {
		f.TraceID = next.TraceID
	}
	return f
}

// Attrs returns slog-compatible key/value pairs for non-empty fields.
func (f Fields) Attrs() []any {
	attrs := make([]any, 0, 18)
	if f.RequestID != "" {
		attrs = append(attrs, "request_id", f.RequestID)
	}
	if f.UserID != "" {
		attrs = append(attrs, "user_id", f.UserID)
	}
	if f.ConversationID != "" {
		attrs = append(attrs, "conversation_id", f.ConversationID)
	}
	if f.SessionID != "" {
		attrs = append(attrs, "session_id", f.SessionID)
	}
	if f.CalendarSourceID != "" {
		attrs = append(attrs, "calendar_source_id", f.CalendarSourceID)
	}
	if f.ConnectionID != "" {
		attrs = append(attrs, "connection_id", f.ConnectionID)
	}
	if f.JobID != "" {
		attrs = append(attrs, "job_id", f.JobID)
	}
	if f.SchedulerID != "" {
		attrs = append(attrs, "scheduler_id", f.SchedulerID)
	}
	if f.TraceID != "" {
		attrs = append(attrs, "trace_id", f.TraceID)
	}
	return attrs
}
