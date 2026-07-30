import { apiRequest } from "@/lib/api/client";

import type {
  CreateDonnaEventInput,
  CreateDonnaReminderInput,
  DonnaEventResponse,
  DonnaReminderResponse,
  TimelineResponse,
  UpdateDonnaEventInput,
  UpdateDonnaReminderInput,
} from "./Timeline.types";

export async function fetchTimeline(input: {
  from: string;
  to: string;
  signal?: AbortSignal;
}): Promise<TimelineResponse> {
  const query = new URLSearchParams({
    from: input.from,
    to: input.to,
  });
  return apiRequest<TimelineResponse>(`/api/v1/timeline?${query.toString()}`, {
    signal: input.signal,
  });
}

export async function createDonnaEvent(
  body: CreateDonnaEventInput,
): Promise<DonnaEventResponse> {
  return apiRequest<DonnaEventResponse>("/api/v1/donna/events", {
    method: "POST",
    body,
  });
}

export async function updateDonnaEvent(
  id: string,
  body: UpdateDonnaEventInput,
): Promise<DonnaEventResponse> {
  return apiRequest<DonnaEventResponse>(`/api/v1/donna/events/${id}`, {
    method: "PATCH",
    body,
  });
}

export async function deleteDonnaEvent(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/donna/events/${id}`, {
    method: "DELETE",
  });
}

export async function createDonnaReminder(
  body: CreateDonnaReminderInput,
): Promise<DonnaReminderResponse> {
  return apiRequest<DonnaReminderResponse>("/api/v1/donna/reminders", {
    method: "POST",
    body,
  });
}

export async function updateDonnaReminder(
  id: string,
  body: UpdateDonnaReminderInput,
): Promise<DonnaReminderResponse> {
  return apiRequest<DonnaReminderResponse>(`/api/v1/donna/reminders/${id}`, {
    method: "PATCH",
    body,
  });
}

export async function deleteDonnaReminder(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/donna/reminders/${id}`, {
    method: "DELETE",
  });
}
