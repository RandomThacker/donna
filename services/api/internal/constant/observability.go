package constant

import "time"

// Observability field keys (log attrs + context).
const (
	LogAttrModule               = "module"
	LogAttrService              = "service"
	LogAttrEnvironment          = "environment"
	LogAttrMessage              = "msg"
	LogAttrUserID               = "user_id"
	LogAttrConversationID       = "conversation_id"
	LogAttrSessionID            = "session_id"
	LogAttrCalendarSourceID     = "calendar_source_id"
	LogAttrConnectionID         = "connection_id"
	LogAttrJobID                = "job_id"
	LogAttrSchedulerID          = "scheduler_id"
	LogAttrTraceID              = "trace_id"
	LogAttrUserAgent            = "user_agent"
	LogAttrEvent                = "event"
	LogAttrModel                = "model"
	LogAttrProvider             = "provider"
	LogAttrInputTokens          = "input_tokens"
	LogAttrOutputTokens         = "output_tokens"
	LogAttrLatencyMS            = "latency_ms"
	LogAttrEstimatedCostUSD     = "estimated_cost_usd"
	LogAttrPromptVersion        = "prompt_version"
	LogAttrToolsUsed            = "tools_used"
	LogAttrMemoryRetrievalCount = "memory_retrieval_count"
	LogAttrQueryOp              = "query_op"
	LogAttrWorkerName           = "worker"
)

// Service identity for the API process.
const (
	ServiceAPI = "donna-api"
)

// Module names for the Logger Factory.
const (
	ModuleApp          = "app"
	ModuleHTTP         = "http"
	ModuleDatabase     = "database"
	ModuleAuth         = "auth"
	ModuleIdentity     = "identity"
	ModuleCalendar     = "calendar"
	ModuleTask         = "task"
	ModuleNote         = "note"
	ModuleChat         = "chat"
	ModuleDashboard    = "dashboard"
	ModuleScheduler    = "scheduler"
	ModuleAI           = "ai"
	ModuleMemory       = "memory"
	ModuleNotification = "notification"
	ModuleWorker       = "worker"
	ModuleHealth       = "health"
)

// Performance / slow-path thresholds.
const (
	SlowRequestThreshold = 500 * time.Millisecond
	BudgetAPILatency     = 200 * time.Millisecond
	BudgetDBQuery        = 50 * time.Millisecond
	BudgetSchedulerJob   = time.Second
	BudgetAIRequest      = 5 * time.Second
)

// HeaderUserAgent is the standard User-Agent request header.
const HeaderUserAgent = "User-Agent"
