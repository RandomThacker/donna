# services/api

Go + Gin REST API for Donna (M1 foundation).

## Layering

```text
Handler → Business → Repository → Database
     ↓         ↓
   model    entity
```

| Layer | Package | Responsibility |
| --- | --- | --- |
| Handler | `internal/handler` | HTTP only — map request/response |
| Model | `internal/model` | Transport DTOs (JSON wire shape) |
| Business | `internal/business` | Domain rules / orchestration |
| Entity | `internal/entity` | Domain types |
| Repository | `internal/repository` | Persistence / SQL |
| Database | `internal/database` | Connection pool |

API prefix: `/api/v1`

## Layout

```text
cmd/api                 process entrypoint
configs/                JSON app config (env-interpolated)
  appconfig.json
  database.json
  api.json
internal/app            dependency wiring + lifecycle
internal/config         JSON + ${ENV} load + validation
internal/constant       routes, headers, codes, env keys
internal/handler        HTTP handlers
internal/model          request/response DTOs
internal/business       business layer
internal/entity         domain entities
internal/repository     data access
internal/database       pgxpool
internal/middleware     recovery, request ID, logging, CORS
internal/response       JSON envelope helpers
internal/router         Gin routes
internal/server         graceful HTTP server
migrations/             golang-migrate SQL (infra only in M1)
```

## Configuration

Runtime config is JSON under `configs/`. Secrets and environment-specific values use `${VAR}` or `${VAR:default}` and resolve from process env (and Compose `environment` / `.env`).

| File | Purpose |
| --- | --- |
| `appconfig.json` | addr, environment, log level, CORS, JWT, shutdown timeout |
| `database.json` | Postgres URL, pool sizes, timeouts, migrations path |
| `api.json` | External HTTP APIs (OpenAI, Google OAuth, AI service) — method, path, timeout, headers |

Set `CONFIG_DIR` to override the configs directory (Docker defaults to `/app/configs`).

### Common env vars

| Variable | Required | Notes |
| --- | --- | --- |
| `DATABASE_URL` | yes | Injected into `database.json` |
| `API_ADDR` | yes* | e.g. `:8080` (*or `PORT`) |
| `API_ENV` | no | defaults via JSON to `development` |
| `LOG_LEVEL` | no | defaults to `info` |
| `JWT_SECRET` | yes* | (*or `SESSION_SECRET`) |
| `CORS_ORIGINS` | no | comma-separated |
| `CONFIG_DIR` | no | defaults to `configs` |
| `OPENAI_API_KEY` | no | wired in `api.json` for later milestones |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | no | unused until M2 |

Copy repo-root `.env.example` to `.env` before `docker compose up`.

## Local run

```bash
# from repo root
cp -n .env.example .env

cd infra/docker
docker compose up --build
```

API listens on `http://localhost:8080`.

### Without Docker

Requires Go 1.25+, Postgres, and [golang-migrate](https://github.com/golang-migrate/migrate) CLI.

```bash
cd services/api
make migrate-up
make run
```

## Endpoints

```bash
curl -s http://localhost:8080/api/v1/health
curl -s http://localhost:8080/api/v1/ready
curl -s http://localhost:8080/api/v1/version
```

## Middleware

Order: Recovery → Request ID → Request logging → CORS.

## Observability

Follow [docs/OBSERVABILITY.md](../../docs/OBSERVABILITY.md).

```go
factory := logger.NewFactory(logger.Options{
  Service:     constant.ServiceAPI,
  Environment: cfg.App.Environment,
  Level:       cfg.App.LogLevel,
})
calendarLog := factory.Module(constant.ModuleCalendar)
calendarLog.Info(ctx, "sync started")
```

- Module loggers always include `module`, `service`, `environment`
- Request middleware sets `request_id` on context and response header
- HTTP requests log method, path, status, duration, IP, user agent (WARN if ≥ 500ms)
- Helpers: `AIUsage`, `AuthEvent`, `CalendarEvent`, `SchedulerEvent`, `DatabaseQuery`, `WorkerEvent`
- Never log secrets; use `logger.RedactMap` for header-like maps

## Migrations

Infra-only in M1: enables `pgcrypto`. No domain tables.

```bash
make migrate-up
make migrate-down
```

## Tests

```bash
cd services/api
make test
```
