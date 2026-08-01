package business

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/occurrence"
)

// NotificationPolicy decides when an Occurrence should notify.
// Delivery remains a separate phase (dispatcher / channels).
type NotificationPolicy interface {
	ReminderTime(occ occurrence.Occurrence) time.Time
}

// googleEventNotificationPolicy reminds 10 minutes before a Google event.
type googleEventNotificationPolicy struct{}

func (googleEventNotificationPolicy) ReminderTime(occ occurrence.Occurrence) time.Time {
	return occ.StartAt.Add(-10 * time.Minute)
}

// microsoftICSEventNotificationPolicy reminds 10 minutes before an ICS/Microsoft event.
type microsoftICSEventNotificationPolicy struct{}

func (microsoftICSEventNotificationPolicy) ReminderTime(occ occurrence.Occurrence) time.Time {
	return occ.StartAt.Add(-10 * time.Minute)
}

// donnaEventNotificationPolicy reminds 15 minutes before a Donna event.
type donnaEventNotificationPolicy struct{}

func (donnaEventNotificationPolicy) ReminderTime(occ occurrence.Occurrence) time.Time {
	return occ.StartAt.Add(-15 * time.Minute)
}

// donnaReminderNotificationPolicy fires exactly at the reminder time.
type donnaReminderNotificationPolicy struct{}

func (donnaReminderNotificationPolicy) ReminderTime(occ occurrence.Occurrence) time.Time {
	return occ.StartAt
}

// NotificationPolicyResolver selects a NotificationPolicy for an Occurrence.
type NotificationPolicyResolver struct {
	google        NotificationPolicy
	microsoftICS  NotificationPolicy
	donnaEvent    NotificationPolicy
	donnaReminder NotificationPolicy
}

// NewNotificationPolicyResolver constructs the default Phase 2.2 policies.
func NewNotificationPolicyResolver() *NotificationPolicyResolver {
	return &NotificationPolicyResolver{
		google:        googleEventNotificationPolicy{},
		microsoftICS:  microsoftICSEventNotificationPolicy{},
		donnaEvent:    donnaEventNotificationPolicy{},
		donnaReminder: donnaReminderNotificationPolicy{},
	}
}

// Resolve returns the policy for an occurrence, or nil if none applies.
func (r *NotificationPolicyResolver) Resolve(occ occurrence.Occurrence) NotificationPolicy {
	switch occ.Source {
	case occurrence.SourceGoogle:
		if occ.Type == occurrence.TypeEvent {
			return r.google
		}
	case occurrence.SourceMicrosoftICS:
		if occ.Type == occurrence.TypeEvent {
			return r.microsoftICS
		}
	case occurrence.SourceDonna:
		switch occ.Type {
		case occurrence.TypeEvent:
			return r.donnaEvent
		case occurrence.TypeReminder:
			return r.donnaReminder
		}
	}
	return nil
}
