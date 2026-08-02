import { apiRequest } from "@/lib/api/client";

import type {
  PersonalityCatalogResponse,
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
