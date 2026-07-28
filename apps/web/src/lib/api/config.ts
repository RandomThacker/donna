/**
 * API origin for browser fetches.
 *
 * - Unset → http://localhost:8080 (local default)
 * - "same-origin" or "/" or "" → relative /api/... (pair with Vercel API_PROXY_TARGET)
 * - Otherwise → absolute API host
 */
export function getApiBaseUrl(): string {
  const base = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();

  if (base === undefined) {
    return "http://localhost:8080";
  }

  if (base === "" || base === "/" || base === "same-origin") {
    return "";
  }

  return base.replace(/\/$/, "");
}

export function getGoogleOAuthStartUrl(): string {
  return `${getApiBaseUrl()}/api/v1/auth/google`;
}

export function getMicrosoftOAuthStartUrl(): string {
  return `${getApiBaseUrl()}/api/v1/auth/microsoft`;
}

export function getGoogleIntegrationStartUrl(): string {
  return `${getApiBaseUrl()}/api/v1/integrations/google`;
}

export function getMicrosoftIntegrationStartUrl(): string {
  return `${getApiBaseUrl()}/api/v1/integrations/microsoft`;
}
