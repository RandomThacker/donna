import { Suspense } from "react";

import { AuthCallback } from "@/features/auth";

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-dvh items-center justify-center bg-donna-bg text-donna-muted">
          Completing sign-in…
        </div>
      }
    >
      <AuthCallback />
    </Suspense>
  );
}
