"use client";

import { Modal } from "@/components/common";

import { listEnabledSignInProviders } from "./SignInModal.logic";
import { signInModalStyles as styles } from "./SignInModal.styles";
import type { SignInModalProps } from "./SignInModal.types";
import { SignInProvidersList } from "./SignInProvidersList";

export function SignInModal({ open, onClose, onSelectProvider }: SignInModalProps) {
  const providers = listEnabledSignInProviders();

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Sign in to Donna"
      description="Choose how you'd like to continue. Your session stays on Donna — provider tokens never land in the browser."
    >
      <SignInProvidersList providers={providers} onSelect={onSelectProvider} />
      <p className={styles.note}>Takes under a minute.</p>
    </Modal>
  );
}
