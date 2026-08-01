package personality

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed config/personalities/*.yaml
var personalityFS embed.FS

type yamlFile struct {
	ID                        string              `yaml:"id"`
	Name                      string              `yaml:"name"`
	Description               string              `yaml:"description"`
	EmojiLevelDefault         string              `yaml:"emoji_level_default"`
	HumorLevelDefault         string              `yaml:"humor_level_default"`
	EncouragementLevelDefault string              `yaml:"encouragement_level_default"`
	ResponseStyleDefault      string              `yaml:"response_style_default"`
	FallbackNicknames         []string            `yaml:"fallback_nicknames"`
	Punchlines                []string            `yaml:"punchlines"`
	Greetings                 map[string][]string `yaml:"greetings"`
	MorningGreetings          []string            `yaml:"morning_greetings"`
	EveningGreetings          []string            `yaml:"evening_greetings"`
	GoodNightGreetings        []string            `yaml:"goodnight_greetings"`
	Acknowledgements          []string            `yaml:"acknowledgements"`
	TaskComplete              []string            `yaml:"task_complete"`
	Errors                    []string            `yaml:"errors"`
	Reminders                 []string            `yaml:"reminders"`
	Notifications             []string            `yaml:"notifications"`
	AutomationIntros          []string            `yaml:"automation_intros"`
	Closings                  []string            `yaml:"closings"`
	Encouragements            []string            `yaml:"encouragements"`
	ChatWrappers              []string            `yaml:"chat_wrappers"`
	MorningBriefs             []string            `yaml:"morning_briefs"`
	Emojis                    map[string][]string `yaml:"emojis"`
}

type fileCatalog struct {
	byID  map[string]Definition
	order []Definition
}

var (
	catalogOnce sync.Once
	catalogInst *fileCatalog
	catalogErr  error
)

// LoadCatalog returns the embedded built-in personalities.
func LoadCatalog() (Catalog, error) {
	catalogOnce.Do(func() {
		entries, err := personalityFS.ReadDir("config/personalities")
		if err != nil {
			catalogErr = fmt.Errorf("personality catalog: %w", err)
			return
		}
		c := &fileCatalog{byID: map[string]Definition{}}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			raw, err := personalityFS.ReadFile("config/personalities/" + e.Name())
			if err != nil {
				catalogErr = fmt.Errorf("personality catalog read %s: %w", e.Name(), err)
				return
			}
			var file yamlFile
			if err := yaml.Unmarshal(raw, &file); err != nil {
				catalogErr = fmt.Errorf("personality catalog parse %s: %w", e.Name(), err)
				return
			}
			def, err := definitionFromYAML(file)
			if err != nil {
				catalogErr = err
				return
			}
			c.byID[def.ID] = def
			c.order = append(c.order, def)
		}
		if len(c.order) == 0 {
			catalogErr = fmt.Errorf("personality catalog: no personalities loaded")
			return
		}
		if _, ok := c.byID[DefaultID]; !ok {
			catalogErr = fmt.Errorf("personality catalog: missing default %q", DefaultID)
			return
		}
		catalogInst = c
	})
	return catalogInst, catalogErr
}

func definitionFromYAML(f yamlFile) (Definition, error) {
	id := strings.ToLower(strings.TrimSpace(f.ID))
	name := strings.TrimSpace(f.Name)
	if id == "" || name == "" {
		return Definition{}, fmt.Errorf("personality catalog: id and name are required")
	}
	return Definition{
		ID:                        id,
		Name:                      name,
		Description:               strings.TrimSpace(f.Description),
		EmojiLevelDefault:         normalizeLevel(f.EmojiLevelDefault, LevelNone),
		HumorLevelDefault:         normalizeLevel(f.HumorLevelDefault, LevelNone),
		EncouragementLevelDefault: normalizeLevel(f.EncouragementLevelDefault, LevelLow),
		ResponseStyleDefault:      strings.TrimSpace(f.ResponseStyleDefault),
		FallbackNicknames:         trimList(f.FallbackNicknames),
		Punchlines:                trimList(f.Punchlines),
		Greetings:                 f.Greetings,
		MorningGreetings:          trimList(f.MorningGreetings),
		EveningGreetings:          trimList(f.EveningGreetings),
		GoodNightGreetings:        trimList(f.GoodNightGreetings),
		Acknowledgements:          trimList(f.Acknowledgements),
		TaskComplete:              trimList(f.TaskComplete),
		Errors:                    trimList(f.Errors),
		Reminders:                 trimList(f.Reminders),
		Notifications:             trimList(f.Notifications),
		AutomationIntros:          f.AutomationIntros, // keep blanks
		Closings:                  f.Closings,
		Encouragements:            trimList(f.Encouragements),
		ChatWrappers:              trimList(f.ChatWrappers),
		MorningBriefs:             trimList(f.MorningBriefs),
		Emojis:                    f.Emojis,
	}, nil
}

func (c *fileCatalog) List() []Definition {
	if c == nil {
		return nil
	}
	out := make([]Definition, len(c.order))
	copy(out, c.order)
	return out
}

func (c *fileCatalog) Get(id string) (Definition, bool) {
	if c == nil {
		return Definition{}, false
	}
	def, ok := c.byID[strings.ToLower(strings.TrimSpace(id))]
	return def, ok
}

func normalizeLevel(raw string, fallback Level) Level {
	switch Level(strings.ToLower(strings.TrimSpace(raw))) {
	case LevelNone, LevelLow, LevelMedium, LevelHigh:
		return Level(strings.ToLower(strings.TrimSpace(raw)))
	default:
		return fallback
	}
}

func trimList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
