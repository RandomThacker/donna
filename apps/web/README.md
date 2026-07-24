# apps/web

Next.js App Router application for Donna.

## Conventions

See `docs/CURSOR_RULES.md` and `.cursor/rules/donna-frontend.mdc`.

Feature modules live under `src/features/<Feature>/` with:

- `Feature.tsx` — UI only
- `Feature.logic.ts` — hooks and handlers
- `Feature.styles.ts` — Tailwind class constants
- `Feature.types.ts` — types
- `index.ts` — public exports

Scaffolded in **M4**.
