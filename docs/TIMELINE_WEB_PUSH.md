# Timeline Notification Delivery (Chat + Web Push)

Promotes due **PENDING** notifications to **SENT** for in-app surfaces and browser push.

## Flow

1. `NotificationDispatcher` ticks every minute
2. Load `status = PENDING` AND `scheduled_for <= now`
3. **CHAT** (default): insert an assistant message (`client_message_id = notif:<public_id>`)
4. **WEB_PUSH** (when VAPID configured + device subscribed): fan-out via Web Push
5. Mark overall `status = SENT`, set `sent_at`
6. Recent already-SENT chat notifications are backfilled into chat if the message is missing

## Default channels

```text
["CHAT", "WEB_PUSH"]
```

## Web Push requirements

| Piece | Notes |
|-------|--------|
| `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` / `VAPID_SUBJECT` | Required on the API |
| `GET /api/v1/push/vapid-public-key` | Auth required |
| `POST /api/v1/push/subscribe` | Stores browser endpoint |
| Service worker `push` handler | `apps/web/src/sw.ts` (production Serwist build) |
| Installed PWA + notification permission | Required on iOS; recommended on Android |

Local `next dev` disables the service worker by default — use a production web build / installed PWA to receive OS pushes.

Push TTL is **24 hours** (`WebPushTTL`) with high urgency so briefly offline / Doze phones still receive calendar alerts. Automations deliver to **chat and Web Push** when `push` is in `delivery.channels` (default).
