import { cn } from "@/lib/cn";

import { logoStyles as styles } from "./Logo.styles";
import type { LogoProps } from "./Logo.types";

export function Logo({ href = "/", className, size = "sm" }: LogoProps) {
  return (
    <a href={href} className={cn(styles.root, className)} aria-label="Donna home">
      <span className={styles.mark} aria-hidden>
        <span className={styles.markCore} />
      </span>
      <span className={cn(styles.word, styles.sizes[size])}>Donna</span>
    </a>
  );
}
