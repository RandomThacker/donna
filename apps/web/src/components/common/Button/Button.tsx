import { cn } from "@/lib/cn";

import { buttonStyles as styles } from "./Button.styles";
import type { ButtonProps } from "./Button.types";

export function Button(props: ButtonProps) {
  const {
    children,
    variant = "primary",
    size = "md",
    className,
    iconLeft,
    iconRight,
  } = props;

  const classes = cn(
    styles.base,
    styles.variants[variant],
    styles.sizes[size],
    className,
  );

  const label = (
    <>
      {variant === "primary" ? <span className={styles.shine} aria-hidden /> : null}
      <span className={styles.label}>
        {iconLeft}
        {children}
        {iconRight ? <span className={styles.iconRight}>{iconRight}</span> : null}
      </span>
    </>
  );

  if ("href" in props && props.href) {
    return (
      <a
        href={props.href}
        className={classes}
        {...(props.external
          ? { target: "_blank", rel: "noopener noreferrer" }
          : undefined)}
      >
        {label}
      </a>
    );
  }

  return (
    <button
      type={props.type ?? "button"}
      className={classes}
      onClick={props.onClick}
      disabled={"disabled" in props ? props.disabled : undefined}
    >
      {label}
    </button>
  );
}
