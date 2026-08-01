package entity

import (
	"encoding/json"
	"strings"
)

// AutomationCommand is a structured automation step (template command).
// Variables customize behavior without changing the UI label layer.
type AutomationCommand struct {
	Command   string            `json:"command" yaml:"command"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// UnmarshalAutomationCommands decodes JSON that may be string[] or structured objects.
func UnmarshalAutomationCommands(raw []byte) ([]AutomationCommand, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []AutomationCommand{}, nil
	}
	var elems []json.RawMessage
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, err
	}
	out := make([]AutomationCommand, 0, len(elems))
	for _, elem := range elems {
		cmd, err := unmarshalOneAutomationCommand(elem)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(cmd.Command) == "" {
			continue
		}
		out = append(out, cmd)
	}
	return out, nil
}

func unmarshalOneAutomationCommand(raw json.RawMessage) (AutomationCommand, error) {
	trimmed := strings.TrimSpace(string(raw))
	raw = json.RawMessage(trimmed)
	if len(raw) == 0 || string(raw) == "null" {
		return AutomationCommand{}, nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return AutomationCommand{}, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return AutomationCommand{}, nil
		}
		return AutomationCommand{
			Command:   "chat_message",
			Variables: map[string]string{"message": s},
		}, nil
	}
	var obj struct {
		Command   string            `json:"command"`
		Variables map[string]string `json:"variables"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return AutomationCommand{}, err
	}
	vars := cloneAutomationVars(obj.Variables)
	return AutomationCommand{
		Command:   strings.TrimSpace(obj.Command),
		Variables: vars,
	}, nil
}

func cloneAutomationVars(in map[string]string) map[string]string {
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
