export type ButtonVariant = "primary" | "ghost" | "outline";
export type ButtonSize = "md" | "lg";

export type ButtonProps = {
  children: React.ReactNode;
  href: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
  className?: string;
  external?: boolean;
  iconLeft?: React.ReactNode;
  iconRight?: React.ReactNode;
};
