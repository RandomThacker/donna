"use client";

import { useEffect } from "react";

/**
 * Previous production builds register Serwist at /sw.js. That SW can keep
 * serving stale dashboard chunks during local `next dev` (old confirms,
 * old styles). Unregister + clear caches so localhost always runs fresh code.
 */
export function ClearDevServiceWorkers() {
  useEffect(() => {
    if (process.env.NODE_ENV !== "development") {
      return;
    }
    if (!("serviceWorker" in navigator)) {
      return;
    }

    void (async () => {
      const regs = await navigator.serviceWorker.getRegistrations();
      await Promise.all(regs.map((reg) => reg.unregister()));
      if ("caches" in window) {
        const keys = await caches.keys();
        await Promise.all(keys.map((key) => caches.delete(key)));
      }
    })();
  }, []);

  return null;
}
