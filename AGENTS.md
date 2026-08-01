# Donna Agent Guide

Read before implementing:

1. [docs/PRD.md](docs/PRD.md) — product scope
2. [docs/SYSTEM_DESIGN.md](docs/SYSTEM_DESIGN.md) — system design
3. [docs/CURSOR_RULES.md](docs/CURSOR_RULES.md) — engineering standards
4. [docs/DONNA_PERSONALITY.md](docs/DONNA_PERSONALITY.md) — voice and assistant behavior
5. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries
6. [docs/ACTION_LAYER.md](docs/ACTION_LAYER.md) — Action Layer (Handler → Actions → Services)
7. [docs/COMMAND_CHAT.md](docs/COMMAND_CHAT.md) — Command Chat MVP (rule-based intents)
8. [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md) — logging, metrics, tracing, audit, AI usage
9. [docs/DOMAIN_MODEL.md](docs/DOMAIN_MODEL.md) — business entities, aggregates, ownership (no SQL)
10. [docs/DATA_MODEL.md](docs/DATA_MODEL.md) — logical data model: fields, constraints, indexes (no SQL)
11. [docs/SCHEMA_DECISIONS.md](docs/SCHEMA_DECISIONS.md) — schema ADRs + formal review (no SQL)
12. [docs/PHYSICAL_DATABASE_DESIGN.md](docs/PHYSICAL_DATABASE_DESIGN.md) — PostgreSQL physical design (no CREATE TABLE yet)
13. [docs/DATABASE.md](docs/DATABASE.md) — persistence standards (no tables yet)
14. [docs/PHASE1_PLAN.md](docs/PHASE1_PLAN.md) — milestone order
15. [docs/TIMELINE_UI.md](docs/TIMELINE_UI.md) — Timeline planning UI (Phase 3.1)
16. [docs/OCCURRENCE_DOMAIN.md](docs/OCCURRENCE_DOMAIN.md) — scheduling Occurrence model vs Timeline
17. [docs/OCCURRENCE_SERVICE.md](docs/OCCURRENCE_SERVICE.md) — OccurrenceService pipeline (scheduling feed)
18. [docs/PERFORMANCE_SPRINT_1A.md](docs/PERFORMANCE_SPRINT_1A.md) — narrow Occurrence SQL projections
19. [docs/PERFORMANCE_SPRINT_1B.md](docs/PERFORMANCE_SPRINT_1B.md) — shared calendar Occurrence query (query count)

Project rules also live in `.cursor/rules/`.

## Non-negotiables

- AI never writes to the database; API executes all mutations.
- Handler → Actions → Services → Repository in Go (models at the edge; Action DTOs in the domain; entities in services).
- Observability: module loggers from the Logger Factory only; follow [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md).
- Feature folders in web: UI / logic / styles / types separation.
- Pass the Donna Test: would a great human personal assistant behave this way?
- Phase 1: no Gmail, Drive, GitHub, Voice, Telegram, WhatsApp.
