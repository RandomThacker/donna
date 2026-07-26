"use client";

import { RequireAuth } from "@/features/auth";
import { Calendar } from "@/features/calendar";

export default function CalendarPage() {
  return (
    <RequireAuth>
      <Calendar />
    </RequireAuth>
  );
}
