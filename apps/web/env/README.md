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
| local | `env/local.env` | `http://localhost:8080` (direct) |
| prod | `env/prod.env` | Railway via `src/app/api/[...path]` proxy |

After switching, **restart** `npm run dev` (env is read at startup).

### Prod profile prerequisites

1. Railway `CORS_ORIGINS` includes `http://localhost:3000`
2. Google OAuth redirect URI includes `http://localhost:3000/api/v1/auth/google/callback`
3. API deployed with `return_to` support

### Work laptop / corporate SSL

If `curl https://donna-assistant.up.railway.app/api/v1/ready` works but
`curl http://localhost:3000/api/v1/ready` returns 502 with a certificate error:

```bash
NODE_TLS_REJECT_UNAUTHORIZED=0 npm run dev
```

That only disables TLS verification for local Node. Prefer installing the corp root CA via `NODE_EXTRA_CA_CERTS`.
