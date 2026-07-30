import { apiRequest } from "@/lib/api/client";

import type {
  DonnaNotification,
  NotificationsListResponse,
  NotificationStatus,
} from "./Notifications.types";

export async function listNotifications(input?: {
  statuses?: NotificationStatus[];
  signal?: AbortSignal;
}): Promise<DonnaNotification[]> {
  const query = new URLSearchParams();
  if (input?.statuses?.length) {
    query.set("status", input.statuses.join(","));
  }
  const suffix = query.toString() ? `?${query.toString()}` : "";
  const data = await apiRequest<NotificationsListResponse>(
    `/api/v1/notifications${suffix}`,
    { signal: input?.signal },
  );
  return data.notifications ?? [];
}

export async function markNotificationRead(
  id: string,
): Promise<DonnaNotification> {
  return apiRequest<DonnaNotification>(`/api/v1/notifications/${id}/read`, {
    method: "PATCH",
  });
}

export async function dismissNotification(
  id: string,
): Promise<DonnaNotification> {
  return apiRequest<DonnaNotification>(
    `/api/v1/notifications/${id}/dismiss`,
    { method: "PATCH" },
  );
}
