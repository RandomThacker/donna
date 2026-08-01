package automationcatalog

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/RandomThacker/donna/services/api/internal/entity"
	"gopkg.in/yaml.v3"
)

//go:embed templates.yaml
var templatesYAML []byte

// Template is a predefined automation recipe (configuration, not business logic).
type Template struct {
	ID              string                     `yaml:"id" json:"id"`
	Name            string                     `yaml:"name" json:"name"`
	Description     string                     `yaml:"description" json:"description"`
	Commands        []entity.AutomationCommand `yaml:"commands" json:"commands"`
	DefaultSchedule DefaultSchedule            `yaml:"default_schedule" json:"default_schedule"`
}

// DefaultSchedule is the suggested daily trigger for a template.
type DefaultSchedule struct {
	Type string `yaml:"type" json:"type"`
	Time string `yaml:"time" json:"time"`
}

type catalogFile struct {
	Templates []Template `yaml:"templates"`
}

var (
	loadOnce sync.Once
	loaded   []Template
	loadErr  error
)

// Load returns Intent Catalog templates embedded at build time.
func Load() ([]Template, error) {
	loadOnce.Do(func() {
		var file catalogFile
		if err := yaml.Unmarshal(templatesYAML, &file); err != nil {
			loadErr = fmt.Errorf("automation catalog: %w", err)
			return
		}
		out := make([]Template, 0, len(file.Templates))
		for _, t := range file.Templates {
			id := strings.TrimSpace(t.ID)
			name := strings.TrimSpace(t.Name)
			if id == "" || name == "" {
				loadErr = fmt.Errorf("automation catalog: template missing id or name")
				return
			}
			t.ID = id
			t.Name = name
			t.Description = strings.TrimSpace(t.Description)
			cmds := make([]entity.AutomationCommand, 0, len(t.Commands))
			for _, c := range t.Commands {
				key := strings.ToLower(strings.TrimSpace(c.Command))
				if key == "" {
					continue
				}
				vars := map[string]string{}
				for k, v := range c.Variables {
					kk := strings.TrimSpace(k)
					vv := strings.TrimSpace(v)
					if kk == "" || vv == "" {
						continue
					}
					vars[kk] = vv
				}
				if len(vars) == 0 {
					vars = nil
				}
				cmds = append(cmds, entity.AutomationCommand{Command: key, Variables: vars})
			}
			t.Commands = cmds
			if strings.TrimSpace(t.DefaultSchedule.Type) == "" {
				t.DefaultSchedule.Type = "daily"
			}
			if strings.TrimSpace(t.DefaultSchedule.Time) == "" {
				t.DefaultSchedule.Time = "09:00"
			}
			out = append(out, t)
		}
		loaded = out
	})
	return loaded, loadErr
}

// Get returns a template by id.
func Get(id string) (Template, bool) {
	templates, err := Load()
	if err != nil {
		return Template{}, false
	}
	want := strings.TrimSpace(id)
	for _, t := range templates {
		if t.ID == want {
			return t, true
		}
	}
	return Template{}, false
}
