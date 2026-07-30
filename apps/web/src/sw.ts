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
};

self.addEventListener("push", (event: PushEvent) => {
  let payload: PushPayload = {};
  try {
    if (event.data) {
      payload = event.data.json() as PushPayload;
    }
  } catch {
    payload = { title: "Donna", body: event.data?.text() || "You have a notification" };
  }

  const title = payload.title?.trim() || "Donna";
  const body = payload.body?.trim() || "You have a notification";
  const deepLink =
    payload.deepLink?.trim() ||
    (payload.occurrenceId
      ? `/donna/timeline?occurrence=${encodeURIComponent(payload.occurrenceId)}`
      : "/dashboard");

  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      icon: "/icons/icon-192.png",
      badge: "/icons/icon-192.png",
      data: { deepLink },
      tag: payload.occurrenceId || "donna-notification",
    }),
  );
});

self.addEventListener("notificationclick", (event: NotificationEvent) => {
  event.notification.close();

  const raw =
    (event.notification.data as { deepLink?: string } | undefined)?.deepLink ||
    "/dashboard";
  const path = raw.startsWith("/") ? raw : `/${raw}`;
  const targetUrl = new URL(path, self.location.origin).href;

  event.waitUntil(
    (async () => {
      const clientsList = await self.clients.matchAll({
        type: "window",
        includeUncontrolled: true,
      });
      for (const client of clientsList) {
        if ("focus" in client) {
          await client.focus();
          if ("navigate" in client) {
            await (client as WindowClient).navigate(targetUrl);
          }
          return;
        }
      }
      await self.clients.openWindow(targetUrl);
    })(),
  );
});
