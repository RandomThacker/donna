"use client";

import { Suspense } from "react";

import { RequireAuth } from "@/features/auth";
import { Calendar } from "@/features/calendar";

export default function CalendarPage() {
  return (
    <RequireAuth>
      <Suspense fallback={null}>
        <Calendar />
      </Suspense>
    </RequireAuth>
  );
}
