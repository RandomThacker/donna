package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Options configures the Logger Factory.
type Options struct {
	Service     string
	Environment string
	Level       string
	Output      io.Writer // optional; defaults to os.Stdout
}

// Factory creates module loggers with shared service/environment attributes.
// Packages must obtain loggers from Factory.Module — never via slog.New directly.
type Factory struct {
	root        *slog.Logger
	service     string
	environment string
}

// NewFactory builds the process-wide Logger Factory.
func NewFactory(opts Options) *Factory {
	service := opts.Service
	if service == "" {
		service = constant.ServiceAPI
	}
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	handlerOpts := &slog.HandlerOptions{Level: parseLevel(opts.Level)}
	var handler slog.Handler
	if strings.EqualFold(opts.Environment, constant.EnvDevelopment) {
		handler = slog.NewTextHandler(out, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(out, handlerOpts)
	}

	root := slog.New(handler).With(
		constant.LogAttrService, service,
		constant.LogAttrEnvironment, opts.Environment,
	)

	return &Factory{
		root:        root,
		service:     service,
		environment: opts.Environment,
	}
}

// Module returns a logger bound to a module name.
// Future global fields added to the factory root propagate automatically.
func (f *Factory) Module(name string) *Logger {
	if f == nil {
		panic("logger: Factory is nil; wire NewFactory at process start")
	}
	if name == "" {
		name = "unknown"
	}
	return &Logger{
		log: f.root.With(constant.LogAttrModule, name),
	}
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
