# services/api

Go + Gin REST API for Donna.

## Layering

```text
Handler → Service → Repository → Database
```

- Handlers: HTTP only
- Services: business logic
- Repositories: SQL / persistence

API prefix: `/api/v1`

Scaffolded in **M1**.
