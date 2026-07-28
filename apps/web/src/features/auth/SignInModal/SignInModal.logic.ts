import type { IconName } from "@/components/common";

/**
 * Stable provider ids. Add new options here as auth methods ship.
 * Keep this as a string union so UI stays typed without coupling to OAuth URLs.
 */
export type SignInProviderId = "google" | "microsoft";

export type SignInProviderDefinition = {
  id: SignInProviderId;
  label: string;
  icon: IconName;
  /** When false, the option stays listed but disabled (useful for coming-soon). */
  enabled: boolean;
};

export type SignInProviderActions = Record<SignInProviderId, () => void>;

/** Registry of available sign-in methods. Extend this list to add providers. */
export const SIGN_IN_PROVIDERS: readonly SignInProviderDefinition[] = [
  {
    id: "google",
    label: "Continue with Google",
    icon: "google",
    enabled: true,
  },
  {
    id: "microsoft",
    label: "Continue with Microsoft",
    icon: "microsoft",
    enabled: true,
  },
] as const;

export function listEnabledSignInProviders(): SignInProviderDefinition[] {
  return SIGN_IN_PROVIDERS.filter((provider) => provider.enabled);
}

export function startSignInProvider(
  id: SignInProviderId,
  actions: SignInProviderActions,
): void {
  actions[id]();
}
