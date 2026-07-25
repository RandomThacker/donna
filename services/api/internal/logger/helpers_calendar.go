package logger

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Calendar event names.
const (
	CalendarEventCreated = "calendar.event_created"
	CalendarEventUpdated = "calendar.event_updated"
	CalendarEventDeleted = "calendar.event_deleted"
	CalendarEventSync    = "calendar.sync"
)

// CalendarEvent logs a calendar business event at INFO.
func (l *Logger) CalendarEvent(ctx context.Context, event string, args ...any) {
	all := append([]any{constant.LogAttrEvent, event}, args...)
	l.Info(ctx, "calendar event", all...)
}
