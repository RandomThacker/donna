package logger_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
)

func TestFactoryInjectsModuleServiceEnvironment(t *testing.T) {
	var buf bytes.Buffer
	f := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: constant.EnvProduction,
		Level:       constant.LogLevelInfo,
		Output:      &buf,
	})
	mod := f.Module(constant.ModuleCalendar)
	mod.Info(context.Background(), "hello")

	out := buf.String()
	for _, want := range []string{
		`"msg":"hello"`,
		`"module":"calendar"`,
		`"service":"donna-api"`,
		`"environment":"production"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log %q missing %q", out, want)
		}
	}
}

func TestContextFieldsPropagate(t *testing.T) {
	var buf bytes.Buffer
	f := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: constant.EnvProduction,
		Level:       constant.LogLevelInfo,
		Output:      &buf,
	})
	mod := f.Module(constant.ModuleHTTP)
	ctx := logger.WithFields(context.Background(), logger.Fields{
		RequestID: "req-1",
		UserID:    "user-9",
	})
	mod.Info(ctx, "with context")

	out := buf.String()
	if !strings.Contains(out, `"request_id":"req-1"`) || !strings.Contains(out, `"user_id":"user-9"`) {
		t.Fatalf("missing context fields: %s", out)
	}
}

func TestAIUsageHelper(t *testing.T) {
	var buf bytes.Buffer
	f := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: constant.EnvProduction,
		Level:       constant.LogLevelInfo,
		Output:      &buf,
	})
	ai := f.Module(constant.ModuleAI)
	ai.AIUsage(context.Background(), logger.AIUsageEvent{
		Model:            "gpt-4o-mini",
		Provider:         "openai",
		InputTokens:      10,
		OutputTokens:     20,
		Latency:          100 * time.Millisecond,
		EstimatedCostUSD: 0.001,
		PromptVersion:    "v1",
	})
	out := buf.String()
	if !strings.Contains(out, `"event":"ai_usage"`) || !strings.Contains(out, `"model":"gpt-4o-mini"`) {
		t.Fatalf("unexpected ai log: %s", out)
	}
}

func TestRedactMap(t *testing.T) {
	got := logger.RedactMap(map[string]string{
		"Authorization": "Bearer secret",
		"X-Request-ID":  "abc",
	})
	if got["Authorization"] != "[REDACTED]" {
		t.Fatalf("Authorization = %q", got["Authorization"])
	}
	if got["X-Request-ID"] != "abc" {
		t.Fatalf("X-Request-ID = %q", got["X-Request-ID"])
	}
}

func TestDevelopmentUsesTextHandler(t *testing.T) {
	var buf bytes.Buffer
	f := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: constant.EnvDevelopment,
		Level:       constant.LogLevelInfo,
		Output:      &buf,
	})
	f.Module(constant.ModuleApp).Info(context.Background(), "dev line")
	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("expected text handler, got JSON: %s", out)
	}
	if !strings.Contains(out, "dev line") {
		t.Fatalf("missing message: %s", out)
	}
}
