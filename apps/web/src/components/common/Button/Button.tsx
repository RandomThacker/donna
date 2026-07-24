import { cn } from "@/lib/cn";

import { buttonStyles as styles } from "./Button.styles";
import type { ButtonProps } from "./Button.types";

export function Button({
  children,
  href,
  variant = "primary",
  size = "md",
  className,
  external = false,
  iconLeft,
  iconRight,
}: ButtonProps) {
  return (
    <a
      href={href}
      className={cn(
        styles.base,
        styles.variants[variant],
        styles.sizes[size],
        className,
      )}
      {...(external
        ? { target: "_blank", rel: "noopener noreferrer" }
        : undefined)}
    >
      {variant === "primary" ? <span className={styles.shine} aria-hidden /> : null}
      <span className={styles.label}>
        {iconLeft}
        {children}
        {iconRight ? <span className={styles.iconRight}>{iconRight}</span> : null}
      </span>
    </a>
  );
}
