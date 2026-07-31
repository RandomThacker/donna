import { apiRequest } from "@/lib/api/client";

type VapidKeyResponse = {
  public_key: string;
};

type SubscribeBody = {
  endpoint: string;
  keys: {
    p256dh: string;
    auth: string;
  };
  user_agent?: string;
};

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = window.atob(base64);
  const output = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    output[i] = raw.charCodeAt(i);
  }
  return output;
}

async function fetchVapidPublicKey(): Promise<string> {
  const data = await apiRequest<VapidKeyResponse>("/api/v1/push/vapid-public-key");
  return data.public_key;
}

async function postSubscription(sub: PushSubscription): Promise<void> {
  const json = sub.toJSON();
  if (!json.endpoint || !json.keys?.p256dh || !json.keys?.auth) {
    throw new Error("incomplete push subscription");
  }
  const body: SubscribeBody = {
    endpoint: json.endpoint,
    keys: {
      p256dh: json.keys.p256dh,
      auth: json.keys.auth,
    },
    user_agent: typeof navigator !== "undefined" ? navigator.userAgent : undefined,
  };
  await apiRequest("/api/v1/push/subscribe", {
    method: "POST",
    body,
  });
}

/**
 * Request notification permission (if needed), subscribe the browser PushManager,
 * and register the endpoint with Donna. No-ops when unsupported or denied.
 */
export async function ensureWebPushSubscription(): Promise<"subscribed" | "skipped" | "unsupported"> {
  if (typeof window === "undefined") {
    return "unsupported";
  }
  if (!("Notification" in window) || !("serviceWorker" in navigator) || !("PushManager" in window)) {
    return "unsupported";
  }

  let permission = Notification.permission;
  if (permission === "default") {
    permission = await Notification.requestPermission();
  }
  if (permission !== "granted") {
    return "skipped";
  }

  const registration = await navigator.serviceWorker.ready;
  let subscription = await registration.pushManager.getSubscription();

  if (!subscription) {
    const publicKey = await fetchVapidPublicKey();
    subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey) as BufferSource,
    });
  }

  await postSubscription(subscription);
  return "subscribed";
}

/** Best-effort unsubscribe locally + on the server. */
export async function revokeWebPushSubscription(): Promise<void> {
  if (typeof window === "undefined" || !("serviceWorker" in navigator)) {
    return;
  }
  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.getSubscription();
  if (!subscription) {
    return;
  }
  const endpoint = subscription.endpoint;
  try {
    await apiRequest("/api/v1/push/unsubscribe", {
      method: "DELETE",
      body: { endpoint },
    });
  } catch {
    // Still drop the local subscription.
  }
  await subscription.unsubscribe();
}
