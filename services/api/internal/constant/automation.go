package constant

import "time"

// PublicIDPrefixAutomation is the atm_ prefix for automations.
const PublicIDPrefixAutomation = "atm_"

// PublicIDPrefixAutomationExecution is the aex_ prefix for automation runs.
const PublicIDPrefixAutomationExecution = "aex_"

// PublicIDPrefixAutomationCommandExecution is the ace_ prefix for per-command results.
const PublicIDPrefixAutomationCommandExecution = "ace_"

// Automation trigger types (Phase 1: daily only).
const (
	AutomationTriggerDaily = "daily"
	// Future: AutomationTriggerWeekly, AutomationTriggerMonthly, AutomationTriggerCron
)

// Automation delivery channels (Phase 1: chat only).
const (
	AutomationDeliveryChat     = "chat"
	AutomationDeliveryTelegram = "telegram" // future
	AutomationDeliveryPush     = "push"     // future
	AutomationDeliveryWhatsApp = "whatsapp" // future
	AutomationDeliveryEmail    = "email"    // future
)

// AllowedAutomationDeliveryChannels is the Phase 1 whitelist.
var AllowedAutomationDeliveryChannels = map[string]struct{}{
	AutomationDeliveryChat: {},
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
	AutomationCommandGreeting     = "greeting"
	AutomationCommandTodaysAgenda = "todays_agenda"
	AutomationCommandTasksDue     = "tasks_due"
	AutomationCommandChatMessage  = "chat_message"
)

// AllowedAutomationCommands is the Phase 1.6 whitelist.
var AllowedAutomationCommands = map[string]struct{}{
	AutomationCommandGreeting:     {},
	AutomationCommandTodaysAgenda: {},
	AutomationCommandTasksDue:     {},
	AutomationCommandChatMessage:  {},
}

// Automation scheduler timing.
const AutomationSchedulerInterval = time.Minute
