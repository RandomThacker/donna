import { cn } from "@/lib/cn";

import { cardStyles as styles } from "./Card.styles";
import type { CardProps } from "./Card.types";

export function Card({ children, interactive = false, className }: CardProps) {
  return (
    <div className={cn(styles.base, interactive && styles.interactive, className)}>
      <span className={styles.sheen} aria-hidden />
      {interactive ? <span className={styles.glow} aria-hidden /> : null}
      {children}
    </div>
  );
}
