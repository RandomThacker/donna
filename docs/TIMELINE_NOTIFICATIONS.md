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
- `delivery_channels` — intended channels (`CHAT` by default; Web Push disabled)
- `channel_delivery_status` — per-channel map, e.g. `{"CHAT":"PENDING"}` → `{"CHAT":"SENT"}`

## APIs

- `GET /notifications?status=PENDING,SENT`
- `PATCH /notifications/:id/read`
- `PATCH /notifications/:id/dismiss`

No delete — notifications are history.

## Out of scope

Telegram, WhatsApp, Web Push, Email, retries.

## Delivery

In-app only (Notification Center + Chat channel) — see [TIMELINE_WEB_PUSH.md](./TIMELINE_WEB_PUSH.md).

## Notification Center (web)

Inbox UI lives in `apps/web/src/features/notifications/`.

- Dedicated tab at `/dashboard/notifications` (sidebar + mobile bottom nav), same shell as Calendar / Todo
- Unread badge on the nav item (`SENT` count)
- Filters, search, day grouping, details pane, status timeline
- Reuses existing APIs only — no new notification endpoints
- Client-side pagination (50 + Load more); polls every 30s
- Developer Info (ids, payload, channel delivery) is visible in development builds only
- Opening a notification with an `occurrence_id` navigates to `/dashboard/calendar?event=…` and marks SENT → READ
