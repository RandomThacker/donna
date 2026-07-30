"use client";

import { RequireAuth } from "@/features/auth";
import { Commands } from "@/features/commands";

export default function CommandsPage() {
  return (
    <RequireAuth>
      <Commands />
    </RequireAuth>
  );
}
