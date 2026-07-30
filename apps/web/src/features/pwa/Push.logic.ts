/** Converts a URL-safe base64 VAPID key to a Uint8Array for PushManager. */
export function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = atob(base64);
  const output = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    output[i] = raw.charCodeAt(i);
  }
  return output;
}

export type PushPermissionOutcome = "granted" | "denied" | "default";

/** Maps Notification.permission / requestPermission result for tests and callers. */
export function mapNotificationPermission(
  permission: NotificationPermission,
): PushPermissionOutcome {
  if (permission === "granted") {
    return "granted";
  }
  if (permission === "denied") {
    return "denied";
  }
  return "default";
}

/**
 * Registers the service worker (if needed), requests notification permission,
 * and returns a browser PushSubscription when granted.
 */
export async function ensurePushSubscription(
  vapidPublicKey: string,
): Promise<
  | { status: "subscribed"; subscription: PushSubscription }
  | { status: "permission_denied" }
  | { status: "unsupported" }
> {
  if (
    typeof window === "undefined" ||
    !("serviceWorker" in navigator) ||
    !("PushManager" in window) ||
    !("Notification" in window)
  ) {
    return { status: "unsupported" };
  }

  // Serwist is disabled in `next dev` — avoid hanging on serviceWorker.ready.
  let registration = await navigator.serviceWorker.getRegistration();
  if (!registration) {
    try {
      registration = await navigator.serviceWorker.register("/sw.js");
    } catch {
      return { status: "unsupported" };
    }
  }

  const permission =
    Notification.permission === "granted"
      ? "granted"
      : await Notification.requestPermission();

  if (mapNotificationPermission(permission) !== "granted") {
    return { status: "permission_denied" };
  }

  const existing = await registration.pushManager.getSubscription();
  if (existing) {
    return { status: "subscribed", subscription: existing };
  }

  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(vapidPublicKey) as BufferSource,
  });
  return { status: "subscribed", subscription };
}

export function pushSubscriptionKeys(subscription: PushSubscription): {
  endpoint: string;
  p256dh: string;
  auth: string;
} | null {
  const json = subscription.toJSON();
  const p256dh = json.keys?.p256dh;
  const auth = json.keys?.auth;
  if (!json.endpoint || !p256dh || !auth) {
    return null;
  }
  return { endpoint: json.endpoint, p256dh, auth };
}
