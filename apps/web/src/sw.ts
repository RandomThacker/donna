/// <reference lib="webworker" />

import { defaultCache } from "@serwist/next/worker";
import type { PrecacheEntry, SerwistGlobalConfig } from "serwist";
import { Serwist } from "serwist";

declare global {
  interface WorkerGlobalScope extends SerwistGlobalConfig {
    __SW_MANIFEST: (PrecacheEntry | string)[] | undefined;
  }
}

declare const self: ServiceWorkerGlobalScope;

const serwist = new Serwist({
  precacheEntries: self.__SW_MANIFEST,
  skipWaiting: true,
  clientsClaim: true,
  navigationPreload: true,
  runtimeCaching: defaultCache,
});

serwist.addEventListeners();

type PushPayload = {
  title?: string;
  body?: string;
  deepLink?: string;
  occurrenceId?: string;
  notificationId?: string;
};

function isMobileUserAgent(): boolean {
  return /Android|iPhone|iPad|iPod|Mobile/i.test(self.navigator.userAgent);
}

/** Desktop → home; mobile → chat. Event deep-links live in the chat "View Event" line. */
function landingPathForNotification(): string {
  return isMobileUserAgent() ? "/dashboard/chat" : "/dashboard";
}

self.addEventListener("push", (event) => {
  event.waitUntil(
    (async () => {
      let payload: PushPayload = {};
      try {
        if (event.data) {
          payload = event.data.json() as PushPayload;
        }
      } catch {
        try {
          const text = event.data?.text();
          if (text) {
            payload = { body: text };
          }
        } catch {
          // Ignore malformed payloads.
        }
      }

      const title = payload.title?.trim() || "Donna";
      const body = payload.body?.trim() || "You have a new notification";
      const url = landingPathForNotification();

      const options: NotificationOptions & { renotify?: boolean } = {
        body,
        icon: "/icons/icon-192.png",
        badge: "/icons/icon-192.png",
        tag: payload.notificationId || "donna-notification",
        renotify: true,
        data: {
          url,
          occurrenceId: payload.occurrenceId,
        },
      };
      await self.registration.showNotification(title, options);
    })(),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const raw = event.notification.data as { url?: string } | undefined;
  const target =
    typeof raw?.url === "string" && raw.url.startsWith("/")
      ? raw.url
      : landingPathForNotification();

  event.waitUntil(
    (async () => {
      const all = await self.clients.matchAll({
        type: "window",
        includeUncontrolled: true,
      });
      for (const client of all) {
        if ("focus" in client) {
          await client.focus();
          if ("navigate" in client) {
            await (client as WindowClient).navigate(target);
          }
          return;
        }
      }
      await self.clients.openWindow(target);
    })(),
  );
});
