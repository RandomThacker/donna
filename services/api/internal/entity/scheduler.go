package entity

import (
	"time"

	"github.com/google/uuid"
)

// SchedulerJob is a durable work item in the job ledger.
type SchedulerJob struct {
	ID                 uuid.UUID
	PublicID           string
	UserID             uuid.UUID
	JobType            string
	Status             string
	RunAt              time.Time
	AttemptCount       int
	MaxAttempts        int
	Payload            []byte
	ReminderID         *uuid.UUID
	ConnectedAccountID *uuid.UUID
	LastError          *string
	StartedAt          *time.Time
	FinishedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
