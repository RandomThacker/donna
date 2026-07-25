package logger

import "strings"

// sensitiveKeySubstrings are matched case-insensitively against map keys.
var sensitiveKeySubstrings = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"authorization",
	"api_key",
	"apikey",
	"refresh",
	"private_key",
	"client_secret",
	"jwt",
	"cookie",
	"session",
}

const redacted = "[REDACTED]"

// RedactMap returns a shallow copy of m with sensitive keys redacted.
// Safe to use when logging header maps or loosely typed metadata.
func RedactMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = redacted
			continue
		}
		out[k] = v
	}
	return out
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, s := range sensitiveKeySubstrings {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
