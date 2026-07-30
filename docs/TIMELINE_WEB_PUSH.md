# Timeline In-App Notification Delivery

Promotes due **PENDING** notifications to **SENT** for in-app surfaces.

Web Push is **disabled**. Delivery targets:

1. **Notification Center** — row becomes visible as unread (`SENT`)
2. **Chat** — `delivery_channels` includes `CHAT`; `channel_delivery_status.CHAT = SENT`

## Flow

1. `NotificationDispatcher` ticks every minute
2. Load `status = PENDING` AND `scheduled_for <= now`
3. Mark overall `status = SENT`, set `sent_at`
4. For each intended channel except `WEB_PUSH`, set channel status to `SENT`
5. Legacy rows that still list `WEB_PUSH` skip that channel without failing the notification

## Default channels

```text
["CHAT"]
```

## Out of scope

Browser Web Push, Telegram, WhatsApp, Email, retries.

Push subscription tables/handlers may remain in the codebase but are not wired.
