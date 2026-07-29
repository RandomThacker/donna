"use client";

import { RequireAuth } from "@/features/auth";
import { Memories } from "@/features/memories";

export default function MemoriesPage() {
  return (
    <RequireAuth>
      <Memories />
    </RequireAuth>
  );
}
