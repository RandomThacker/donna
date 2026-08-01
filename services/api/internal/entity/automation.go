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
	TriggerType      string   // daily | weekly
	TriggerTime      string   // HH:MM civil clock in Timezone
	TriggerDays      []string // RRULE weekday codes for weekly (MO..SU); empty for daily
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
