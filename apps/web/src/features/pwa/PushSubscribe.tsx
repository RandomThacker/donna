"use client";

import { useEffect, useRef } from "react";

import { useAuth } from "@/features/auth";

import { fetchVapidPublicKey, subscribePush } from "./Push.api";
import {
  ensurePushSubscription,
  pushSubscriptionKeys,
} from "./Push.logic";

/**
 * After auth, request notification permission and sync the Web Push
 * subscription to the API. No-ops when unsupported or permission denied.
 */
export function PushSubscribe() {
  const { status } = useAuth();
  const ranForSession = useRef(false);

  useEffect(() => {
    if (status !== "authenticated") {
      ranForSession.current = false;
      return;
    }
    if (ranForSession.current) {
      return;
    }
    ranForSession.current = true;

    let cancelled = false;
    void (async () => {
      try {
        const publicKey = await fetchVapidPublicKey();
        if (cancelled) {
          return;
        }
        const result = await ensurePushSubscription(publicKey);
        if (cancelled || result.status !== "subscribed") {
          return;
        }
        const keys = pushSubscriptionKeys(result.subscription);
        if (!keys) {
          return;
        }
        await subscribePush(keys);
      } catch {
        // VAPID missing, SW disabled in dev, or network — skip quietly.
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [status]);

  return null;
}
