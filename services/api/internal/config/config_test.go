package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/config"
	"github.com/RandomThacker/donna/services/api/internal/constant"
)

func writeConfigs(t *testing.T, dir string, app, database, api string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, constant.ConfigFileApp), []byte(app), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, constant.ConfigFileDatabase), []byte(database), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, constant.ConfigFileAPI), []byte(api), 0o644); err != nil {
		t.Fatal(err)
	}
}

func minimalAPIJSON() string {
	return `{
  "openai": {"name":"openai","base_url":"https://api.openai.com/v1","path":"/chat/completions","method":"POST","timeout":"30s","headers":{}},
  "google_oauth": {"name":"google_oauth","base_url":"https://oauth2.googleapis.com","path":"/token","method":"POST","timeout":"15s","headers":{}},
  "ai_service": {"name":"ai_service","base_url":"http://localhost:8090","path":"/health","method":"GET","timeout":"10s","headers":{}}
}`
}

func TestLoadRequiresCoreValues(t *testing.T) {
	dir := t.TempDir()
	writeConfigs(t, dir,
		`{"environment":"development","addr":"","log_level":"info","cors_origins":"","shutdown_timeout":"15s","jwt_secret":"","session_secret":""}`,
		`{"url":"","max_conns":"10","min_conns":"0","max_conn_lifetime":"1h","max_conn_idle_time":"30m","connect_ping_timeout":"5s","migrations_path":"migrations"}`,
		minimalAPIJSON(),
	)

	_, err := config.LoadFromDir(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	for _, want := range []string{"addr", "database.url", "jwt_secret"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}

func TestLoadExpandsEnvAndAliases(t *testing.T) {
	dir := t.TempDir()
	writeConfigs(t, dir,
		`{
  "environment":"${API_ENV:development}",
  "addr":"${API_ADDR}",
  "log_level":"${LOG_LEVEL:info}",
  "cors_origins":"${CORS_ORIGINS}",
  "shutdown_timeout":"15s",
  "jwt_secret":"${JWT_SECRET}",
  "session_secret":"${SESSION_SECRET}"
}`,
		`{
  "url":"${DATABASE_URL}",
  "max_conns":"${DB_MAX_CONNS:10}",
  "min_conns":"0",
  "max_conn_lifetime":"1h",
  "max_conn_idle_time":"30m",
  "connect_ping_timeout":"5s",
  "migrations_path":"migrations"
}`,
		minimalAPIJSON(),
	)

	t.Setenv(constant.EnvVarPort, "9090")
	t.Setenv(constant.EnvVarAPIAddr, "")
	t.Setenv(constant.EnvVarDatabaseURL, "postgres://donna:donna@localhost:5432/donna?sslmode=disable")
	t.Setenv(constant.EnvVarSessionSecret, "test-secret")
	t.Setenv(constant.EnvVarJWTSecret, "")
	t.Setenv(constant.EnvVarLogLevel, constant.LogLevelDebug)
	t.Setenv(constant.EnvVarAPIEnv, constant.EnvDevelopment)
	t.Setenv(constant.EnvVarCORSOrigins, "http://localhost:3000, http://127.0.0.1:3000")

	cfg, err := config.LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if cfg.App.Addr != ":9090" {
		t.Fatalf("Addr = %q, want :9090", cfg.App.Addr)
	}
	if cfg.App.JWTSecret != "test-secret" {
		t.Fatalf("JWTSecret = %q", cfg.App.JWTSecret)
	}
	if len(cfg.App.CORSOrigins) != 2 {
		t.Fatalf("CORSOrigins = %#v", cfg.App.CORSOrigins)
	}
	if cfg.App.LogLevel != constant.LogLevelDebug {
		t.Fatalf("LogLevel = %q", cfg.App.LogLevel)
	}
	if cfg.Database.MaxConns != 10 {
		t.Fatalf("MaxConns = %d", cfg.Database.MaxConns)
	}
	if cfg.API.OpenAI.Method != "POST" {
		t.Fatalf("OpenAI.Method = %q", cfg.API.OpenAI.Method)
	}
}

func TestLoadRejectsInvalidEnv(t *testing.T) {
	dir := t.TempDir()
	writeConfigs(t, dir,
		`{"environment":"nope","addr":":8080","log_level":"verbose","cors_origins":"","shutdown_timeout":"15s","jwt_secret":"secret","session_secret":""}`,
		`{"url":"postgres://x","max_conns":"10","min_conns":"0","max_conn_lifetime":"1h","max_conn_idle_time":"30m","connect_ping_timeout":"5s","migrations_path":"migrations"}`,
		minimalAPIJSON(),
	)

	_, err := config.LoadFromDir(dir)
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "environment") || !strings.Contains(msg, "log_level") {
		t.Fatalf("unexpected error: %v", err)
	}
}
