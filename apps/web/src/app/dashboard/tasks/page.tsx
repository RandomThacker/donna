"use client";

import { RequireAuth } from "@/features/auth";
import { Tasks } from "@/features/tasks";

export default function TasksPage() {
  return (
    <RequireAuth>
      <Tasks />
    </RequireAuth>
  );
}
