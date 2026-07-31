"use client";

import { useEffect, useRef } from "react";

import { useAuth } from "@/features/auth/AuthProvider";

import { ensureWebPushSubscription } from "./webPush";

/**
 * After sign-in, ask for notification permission and register a Web Push
 * subscription with the API (when a service worker is available).
 */
export function WebPushRegister() {
  const { status } = useAuth();
  const attempted = useRef(false);

  useEffect(() => {
    if (status !== "authenticated" || attempted.current) {
      return;
    }
    attempted.current = true;
    void ensureWebPushSubscription().catch(() => {
      // Permission denied / no SW in local next dev — ignore.
    });
  }, [status]);

  return null;
}
