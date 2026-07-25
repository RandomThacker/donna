package constant

// Gin context keys.
const (
	ContextKeyRequestID = "request_id"
)

// Structured log attribute keys.
const (
	LogAttrRequestID  = "request_id"
	LogAttrMethod     = "method"
	LogAttrPath       = "path"
	LogAttrStatus     = "status"
	LogAttrDurationMS = "duration_ms"
	LogAttrClientIP   = "client_ip"
	LogAttrError      = "error"
	LogAttrSignal     = "signal"
	LogAttrAddr       = "addr"
)
