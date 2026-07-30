package business

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// googleEventNotificationPolicy reminds 10 minutes before a Google event.
type googleEventNotificationPolicy struct{}

func (googleEventNotificationPolicy) ReminderTime(item entity.TimelineItem) time.Time {
	return item.StartAt.Add(-10 * time.Minute)
}

// microsoftICSEventNotificationPolicy reminds 10 minutes before an ICS/Microsoft event.
type microsoftICSEventNotificationPolicy struct{}

func (microsoftICSEventNotificationPolicy) ReminderTime(item entity.TimelineItem) time.Time {
	return item.StartAt.Add(-10 * time.Minute)
}

// donnaEventNotificationPolicy reminds 15 minutes before a Donna event.
type donnaEventNotificationPolicy struct{}

func (donnaEventNotificationPolicy) ReminderTime(item entity.TimelineItem) time.Time {
	return item.StartAt.Add(-15 * time.Minute)
}

// donnaReminderNotificationPolicy fires exactly at the reminder time.
type donnaReminderNotificationPolicy struct{}

func (donnaReminderNotificationPolicy) ReminderTime(item entity.TimelineItem) time.Time {
	return item.StartAt
}

// NotificationPolicyResolver selects a NotificationPolicy for a timeline item.
type NotificationPolicyResolver struct {
	google       NotificationPolicy
	microsoftICS NotificationPolicy
	donnaEvent   NotificationPolicy
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

// Resolve returns the policy for an item, or nil if none applies.
func (r *NotificationPolicyResolver) Resolve(item entity.TimelineItem) NotificationPolicy {
	switch item.Source {
	case constant.TimelineSourceGoogle:
		if item.Type == constant.TimelineTypeEvent {
			return r.google
		}
	case constant.TimelineSourceMicrosoftICS:
		if item.Type == constant.TimelineTypeEvent {
			return r.microsoftICS
		}
	case constant.TimelineSourceDonna:
		switch item.Type {
		case constant.TimelineTypeEvent:
			return r.donnaEvent
		case constant.TimelineTypeReminder:
			return r.donnaReminder
		}
	}
	return nil
}
