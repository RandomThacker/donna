"use client";

import { RequireAuth } from "@/features/auth";
import { AutomationHistory } from "@/features/automations/AutomationHistory";

export default function AutomationHistoryPage() {
  return (
    <RequireAuth>
      <AutomationHistory />
    </RequireAuth>
  );
}
