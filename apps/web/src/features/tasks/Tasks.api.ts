import { apiRequest } from "@/lib/api/client";

import type {
  ReorderTaskOccurrencesInput,
  TaskDayResponse,
  TaskHistoryResponse,
  TaskOccurrence,
} from "./Tasks.types";

export async function fetchTaskDay(
  date: string,
  signal?: AbortSignal,
): Promise<TaskDayResponse> {
  return apiRequest<TaskDayResponse>(`/api/v1/tasks/day/${date}`, { signal });
}

export async function fetchTaskHistory(
  from: string,
  to: string,
  signal?: AbortSignal,
): Promise<TaskHistoryResponse> {
  const query = new URLSearchParams({ from, to });
  return apiRequest<TaskHistoryResponse>(
    `/api/v1/tasks/history?${query.toString()}`,
    { signal },
  );
}

export async function createTask(input: {
  title: string;
  date: string;
}): Promise<TaskOccurrence> {
  return apiRequest<TaskOccurrence>("/api/v1/tasks", {
    method: "POST",
    body: input,
  });
}

export async function updateTaskOccurrence(
  id: string,
  completed: boolean,
): Promise<TaskOccurrence> {
  return apiRequest<TaskOccurrence>(`/api/v1/task-occurrences/${id}`, {
    method: "PATCH",
    body: { completed },
  });
}

export async function reorderTaskOccurrences(
  input: ReorderTaskOccurrencesInput,
): Promise<void> {
  await apiRequest<unknown>("/api/v1/task-occurrences/reorder", {
    method: "PATCH",
    body: input,
  });
}

export async function deleteTask(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/tasks/${id}`, {
    method: "DELETE",
  });
}
