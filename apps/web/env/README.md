# Frontend env profiles

Switch which backend the local Next app talks to:

```bash
cd apps/web

npm run env:local   # local FE → local API (:8080)
npm run env:prod    # local FE → Railway (proxied via /api)

npm run dev         # start Next (uses current .env.local)
# or in one step:
npm run dev:local
npm run dev:prod
```

| Profile | File | API |
| --- | --- | --- |
| local | `env/local.env` | `http://localhost:8080` |
| prod | `env/prod.env` | Railway via `API_PROXY_TARGET` rewrite |

After switching, restart `npm run dev` if it was already running.

### Prod profile prerequisites

1. Railway `CORS_ORIGINS` includes `http://localhost:3000`
2. Google OAuth redirect URI includes `http://localhost:3000/api/v1/auth/google/callback`
3. API deployed with `return_to` support (so login returns to localhost)
