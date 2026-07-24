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
- Go 1.22+
- Python 3.12+
- Docker

## Quick start (M0)

```bash
cd infra/docker
docker compose up -d
```

Postgres listens on `localhost:5432` (see `.env.example`).

### Web (landing)

```bash
cd apps/web
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Current milestone

**M0 — Monorepo foundation** (scaffold, docs, Cursor rules, Postgres compose).

Next: **M1 — Go API skeleton** (`/api/v1/health`, migrations, layering).
