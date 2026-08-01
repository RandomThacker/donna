package entity

import (
	"time"

	"github.com/google/uuid"
)

// AutomationExecution is one run of an automation template.
type AutomationExecution struct {
	ID               uuid.UUID
	PublicID         string
	AutomationID     uuid.UUID
	UserID           uuid.UUID
	StartedAt        time.Time
	CompletedAt      *time.Time
	Status           string
	DurationMs       *int
	CommandsTotal    int
	CommandsSuccess  int
	CommandsFailed   int
	TriggerSource    string
	DeliveryChannels []string
	DeliveryStatus   *string
	Response         *string
	Error            *string
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Optional joined fields for API responses.
	AutomationName *string
	Commands       []AutomationCommandExecution
}

// AutomationCommandExecution is one command within an execution.
type AutomationCommandExecution struct {
	ID           uuid.UUID
	PublicID     string
	ExecutionID  uuid.UUID
	OrderIndex   int
	Command      string
	CommandType  *string
	StartedAt    time.Time
	CompletedAt  *time.Time
	Status       string
	DurationMs   *int
	Response     *string
	Error        *string
	CreatedAt    time.Time
}

// AutomationRunMetrics summarizes execution history for list cards.
type AutomationRunMetrics struct {
	AutomationID       uuid.UUID
	LastStatus         *string
	SuccessRate        *float64 // 0..1 over completed runs
	AverageDurationMs  *float64
	LastCommandsTotal  *int
	LastCommandsSuccess *int
	TotalExecutions    int
	SuccessfulRuns     int
}
