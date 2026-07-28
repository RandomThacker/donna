import { apiRequest } from "@/lib/api/client";
import {
  getGoogleIntegrationStartUrl,
  getMicrosoftIntegrationStartUrl,
} from "@/lib/api/config";

import { listCalendarSources } from "@/features/calendar/Calendar.api";

import type { ConnectedAccount, ICSIntegration } from "./Integrations.types";

export async function listConnectedAccounts(): Promise<ConnectedAccount[]> {
  const data = await apiRequest<ConnectedAccount[]>("/api/v1/integrations");
  return Array.isArray(data) ? data : [];
}

export async function disconnectConnectedAccount(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/integrations/${id}`, {
    method: "DELETE",
  });
}

export async function listICSIntegrations(): Promise<ICSIntegration[]> {
  const data = await apiRequest<ICSIntegration[]>("/api/v1/integrations/ics");
  return Array.isArray(data) ? data : [];
}

export async function createICSIntegration(input: {
  name: string;
  ics_url: string;
  sync_enabled: boolean;
}): Promise<ICSIntegration> {
  return apiRequest<ICSIntegration>("/api/v1/integrations/ics", {
    method: "POST",
    body: input,
  });
}

export async function updateICSIntegration(
  id: string,
  input: { name?: string; sync_enabled?: boolean },
): Promise<ICSIntegration> {
  return apiRequest<ICSIntegration>(`/api/v1/integrations/ics/${id}`, {
    method: "PATCH",
    body: input,
  });
}

export async function deleteICSIntegration(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/integrations/ics/${id}`, {
    method: "DELETE",
  });
}

export async function syncICSIntegration(id: string): Promise<ICSIntegration> {
  return apiRequest<ICSIntegration>(`/api/v1/integrations/ics/${id}/sync`, {
    method: "POST",
  });
}

export function startGoogleIntegration(): void {
  window.location.assign(getGoogleIntegrationStartUrl());
}

export function startMicrosoftIntegration(): void {
  window.location.assign(getMicrosoftIntegrationStartUrl());
}

export { listCalendarSources };
