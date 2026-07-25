export type ButtonVariant = "primary" | "ghost" | "outline";
export type ButtonSize = "md" | "lg";

type ButtonSharedProps = {
  children: React.ReactNode;
  variant?: ButtonVariant;
  size?: ButtonSize;
  className?: string;
  iconLeft?: React.ReactNode;
  iconRight?: React.ReactNode;
};

export type ButtonAsLinkProps = ButtonSharedProps & {
  href: string;
  onClick?: never;
  external?: boolean;
  type?: never;
};

export type ButtonAsButtonProps = ButtonSharedProps & {
  href?: never;
  onClick?: () => void;
  external?: never;
  type?: "button" | "submit";
  disabled?: boolean;
};

export type ButtonProps = ButtonAsLinkProps | ButtonAsButtonProps;
