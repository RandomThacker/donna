package logger

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Identity event names (stable for log queries).
const (
	IdentityEventUserCreated = "identity.user_created"
	IdentityEventUserUpdated = "identity.user_updated"
	IdentityEventUserDeleted = "identity.user_deleted"
)

// IdentityEvent logs an identity business event at INFO.
func (l *Logger) IdentityEvent(ctx context.Context, event string, args ...any) {
	all := append([]any{constant.LogAttrEvent, event}, args...)
	l.Info(ctx, "identity event", all...)
}
