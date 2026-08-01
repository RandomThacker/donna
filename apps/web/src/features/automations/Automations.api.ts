import { apiRequest } from "@/lib/api/client";

import type {
  Automation,
  AutomationAnalytics,
  AutomationExecution,
  AutomationHistoryResponse,
  AutomationRunResult,
  AutomationTemplatesResponse,
  AutomationsListResponse,
  CreateAutomationInput,
  UpdateAutomationInput,
} from "./Automations.types";

export async function fetchAutomations(
  signal?: AbortSignal,
): Promise<AutomationsListResponse> {
  return apiRequest<AutomationsListResponse>("/api/v1/automations", { signal });
}

export async function fetchAutomationTemplates(
  signal?: AbortSignal,
): Promise<AutomationTemplatesResponse> {
  return apiRequest<AutomationTemplatesResponse>("/api/v1/automations/templates", {
    signal,
  });
}

export async function createAutomation(
  input: CreateAutomationInput,
): Promise<Automation> {
  return apiRequest<Automation>("/api/v1/automations", {
    method: "POST",
    body: input,
  });
}

export async function updateAutomation(
  id: string,
  input: UpdateAutomationInput,
): Promise<Automation> {
  return apiRequest<Automation>(`/api/v1/automations/${id}`, {
    method: "PATCH",
    body: input,
  });
}

export async function deleteAutomation(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/automations/${id}`, {
    method: "DELETE",
  });
}

export async function fetchAutomationHistoryAll(
  signal?: AbortSignal,
  limit = 50,
): Promise<AutomationHistoryResponse> {
  return apiRequest<AutomationHistoryResponse>(
    `/api/v1/automations/history?limit=${limit}`,
    { signal },
  );
}

export async function fetchAutomationHistory(
  automationId: string,
  signal?: AbortSignal,
  limit = 50,
): Promise<AutomationHistoryResponse> {
  return apiRequest<AutomationHistoryResponse>(
    `/api/v1/automations/${automationId}/history?limit=${limit}`,
    { signal },
  );
}

export async function fetchAutomationExecution(
  executionId: string,
  signal?: AbortSignal,
): Promise<AutomationExecution> {
  return apiRequest<AutomationExecution>(
    `/api/v1/automations/executions/${executionId}`,
    { signal },
  );
}

export async function fetchAutomationAnalytics(
  signal?: AbortSignal,
): Promise<AutomationAnalytics> {
  return apiRequest<AutomationAnalytics>("/api/v1/automations/analytics", {
    signal,
  });
}

export async function runAutomation(id: string): Promise<AutomationRunResult> {
  return apiRequest<AutomationRunResult>(`/api/v1/automations/${id}/run`, {
    method: "POST",
    body: {},
  });
}

export async function previewAutomation(
  id: string,
): Promise<AutomationRunResult> {
  return apiRequest<AutomationRunResult>(`/api/v1/automations/${id}/preview`, {
    method: "POST",
    body: {},
  });
}
