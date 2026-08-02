package business

import (
	"fmt"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
)

// ResolveAutomationCommand turns a structured command into a chat message + display label.
// The runner executes the resolved message through the existing chat executor.
func ResolveAutomationCommand(cmd entity.AutomationCommand) (message, label string, err error) {
	key := strings.ToLower(strings.TrimSpace(cmd.Command))
	vars := cmd.Variables
	if vars == nil {
		vars = map[string]string{}
	}

	switch key {
	case constant.AutomationCommandGreeting:
		return "Hi", "Greeting", nil
	case constant.AutomationCommandMorningGreeting:
		return "Good morning", "Morning greeting", nil
	case constant.AutomationCommandEveningGreeting:
		return "Evening greeting", "Evening greeting", nil
	case constant.AutomationCommandGoodNight:
		return "Good night", "Good night", nil
	case constant.AutomationCommandTodaysAgenda:
		rng := strings.ToLower(strings.TrimSpace(vars["range"]))
		if rng == "" {
			rng = "today"
		}
		switch rng {
		case "today":
			return "What do I have today?", "Today's Agenda", nil
		case "tomorrow":
			return "What do I have tomorrow?", "Tomorrow's Agenda", nil
		default:
			return "", "", fmt.Errorf("%w: todays_agenda range must be today or tomorrow", apperr.ErrValidation)
		}
	case constant.AutomationCommandTasksDue:
		// priority is reserved for future filtering; Phase 1.6 always uses today's open tasks.
		return "What's due today?", "Tasks Due", nil
	case constant.AutomationCommandTasksBacklog:
		return "Show backlog", "Task Backlog", nil
	case constant.AutomationCommandChatMessage:
		msg := strings.TrimSpace(vars["message"])
		if msg == "" {
			return "", "", fmt.Errorf("%w: chat_message requires variables.message", apperr.ErrValidation)
		}
		return msg, msg, nil
	case "":
		return "", "", fmt.Errorf("%w: command is required", apperr.ErrValidation)
	default:
		return "", "", fmt.Errorf("%w: unknown command %q", apperr.ErrValidation, key)
	}
}

// AutomationCommandLabel returns a UI-friendly label for a structured command.
func AutomationCommandLabel(cmd entity.AutomationCommand) string {
	_, label, err := ResolveAutomationCommand(cmd)
	if err != nil {
		if key := strings.TrimSpace(cmd.Command); key != "" {
			return key
		}
		return "Command"
	}
	return label
}

// NormalizeAutomationCommands validates and normalizes structured commands.
func NormalizeAutomationCommands(raw []entity.AutomationCommand) ([]entity.AutomationCommand, error) {
	out := make([]entity.AutomationCommand, 0, len(raw))
	for _, c := range raw {
		key := strings.ToLower(strings.TrimSpace(c.Command))
		if key == "" {
			continue
		}
		if _, ok := constant.AllowedAutomationCommands[key]; !ok {
			return nil, fmt.Errorf("%w: unknown command %q", apperr.ErrValidation, key)
		}
		vars := cloneCommandVars(c.Variables)
		normalized := entity.AutomationCommand{Command: key, Variables: vars}
		if _, _, err := ResolveAutomationCommand(normalized); err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}

// CommandsFromChatLines converts free-text lines into chat_message commands (custom UI).
func CommandsFromChatLines(lines []string) []entity.AutomationCommand {
	out := make([]entity.AutomationCommand, 0, len(lines))
	for _, line := range lines {
		msg := strings.TrimSpace(line)
		if msg == "" {
			continue
		}
		out = append(out, entity.AutomationCommand{
			Command:   constant.AutomationCommandChatMessage,
			Variables: map[string]string{"message": msg},
		})
	}
	return out
}

func cloneCommandVars(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		if key == "" || val == "" {
			continue
		}
		out[key] = val
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
