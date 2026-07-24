# apps/web

Next.js App Router application for Donna.

## Run locally

```bash
cd apps/web
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Conventions

See `docs/CURSOR_RULES.md` and `.cursor/rules/donna-frontend.mdc`.

Feature modules live under `src/features/<Feature>/` with:

- `Feature.tsx` — UI only
- `Feature.logic.ts` — hooks and handlers
- `Feature.styles.ts` — Tailwind class constants
- `Feature.types.ts` — types
- `index.ts` — public exports

Reusable primitives live in `src/components/common/`.

## Current surface

- Landing page (`/`) — copper-bronze dark theme
