package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// New builds a slog logger. JSON in non-development environments; text otherwise.
func New(environment, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	if strings.EqualFold(environment, constant.EnvDevelopment) {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case constant.LogLevelDebug:
		return slog.LevelDebug
	case constant.LogLevelWarn:
		return slog.LevelWarn
	case constant.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
