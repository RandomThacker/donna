"use client";

import { Icon } from "@/components/common";

import { signInModalStyles as styles } from "./SignInModal.styles";
import type { SignInProvidersListProps } from "./SignInModal.types";

export function SignInProvidersList({
  providers,
  onSelect,
}: SignInProvidersListProps) {
  return (
    <div className={styles.list} role="list">
      {providers.map((provider) => (
        <button
          key={provider.id}
          type="button"
          role="listitem"
          className={styles.provider}
          disabled={!provider.enabled}
          onClick={() => onSelect(provider.id)}
        >
          <span className={styles.providerLabel}>
            <Icon name={provider.icon} className="h-5 w-5" />
            {provider.label}
          </span>
        </button>
      ))}
    </div>
  );
}
