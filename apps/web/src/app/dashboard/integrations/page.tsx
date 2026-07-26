"use client";

import { RequireAuth } from "@/features/auth";
import { IntegrationsPage } from "@/features/settings";

export default function IntegrationsRoute() {
  return (
    <RequireAuth>
      <IntegrationsPage />
    </RequireAuth>
  );
}
