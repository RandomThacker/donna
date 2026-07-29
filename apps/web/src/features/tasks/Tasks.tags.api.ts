import { apiRequest } from "@/lib/api/client";

import type { TaskTag, TaskTagsResponse } from "./Tasks.tags.types";

export async function fetchTaskTags(signal?: AbortSignal): Promise<TaskTag[]> {
  const data = await apiRequest<TaskTagsResponse>("/api/v1/task-tags", {
    signal,
  });
  return data.tags ?? [];
}

export async function createTaskTag(input: {
  name: string;
  color: string;
}): Promise<TaskTag> {
  return apiRequest<TaskTag>("/api/v1/task-tags", {
    method: "POST",
    body: input,
  });
}

export async function updateTaskTag(
  id: string,
  input: { name?: string; color?: string },
): Promise<TaskTag> {
  return apiRequest<TaskTag>(`/api/v1/task-tags/${id}`, {
    method: "PATCH",
    body: input,
  });
}

export async function deleteTaskTag(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/task-tags/${id}`, {
    method: "DELETE",
  });
}

export async function updateTaskTagAssignments(
  taskId: string,
  tagIds: string[],
): Promise<void> {
  await apiRequest<unknown>(`/api/v1/tasks/${taskId}`, {
    method: "PATCH",
    body: { tag_ids: tagIds },
  });
}
