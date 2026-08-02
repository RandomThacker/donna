package model

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/automationcatalog"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// AutomationTriggerRequest is the trigger shape on create/update.
type AutomationTriggerRequest struct {
	Type string   `json:"type"`
	Time string   `json:"time"`
	Days []string `json:"days,omitempty"`
}

// AutomationDeliveryRequest is the delivery shape on create/update.
type AutomationDeliveryRequest struct {
	Channels []string `json:"channels"`
}

// AutomationCommandRequest accepts either a string (legacy chat line) or a structured object.
type AutomationCommandRequest struct {
	Command   string            `json:"command"`
	Variables map[string]string `json:"variables,omitempty"`
}

// UnmarshalJSON supports "What do I have today?" or {"command":"todays_agenda",...}.
func (c *AutomationCommandRequest) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		*c = AutomationCommandRequest{
			Command:   constant.AutomationCommandChatMessage,
			Variables: map[string]string{"message": s},
		}
		return nil
	}
	var raw struct {
		Command   string            `json:"command"`
		Variables map[string]string `json:"variables"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*c = AutomationCommandRequest{
		Command:   strings.TrimSpace(raw.Command),
		Variables: raw.Variables,
	}
	return nil
}

// ToEntity maps a request command to the domain shape.
func (c AutomationCommandRequest) ToEntity() entity.AutomationCommand {
	return entity.AutomationCommand{
		Command:   c.Command,
		Variables: c.Variables,
	}
}

// AutomationCommandsToEntities maps request commands.
func AutomationCommandsToEntities(cmds []AutomationCommandRequest) []entity.AutomationCommand {
	out := make([]entity.AutomationCommand, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, c.ToEntity())
	}
	return out
}

// CreateAutomationRequest is POST /automations.
type CreateAutomationRequest struct {
	Name        string                     `json:"name"`
	Description *string                    `json:"description"`
	Enabled     *bool                      `json:"enabled"`
	Trigger     *AutomationTriggerRequest  `json:"trigger"`
	Timezone    string                     `json:"timezone"`
	Commands    []AutomationCommandRequest `json:"commands"`
	Delivery    *AutomationDeliveryRequest `json:"delivery"`
	TemplateID  *string                    `json:"template_id"`
}

// UpdateAutomationRequest is PATCH /automations/:id.
type UpdateAutomationRequest struct {
	Name        *string                    `json:"name"`
	Description *string                    `json:"description"`
	Enabled     *bool                      `json:"enabled"`
	Trigger     *AutomationTriggerRequest  `json:"trigger"`
	Timezone    *string                    `json:"timezone"`
	Commands    []AutomationCommandRequest `json:"commands"`
	Delivery    *AutomationDeliveryRequest `json:"delivery"`
}

// AutomationTriggerResponse is the trigger shape in API responses.
type AutomationTriggerResponse struct {
	Type string   `json:"type"`
	Time string   `json:"time"`
	Days []string `json:"days,omitempty"`
}

// AutomationDeliveryResponse is the delivery shape in API responses.
type AutomationDeliveryResponse struct {
	Channels []string `json:"channels"`
}

// AutomationCommandResponse is a structured command with a display label.
type AutomationCommandResponse struct {
	Command   string            `json:"command"`
	Variables map[string]string `json:"variables,omitempty"`
	Label     string            `json:"label"`
}

// AutomationResponse is the API shape for an automation.
type AutomationResponse struct {
	ID                  string                      `json:"id"`
	PublicID            string                      `json:"public_id"`
	Name                string                      `json:"name"`
	Description         *string                     `json:"description,omitempty"`
	Enabled             bool                        `json:"enabled"`
	Trigger             AutomationTriggerResponse   `json:"trigger"`
	Timezone            string                      `json:"timezone"`
	Commands            []AutomationCommandResponse `json:"commands"`
	Delivery            AutomationDeliveryResponse  `json:"delivery"`
	TemplateID          *string                     `json:"template_id,omitempty"`
	LastRunAt           *string                     `json:"last_run_at,omitempty"`
	NextRunAt           *string                     `json:"next_run_at,omitempty"`
	LastStatus          *string                     `json:"last_status,omitempty"`
	SuccessRate         *float64                    `json:"success_rate,omitempty"`
	AverageDurationMs   *float64                    `json:"average_duration_ms,omitempty"`
	LastCommandsTotal   *int                        `json:"last_commands_total,omitempty"`
	LastCommandsSuccess *int                        `json:"last_commands_success,omitempty"`
	TotalExecutions     int                         `json:"total_executions"`
	CreatedAt           string                      `json:"created_at"`
	UpdatedAt           string                      `json:"updated_at"`
}

// AutomationTemplateResponse is a catalog template.
type AutomationTemplateResponse struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	Commands        []AutomationCommandResponse `json:"commands"`
	DefaultSchedule AutomationTriggerResponse   `json:"default_schedule"`
}

// AutomationRunCommandResponse is one command result from run/preview.
type AutomationRunCommandResponse struct {
	OrderIndex  int    `json:"order_index"`
	Command     string `json:"command"`
	CommandKey  string `json:"command_key,omitempty"`
	CommandType string `json:"command_type,omitempty"`
	Status      string `json:"status"`
	DurationMs  int    `json:"duration_ms"`
	Response    string `json:"response,omitempty"`
	Error       string `json:"error,omitempty"`
}

// AutomationRunResponse is POST /automations/:id/run or /preview.
type AutomationRunResponse struct {
	Response        string                         `json:"response"`
	Status          string                         `json:"status"`
	DeliveryStatus  string                         `json:"delivery_status"`
	CommandsTotal   int                            `json:"commands_total"`
	CommandsSuccess int                            `json:"commands_success"`
	CommandsFailed  int                            `json:"commands_failed"`
	DurationMs      int                            `json:"duration_ms"`
	TriggerSource   string                         `json:"trigger_source"`
	Commands        []AutomationRunCommandResponse `json:"commands"`
	ExecutionID     *string                        `json:"execution_id,omitempty"`
	DryRun          bool                           `json:"dry_run"`
}

// AutomationCommandLabel is a fallback display label without importing business.
func AutomationCommandLabel(c entity.AutomationCommand) string {
	key := strings.ToLower(strings.TrimSpace(c.Command))
	vars := c.Variables
	switch key {
	case constant.AutomationCommandGreeting:
		return "Greeting"
	case constant.AutomationCommandMorningGreeting:
		return "Morning greeting"
	case constant.AutomationCommandEveningGreeting:
		return "Evening greeting"
	case constant.AutomationCommandGoodNight:
		return "Good night"
	case constant.AutomationCommandTodaysAgenda:
		if strings.EqualFold(strings.TrimSpace(vars["range"]), "tomorrow") {
			return "Tomorrow's Agenda"
		}
		return "Today's Agenda"
	case constant.AutomationCommandTasksDue:
		return "Tasks Due"
	case constant.AutomationCommandTasksBacklog:
		return "Task Backlog"
	case constant.AutomationCommandChatMessage:
		if msg := strings.TrimSpace(vars["message"]); msg != "" {
			return msg
		}
		return "Chat message"
	default:
		if key != "" {
			return key
		}
		return "Command"
	}
}

// AutomationCommandFromEntity maps a structured command for API responses.
func AutomationCommandFromEntity(c entity.AutomationCommand) AutomationCommandResponse {
	return AutomationCommandResponse{
		Command:   c.Command,
		Variables: c.Variables,
		Label:     AutomationCommandLabel(c),
	}
}

// AutomationCommandsFromEntities maps structured commands.
func AutomationCommandsFromEntities(cmds []entity.AutomationCommand) []AutomationCommandResponse {
	out := make([]AutomationCommandResponse, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, AutomationCommandFromEntity(c))
	}
	return out
}

// AutomationFromEntity maps an entity to the API response.
func AutomationFromEntity(a entity.Automation) AutomationResponse {
	out := AutomationResponse{
		ID:          a.ID.String(),
		PublicID:    a.PublicID,
		Name:        a.Name,
		Description: a.Description,
		Enabled:     a.Enabled,
		Trigger: AutomationTriggerResponse{
			Type: a.TriggerType,
			Time: a.TriggerTime,
			Days: a.TriggerDays,
		},
		Timezone: a.Timezone,
		Commands: AutomationCommandsFromEntities(a.Commands),
		Delivery: AutomationDeliveryResponse{
			Channels: a.DeliveryChannels,
		},
		TemplateID: a.TemplateID,
		CreatedAt:  a.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if a.LastRunAt != nil {
		s := a.LastRunAt.UTC().Format(time.RFC3339Nano)
		out.LastRunAt = &s
	}
	if a.NextRunAt != nil {
		s := a.NextRunAt.UTC().Format(time.RFC3339Nano)
		out.NextRunAt = &s
	}
	return out
}

// AutomationsFromEntities maps a slice of entities.
func AutomationsFromEntities(autos []entity.Automation) []AutomationResponse {
	out := make([]AutomationResponse, 0, len(autos))
	for _, a := range autos {
		out = append(out, AutomationFromEntity(a))
	}
	return out
}

// AutomationTemplateFromCatalog maps a catalog template to the API response.
func AutomationTemplateFromCatalog(t automationcatalog.Template) AutomationTemplateResponse {
	return AutomationTemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Commands:    AutomationCommandsFromEntities(t.Commands),
		DefaultSchedule: AutomationTriggerResponse{
			Type: t.DefaultSchedule.Type,
			Time: t.DefaultSchedule.Time,
		},
	}
}

// AutomationTemplatesFromCatalog maps catalog templates.
func AutomationTemplatesFromCatalog(templates []automationcatalog.Template) []AutomationTemplateResponse {
	out := make([]AutomationTemplateResponse, 0, len(templates))
	for _, t := range templates {
		out = append(out, AutomationTemplateFromCatalog(t))
	}
	return out
}
