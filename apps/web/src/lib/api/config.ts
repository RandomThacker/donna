export function getApiBaseUrl(): string {
  const base = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (!base) {
    return "http://localhost:8080";
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
