package constant

import "time"

// Notification lifecycle (overall queue status).
const (
	NotificationStatusPending   = "PENDING"
	NotificationStatusSent      = "SENT"
	NotificationStatusRead      = "READ"
	NotificationStatusDismissed = "DISMISSED"
	NotificationStatusFailed    = "FAILED"
)

// Notification kinds derived from timeline items.
const (
	NotificationTypeEvent    = "EVENT"
	NotificationTypeReminder = "REMINDER"
)

// Intended delivery channels.
const (
	DeliveryChannelWebPush  = "WEB_PUSH"
	DeliveryChannelChat     = "CHAT"
	DeliveryChannelTelegram = "TELEGRAM"
	DeliveryChannelWhatsApp = "WHATSAPP"
)

// Per-channel delivery states (stored in channel_delivery_status jsonb).
const (
	ChannelDeliveryPending = "PENDING"
	ChannelDeliverySent    = "SENT"
	ChannelDeliveryFailed  = "FAILED"
)

// PublicIDPrefixNotification is the ntf_ prefix (existing table).
const PublicIDPrefixNotification = "ntf_"

// Notification scheduler timing.
const (
	NotificationSchedulerInterval = time.Minute
	NotificationLookaheadWindow   = 20 * time.Minute
	NotificationMaxPolicyLead     = 15 * time.Minute // Donna event offset
)

// Default intended channels until user preferences exist.
var DefaultDeliveryChannels = []string{
	DeliveryChannelChat,
	DeliveryChannelWebPush,
}

// NotificationDeepLinkPath is the default in-app path for Web Push clicks (desktop).
// Mobile clients open /dashboard?phone=1 so the Home phone sheet opens (service worker).
const NotificationDeepLinkPath = "/dashboard"

// NotificationChatLandingPath is the mobile push landing (Home + phone popup).
const NotificationChatLandingPath = "/dashboard?phone=1"

// NotificationCalendarEventPath prefixes occurrence ids for "View Event" chat links.
const NotificationCalendarEventPath = "/dashboard/calendar?event="

// NotificationDispatcherBatchLimit caps due notifications processed per tick.
const NotificationDispatcherBatchLimit = 100

// WebPushTTL is how long the push service may retain a message if the device
// is offline (seconds). 60s was too short for mobile Doze / brief offline —
// FCM accepted the send then dropped it before the phone woke up.
const WebPushTTL = 86400
