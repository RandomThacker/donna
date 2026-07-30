package constant

import "time"

// Push subscription public id prefix.
const PublicIDPrefixPushSubscription = "psub_"

// Web Push dispatcher cadence.
const NotificationDispatcherInterval = time.Minute

// Env / config keys for VAPID.
const (
	EnvVarVAPIDPublicKey  = "VAPID_PUBLIC_KEY"
	EnvVarVAPIDPrivateKey = "VAPID_PRIVATE_KEY"
	EnvVarVAPIDSubject    = "VAPID_SUBJECT"
)
