import type { SignInProviderDefinition } from "./SignInModal.logic";

export type SignInProvidersListProps = {
  providers: readonly SignInProviderDefinition[];
  onSelect: (id: SignInProviderDefinition["id"]) => void;
};

export type SignInModalProps = {
  open: boolean;
  onClose: () => void;
  onSelectProvider: (id: SignInProviderDefinition["id"]) => void;
};
