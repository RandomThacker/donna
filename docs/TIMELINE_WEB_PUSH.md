# Timeline Web Push Delivery (Phase 2.3)

Delivers **PENDING** notifications whose `scheduled_for` has arrived via browser Web Push.

## Flow

1. `NotificationDispatcher` ticks every minute
2. Load `status = PENDING` AND `scheduled_for <= now`
3. For each row with `WEB_PUSH` in `delivery_channels`:
   - Load the user's push subscriptions (all devices)
   - Send encrypted push payload to each endpoint
4. On ≥1 successful send → `status = SENT`, `channel_delivery_status.WEB_PUSH = SENT`
5. On zero successes (no subs / all failed / VAPID missing) → `status = FAILED`, `WEB_PUSH = FAILED`
6. Failed rows stay `FAILED` — **no retry** in this phase

## Subscription APIs

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/push/vapid-public-key` | Auth required; browser applicationServerKey |
| POST | `/api/v1/push/subscribe` | Upsert by `(user_id, endpoint)` |
| DELETE | `/api/v1/push/unsubscribe` | Soft-delete by endpoint |

Body for subscribe:

```json
{
  "endpoint": "https://…",
  "keys": { "p256dh": "…", "auth": "…" },
  "user_agent": "optional"
}
```

Multiple devices per user are supported. Gone endpoints (HTTP 404/410) are soft-deleted.

## Payload

```json
{
  "title": "…",
  "body": "…",
  "occurrenceId": "…",
  "timelineType": "EVENT",
  "source": "DONNA",
  "startTime": "…",
  "deepLink": "/donna/timeline?occurrence=…",
  "notificationId": "…"
}
```

## Frontend / PWA

- `PushSubscribe` (root layout) runs after auth: fetch VAPID → permission → SW subscribe → POST subscribe
- Service worker (`src/sw.ts`): `push` shows notification; `notificationclick` opens `deepLink`
- Serwist is **disabled in `next dev`** — use `next build && next start` (HTTPS or localhost) to exercise push

## Config

```bash
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=mailto:donna@localhost
```

Generate keys: `npx web-push generate-vapid-keys`

## Out of scope

Telegram, Chat, WhatsApp, Email, retries.
