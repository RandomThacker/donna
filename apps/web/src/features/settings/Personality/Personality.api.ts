import { apiRequest } from "@/lib/api/client";

import type {
  PersonalityCatalogResponse,
  PersonalityPreview,
  PersonalityPreviewInput,
  PersonalityProfile,
  UpdatePersonalityInput,
} from "./Personality.types";

export async function fetchPersonalityProfile(
  signal?: AbortSignal,
): Promise<PersonalityProfile> {
  return apiRequest<PersonalityProfile>("/api/v1/settings/personality", {
    signal,
  });
}

export async function fetchPersonalityCatalog(
  signal?: AbortSignal,
): Promise<PersonalityCatalogResponse> {
  return apiRequest<PersonalityCatalogResponse>(
    "/api/v1/settings/personality/catalog",
    { signal },
  );
}

export async function updatePersonalityProfile(
  input: UpdatePersonalityInput,
): Promise<PersonalityProfile> {
  return apiRequest<PersonalityProfile>("/api/v1/settings/personality", {
    method: "PATCH",
    body: input,
  });
}

export async function previewPersonality(
  input: PersonalityPreviewInput,
): Promise<PersonalityPreview> {
  return apiRequest<PersonalityPreview>("/api/v1/settings/personality/preview", {
    method: "POST",
    body: input,
  });
}
