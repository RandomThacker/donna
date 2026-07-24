# Donna Agent Guide

Read before implementing:

1. [docs/PRD.md](docs/PRD.md) — product scope
2. [docs/SYSTEM_DESIGN.md](docs/SYSTEM_DESIGN.md) — system design
3. [docs/CURSOR_RULES.md](docs/CURSOR_RULES.md) — engineering standards
4. [docs/DONNA_PERSONALITY.md](docs/DONNA_PERSONALITY.md) — voice and assistant behavior
5. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system boundaries
6. [docs/PHASE1_PLAN.md](docs/PHASE1_PLAN.md) — milestone order

Project rules also live in `.cursor/rules/`.

## Non-negotiables

- AI never writes to the database; API executes all mutations.
- Handler → Service → Repository in Go.
- Feature folders in web: UI / logic / styles / types separation.
- Pass the Donna Test: would a great human personal assistant behave this way?
- Phase 1: no Gmail, Drive, GitHub, Voice, Telegram, WhatsApp.
