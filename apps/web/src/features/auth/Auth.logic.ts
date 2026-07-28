import { apiRequest } from "@/lib/api/client";
import {
  getGoogleOAuthStartUrl,
  getMicrosoftOAuthStartUrl,
} from "@/lib/api/config";

import type { AuthUser } from "./Auth.types";

const NEW_USER_KEY = "donna_is_new_user";

export function markNewUser(isNew: boolean): void {
  if (typeof window === "undefined") {
    return;
  }
  if (isNew) {
    sessionStorage.setItem(NEW_USER_KEY, "1");
  } else {
    sessionStorage.removeItem(NEW_USER_KEY);
  }
}

export function consumeNewUserFlag(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  const value = sessionStorage.getItem(NEW_USER_KEY) === "1";
  sessionStorage.removeItem(NEW_USER_KEY);
  return value;
}

export function clearClientAuthFlags(): void {
  if (typeof window === "undefined") {
    return;
  }

  const purgeDonnaKeys = (storage: Storage) => {
    const keys: string[] = [];
    for (let i = 0; i < storage.length; i += 1) {
      const key = storage.key(i);
      if (key?.startsWith("donna_")) {
        keys.push(key);
      }
    }
    for (const key of keys) {
      storage.removeItem(key);
    }
  };

  purgeDonnaKeys(sessionStorage);
  purgeDonnaKeys(localStorage);
}

export function startGoogleOAuth(): void {
  const returnTo = `${window.location.origin}/auth/callback`;
  const url = new URL(getGoogleOAuthStartUrl(), window.location.origin);
  url.searchParams.set("return_to", returnTo);
  window.location.assign(url.toString());
}

export function startMicrosoftOAuth(): void {
  window.location.assign(getMicrosoftOAuthStartUrl());
}

/** Loads the current user using the HttpOnly session cookie. */
export async function fetchCurrentUser(): Promise<AuthUser> {
  return apiRequest<AuthUser>("/api/v1/me");
}

export async function logoutSession(): Promise<void> {
  try {
    await apiRequest<unknown>("/api/v1/auth/logout", { method: "POST" });
  } catch {
    // Local sign-out should still succeed if the API is unreachable.
  }
}

export function parseOAuthCallback(params: URLSearchParams): {
  ok: boolean;
  isNewUser: boolean;
  error: string | null;
} {
  const status = params.get("status");
  const error = params.get("error");
  return {
    ok: status === "ok",
    isNewUser: params.get("new_user") === "1",
    error,
  };
}
