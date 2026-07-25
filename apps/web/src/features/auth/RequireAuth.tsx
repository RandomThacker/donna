"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";

import { authStyles as styles } from "./Auth.styles";
import { useAuth } from "./AuthProvider";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/sign-in");
    }
  }, [router, status]);

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
