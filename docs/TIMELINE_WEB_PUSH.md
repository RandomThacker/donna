# Timeline In-App Notification Delivery

Promotes due **PENDING** notifications to **SENT** for in-app surfaces.

Web Push is **disabled**. Delivery targets:

1. **Notification Center** — row becomes visible as unread (`SENT`)
2. **Chat** — dispatcher posts a Donna message into the primary web conversation, then sets `channel_delivery_status.CHAT = SENT`

## Flow

1. `NotificationDispatcher` ticks every minute
2. Load `status = PENDING` AND `scheduled_for <= now`
3. For `CHAT` (default): insert an assistant message (`client_message_id = notif:<public_id>`)
4. Mark overall `status = SENT`, set `sent_at`
5. Legacy rows that still list `WEB_PUSH` skip that channel without failing
6. Recent already-SENT chat notifications are backfilled into chat if the message is missing

## Default channels

```text
["CHAT"]
```

## Out of scope

Browser Web Push, Telegram, WhatsApp, Email, retries.

Push subscription tables/handlers may remain in the codebase but are not wired.
