package constant

import "time"

// PublicIDPrefixAutomation is the atm_ prefix for automations.
const PublicIDPrefixAutomation = "atm_"

// PublicIDPrefixAutomationExecution is the aex_ prefix for automation runs.
const PublicIDPrefixAutomationExecution = "aex_"

// PublicIDPrefixAutomationCommandExecution is the ace_ prefix for per-command results.
const PublicIDPrefixAutomationCommandExecution = "ace_"

// Automation trigger types.
const (
	AutomationTriggerDaily  = "daily"
	AutomationTriggerWeekly = "weekly"
)

// Automation weekday codes (RRULE BYDAY).
var AutomationWeekdays = []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}

// AllowedAutomationTriggerTypes is the Phase 1.7 whitelist.
var AllowedAutomationTriggerTypes = map[string]struct{}{
	AutomationTriggerDaily:  {},
	AutomationTriggerWeekly: {},
}

// Automation delivery channels.
const (
	AutomationDeliveryChat     = "chat"
	AutomationDeliveryPush     = "push"
	AutomationDeliveryTelegram = "telegram" // future
	AutomationDeliveryWhatsApp = "whatsapp" // future
	AutomationDeliveryEmail    = "email"    // future
)

// AllowedAutomationDeliveryChannels is the whitelist for create/update.
var AllowedAutomationDeliveryChannels = map[string]struct{}{
	AutomationDeliveryChat: {},
	AutomationDeliveryPush: {},
}

// DefaultAutomationDeliveryChannels is applied when create omits delivery.
var DefaultAutomationDeliveryChannels = []string{
	AutomationDeliveryChat,
	AutomationDeliveryPush,
}

// Automation execution statuses.
const (
	AutomationExecutionRunning         = "RUNNING"
	AutomationExecutionSuccess         = "SUCCESS"
	AutomationExecutionPartialSuccess  = "PARTIAL_SUCCESS"
	AutomationExecutionFailed          = "FAILED"
	AutomationExecutionCancelled       = "CANCELLED"
)

// Automation command execution statuses.
const (
	AutomationCommandSuccess = "SUCCESS"
	AutomationCommandFailed  = "FAILED"
	AutomationCommandSkipped = "SKIPPED"
)

// Automation delivery statuses on an execution.
const (
	AutomationDeliveryPending = "PENDING"
	AutomationDeliverySent    = "SENT"
	AutomationDeliveryFailed  = "FAILED"
	AutomationDeliverySkipped = "SKIPPED"
)

// Automation trigger sources.
const (
	AutomationTriggerSourceScheduler = "scheduler"
	AutomationTriggerSourceManual    = "manual"
	AutomationTriggerSourcePreview   = "preview" // dry-run; never persisted
	AutomationTriggerSourceRetry     = "retry"   // future
	AutomationTriggerSourceReplay    = "replay"  // future
	AutomationTriggerSourceTelegram  = "telegram" // future
	AutomationTriggerSourceChat      = "chat"     // future
	AutomationTriggerSourceAPI       = "api"      // future
	AutomationTriggerSourceAI        = "ai"       // future
)

// Structured automation command keys.
const (
	AutomationCommandGreeting         = "greeting"
	AutomationCommandMorningGreeting  = "morning_greeting"
	AutomationCommandEveningGreeting  = "evening_greeting"
	AutomationCommandGoodNight        = "goodnight_greeting"
	AutomationCommandTodaysAgenda     = "todays_agenda"
	AutomationCommandTasksDue         = "tasks_due"
	AutomationCommandTasksBacklog     = "tasks_backlog"
	AutomationCommandChatMessage      = "chat_message"
)

// AllowedAutomationCommands is the Phase 1.6+ whitelist.
var AllowedAutomationCommands = map[string]struct{}{
	AutomationCommandGreeting:        {},
	AutomationCommandMorningGreeting: {},
	AutomationCommandEveningGreeting: {},
	AutomationCommandGoodNight:       {},
	AutomationCommandTodaysAgenda:    {},
	AutomationCommandTasksDue:        {},
	AutomationCommandTasksBacklog:    {},
	AutomationCommandChatMessage:     {},
}

// Automation scheduler timing.
const AutomationSchedulerInterval = time.Minute
