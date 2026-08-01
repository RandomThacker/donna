package entity

import (
	"time"

	"github.com/google/uuid"
)

// Automation is a user-owned scheduled run of one or more Donna chat commands.
type Automation struct {
	ID               uuid.UUID
	PublicID         string
	UserID           uuid.UUID
	Name             string
	Description      *string
	Enabled          bool
	TriggerType      string   // daily (future: weekly, monthly, cron)
	TriggerTime      string   // HH:MM civil clock in Timezone
	Timezone         string
	Commands         []AutomationCommand
	DeliveryChannels []string
	TemplateID       *string
	LastRunAt        *time.Time
	NextRunAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
