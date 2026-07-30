package constant

// Timeline source identifiers on TimelineItem.
const (
	TimelineSourceGoogle       = "GOOGLE"
	TimelineSourceMicrosoftICS = "MICROSOFT_ICS"
	TimelineSourceDonna        = "DONNA"
)

// Timeline item kinds.
const (
	TimelineTypeEvent    = "EVENT"
	TimelineTypeReminder = "REMINDER"
)

// Timeline item lifecycle status (cross-source).
const (
	TimelineStatusActive    = "ACTIVE"
	TimelineStatusCompleted = "COMPLETED"
	TimelineStatusCancelled = "CANCELLED"
	TimelineStatusMissed    = "MISSED"
)

// Donna-owned event / reminder status values (persistence).
const (
	DonnaEventStatusConfirmed = "confirmed"
	DonnaEventStatusCancelled = "cancelled"

	DonnaReminderStatusScheduled = "scheduled"
	DonnaReminderStatusCancelled = "cancelled"
)

// Public ID prefixes for Donna timeline domain tables.
const (
	PublicIDPrefixDonnaEvent    = "dev_"
	PublicIDPrefixDonnaReminder = "drm_"
)

// Logger module for timeline.
const ModuleTimeline = "timeline"
