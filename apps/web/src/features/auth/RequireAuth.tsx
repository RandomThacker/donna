"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";

import {
  calendarQueryKeys,
  ensureCalendarSourcesFresh,
} from "@/features/calendar";

import { authStyles as styles } from "./Auth.styles";
import { useAuth } from "./AuthProvider";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/sign-in");
    }
  }, [router, status]);

  // App startup freshness: incremental sync when last success is older than 2 minutes.
  useEffect(() => {
    if (status !== "authenticated") {
      return;
    }
    const controller = new AbortController();
    void ensureCalendarSourcesFresh(controller.signal)
      .then(() =>
        queryClient.invalidateQueries({ queryKey: calendarQueryKeys.all }),
      )
      .catch(() => {
        // Soft-fail: local DB remains source of truth; background job / manual sync retry.
      });
    return () => controller.abort();
  }, [status, queryClient]);

  if (status === "loading") {
    return (
      <div className={styles.page}>
        <p className={styles.status}>Loading your workspace…</p>
      </div>
    );
  }

  if (status !== "authenticated") {
    return null;
  }

  return children;
}
