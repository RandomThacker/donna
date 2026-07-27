package constant

// Public ID prefixes for task journal entities.
const (
	PublicIDPrefixTask           = "tsk_"
	PublicIDPrefixTaskOccurrence = "toc_"
	PublicIDPrefixDailyNote      = "dnt_"
)

// Task occurrence sources.
const (
	TaskOccurrenceSourceManual       = "manual"
	TaskOccurrenceSourceRecurring    = "recurring"
	TaskOccurrenceSourceCalendar     = "calendar"
	TaskOccurrenceSourceAI           = "ai"
	TaskOccurrenceSourceCarryForward = "carry_forward"
)

// Task status values (tasks.status).
const (
	TaskStatusOpen      = "open"
	TaskStatusCompleted = "completed"
	TaskStatusCancelled = "cancelled"
)
