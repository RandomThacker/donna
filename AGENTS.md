# Donna Agent Guide

Read before implementing:

1. [docs/PRD.md](docs/PRD.md) — product scope
2. [docs/SYSTEM_DESIGN.md](docs/SYSTEM_DESIGN.md) — system design
3. [docs/CURSOR_RULES.md](docs/CURSOR_RULES.md) — engineering standards
4. [docs/DONNA_PERSONALITY.md](docs/DONNA_PERSONALITY.md) — voice and assistant behavior
5. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries
6. [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md) — logging, metrics, tracing, audit, AI usage
7. [docs/DOMAIN_MODEL.md](docs/DOMAIN_MODEL.md) — business entities, aggregates, ownership (no SQL)
8. [docs/DATA_MODEL.md](docs/DATA_MODEL.md) — logical data model: fields, constraints, indexes (no SQL)
9. [docs/DATABASE.md](docs/DATABASE.md) — persistence standards (no tables yet)
10. [docs/PHASE1_PLAN.md](docs/PHASE1_PLAN.md) — milestone order

Project rules also live in `.cursor/rules/`.

## Non-negotiables

- AI never writes to the database; API executes all mutations.
- Handler → Business → Repository in Go (models at the edge; entities in the domain).
- Observability: module loggers from the Logger Factory only; follow [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md).
- Feature folders in web: UI / logic / styles / types separation.
- Pass the Donna Test: would a great human personal assistant behave this way?
- Phase 1: no Gmail, Drive, GitHub, Voice, Telegram, WhatsApp.
