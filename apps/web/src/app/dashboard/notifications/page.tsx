"use client";

import { RequireAuth } from "@/features/auth";
import { Notifications } from "@/features/notifications/Notifications";

export default function NotificationsPage() {
  return (
    <RequireAuth>
      <Notifications />
    </RequireAuth>
  );
}
