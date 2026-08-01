package model

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// AutomationCommandExecutionResponse is one command within an execution.
type AutomationCommandExecutionResponse struct {
	ID          string  `json:"id"`
	PublicID    string  `json:"public_id"`
	OrderIndex  int     `json:"order_index"`
	Command     string  `json:"command"`
	CommandType *string `json:"command_type,omitempty"`
	StartedAt   string  `json:"started_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
	Status      string  `json:"status"`
	DurationMs  *int    `json:"duration_ms,omitempty"`
	Response    *string `json:"response,omitempty"`
	Error       *string `json:"error,omitempty"`
}

// AutomationExecutionResponse is a recorded automation run.
type AutomationExecutionResponse struct {
	ID               string                               `json:"id"`
	PublicID         string                               `json:"public_id"`
	AutomationID     string                               `json:"automation_id"`
	AutomationName   *string                              `json:"automation_name,omitempty"`
	StartedAt        string                               `json:"started_at"`
	CompletedAt      *string                              `json:"completed_at,omitempty"`
	Status           string                               `json:"status"`
	DurationMs       *int                                 `json:"duration_ms,omitempty"`
	CommandsTotal    int                                  `json:"commands_total"`
	CommandsSuccess  int                                  `json:"commands_success"`
	CommandsFailed   int                                  `json:"commands_failed"`
	TriggerSource    string                               `json:"trigger_source"`
	Delivery         AutomationDeliveryResponse           `json:"delivery"`
	DeliveryStatus   *string                              `json:"delivery_status,omitempty"`
	Response         *string                              `json:"response,omitempty"`
	Error            *string                              `json:"error,omitempty"`
	Commands         []AutomationCommandExecutionResponse `json:"commands,omitempty"`
	CreatedAt        string                               `json:"created_at"`
	UpdatedAt        string                               `json:"updated_at"`
	Debug            map[string]any                       `json:"debug,omitempty"`
}

// AutomationAnalyticsResponse is aggregate execution analytics.
type AutomationAnalyticsResponse struct {
	TotalExecutions            int      `json:"total_executions"`
	SuccessRate                float64  `json:"success_rate"`
	FailureRate                float64  `json:"failure_rate"`
	AverageDurationMs          *float64 `json:"average_duration_ms,omitempty"`
	AverageCommandsPerRun      *float64 `json:"average_commands_per_run,omitempty"`
	MostFrequentAutomationID   *string  `json:"most_frequent_automation_id,omitempty"`
	MostFrequentAutomationName *string  `json:"most_frequent_automation_name,omitempty"`
}

// AutomationExecutionFromEntity maps an entity execution to the API response.
func AutomationExecutionFromEntity(e entity.AutomationExecution, includeDebug bool) AutomationExecutionResponse {
	out := AutomationExecutionResponse{
		ID:              e.ID.String(),
		PublicID:        e.PublicID,
		AutomationID:    e.AutomationID.String(),
		AutomationName:  e.AutomationName,
		StartedAt:       e.StartedAt.UTC().Format(time.RFC3339Nano),
		Status:          e.Status,
		DurationMs:      e.DurationMs,
		CommandsTotal:   e.CommandsTotal,
		CommandsSuccess: e.CommandsSuccess,
		CommandsFailed:  e.CommandsFailed,
		TriggerSource:   e.TriggerSource,
		Delivery: AutomationDeliveryResponse{
			Channels: e.DeliveryChannels,
		},
		DeliveryStatus: e.DeliveryStatus,
		Response:       e.Response,
		Error:          e.Error,
		CreatedAt:      e.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if e.CompletedAt != nil {
		s := e.CompletedAt.UTC().Format(time.RFC3339Nano)
		out.CompletedAt = &s
	}
	if len(e.Commands) > 0 {
		out.Commands = make([]AutomationCommandExecutionResponse, 0, len(e.Commands))
		for _, c := range e.Commands {
			cmd := AutomationCommandExecutionResponse{
				ID:          c.ID.String(),
				PublicID:    c.PublicID,
				OrderIndex:  c.OrderIndex,
				Command:     c.Command,
				CommandType: c.CommandType,
				StartedAt:   c.StartedAt.UTC().Format(time.RFC3339Nano),
				Status:      c.Status,
				DurationMs:  c.DurationMs,
				Response:    c.Response,
				Error:       c.Error,
			}
			if c.CompletedAt != nil {
				s := c.CompletedAt.UTC().Format(time.RFC3339Nano)
				cmd.CompletedAt = &s
			}
			out.Commands = append(out.Commands, cmd)
		}
	}
	if includeDebug {
		out.Debug = map[string]any{
			"raw_commands": out.Commands,
			"timing": map[string]any{
				"started_at":   out.StartedAt,
				"completed_at": out.CompletedAt,
				"duration_ms":  out.DurationMs,
			},
			"payload": map[string]any{
				"trigger_source":    e.TriggerSource,
				"delivery_channels": e.DeliveryChannels,
				"delivery_status":   e.DeliveryStatus,
				"status":            e.Status,
			},
		}
	}
	return out
}

// AutomationExecutionsFromEntities maps a slice of entities.
func AutomationExecutionsFromEntities(rows []entity.AutomationExecution, includeDebug bool) []AutomationExecutionResponse {
	out := make([]AutomationExecutionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, AutomationExecutionFromEntity(row, includeDebug))
	}
	return out
}
