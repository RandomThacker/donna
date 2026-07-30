# Timeline Notification Queue (Phase 2.2)

Notifications are **derived from Timeline Items**. This phase only enqueues records — no delivery.

## Flow

1. `NotificationScheduler` ticks every minute
2. For each active user, fetch timeline for `[now, now+35m]` (20m lookahead + 15m max policy lead)
3. Resolve `NotificationPolicy` → `ReminderTime(item)`
4. If `ReminderTime ∈ [now, now+20m)`, insert `PENDING` notification

## Idempotency

Unique live index on `(occurrence_id, notification_type)`. Re-running the scheduler does not duplicate rows.

## Policies

| Source | Type | ReminderTime |
| --- | --- | --- |
| GOOGLE | EVENT | start − 10m |
| MICROSOFT_ICS | EVENT | start − 10m |
| DONNA | EVENT | start − 15m |
| DONNA | REMINDER | start (exact) |

## Status vs channel delivery

- `status` — overall lifecycle: `PENDING` → `SENT` / `READ` / `DISMISSED` / `FAILED`
- `delivery_channels` — intended channels (`WEB_PUSH`, `CHAT`, …)
- `channel_delivery_status` — per-channel map for the future delivery phase, e.g. `{"WEB_PUSH":"PENDING","CHAT":"PENDING"}`

## APIs

- `GET /notifications?status=PENDING,SENT`
- `PATCH /notifications/:id/read`
- `PATCH /notifications/:id/dismiss`

No delete — notifications are history.

## Out of scope

Telegram, WhatsApp, Chat delivery, AI, retries.

## Delivery

Web Push delivery is Phase 2.3 — see [TIMELINE_WEB_PUSH.md](./TIMELINE_WEB_PUSH.md).
