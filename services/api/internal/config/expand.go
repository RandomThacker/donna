package config

import (
	"os"
	"regexp"
)

// Matches ${VAR} or ${VAR:default}. Default may contain colons (e.g. URLs, :8080).
var envPlaceholder = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?\}`)

// expandEnv replaces ${VAR} and ${VAR:default} placeholders using process environment.
// Empty env values are treated as unset so defaults apply.
func expandEnv(input string) string {
	return envPlaceholder.ReplaceAllStringFunc(input, func(match string) string {
		parts := envPlaceholder.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		key := parts[1]
		def := ""
		if len(parts) > 2 {
			def = parts[2]
		}
		if val := os.Getenv(key); val != "" {
			return val
		}
		return def
	})
}
