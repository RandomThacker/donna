import { apiRequest } from "@/lib/api/client";

import type {
  CreateNoteInput,
  Note,
  NotesListResponse,
  UpdateNoteInput,
} from "./Notes.types";

export async function fetchNotes(signal?: AbortSignal): Promise<NotesListResponse> {
  return apiRequest<NotesListResponse>("/api/v1/notes", { signal });
}

export async function createNote(input: CreateNoteInput): Promise<Note> {
  return apiRequest<Note>("/api/v1/notes", {
    method: "POST",
    body: input,
  });
}

export async function updateNote(id: string, input: UpdateNoteInput): Promise<Note> {
  return apiRequest<Note>(`/api/v1/notes/${id}`, {
    method: "PATCH",
    body: input,
  });
}

export async function deleteNote(id: string): Promise<void> {
  await apiRequest<unknown>(`/api/v1/notes/${id}`, {
    method: "DELETE",
  });
}
