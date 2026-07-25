package logger

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Auth event names (stable for log queries).
const (
	AuthEventLogin               = "auth.login"
	AuthEventLogout              = "auth.logout"
	AuthEventRefresh             = "auth.refresh"
	AuthEventGoogleAccountLinked = "auth.google_account_linked"
)

// AuthEvent logs an authentication business event at INFO.
func (l *Logger) AuthEvent(ctx context.Context, event string, args ...any) {
	all := append([]any{constant.LogAttrEvent, event}, args...)
	l.Info(ctx, "auth event", all...)
}
