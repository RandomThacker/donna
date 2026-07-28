"use client";

import { RequireAuth } from "@/features/auth";
import { Settings } from "@/features/settings";

export default function SettingsPage() {
  return (
    <RequireAuth>
      <Settings />
    </RequireAuth>
  );
}
