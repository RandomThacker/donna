# Donna

Personal AI Operating System — Phase 1.

Donna is not a chatbot. It is a proactive personal assistant with a dashboard homepage and a persistent phone-style chat.

## Stack

| Layer | Tech |
| --- | --- |
| Web | Next.js, TypeScript, Tailwind, shadcn/ui |
| API | Go, Gin, PostgreSQL |
| AI | Python, FastAPI |
| Infra | Docker Compose |

## Repository layout

```text
apps/web          Next.js application
services/api      Go Gin REST API
services/ai       FastAPI reasoning service
packages/shared   Shared pure helpers
packages/ui       Shared UI primitives
packages/types    Shared API types
docs/             PRD, architecture, plans
infra/docker      Local infrastructure
```

## Architecture rule

**AI reasons. Backend executes. Database stores.**

The AI service never writes to the database. Persistence always goes through the API.

## Docs

- [Product Requirements](docs/PRD.md)
- [System Design](docs/SYSTEM_DESIGN.md)
- [Engineering Standards](docs/CURSOR_RULES.md)
- [Personality Guide](docs/DONNA_PERSONALITY.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Phase 1 Plan](docs/PHASE1_PLAN.md)

## Local prerequisites

Install before M1+:

- Node.js 22+
- Go 1.25+
- Python 3.12+
- Docker

## Quick start

```bash
cp -n .env.example .env
cd infra/docker
docker compose up --build
# or: podman compose up --build
```

Postgres listens on `localhost:5432`. API listens on `localhost:8080`.

```bash
curl -s http://localhost:8080/api/v1/health
curl -s http://localhost:8080/api/v1/ready
curl -s http://localhost:8080/api/v1/version
```

See [services/api/README.md](services/api/README.md) for env vars and layering.

### Web (landing)

```bash
cd apps/web
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Current milestone

**M1 — Go API foundation** (Gin, health/ready/version, migrations, Handler → Service → Repository).

Next: **M2 — Auth** (Google OAuth, sessions).
