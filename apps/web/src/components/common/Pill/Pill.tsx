import { cn } from "@/lib/cn";

import { pillStyles as styles } from "./Pill.styles";
import type { PillProps } from "./Pill.types";

export function Pill({ children, withDot = false, className }: PillProps) {
  return (
    <span className={cn(styles.root, className)}>
      {withDot ? (
        <span className={styles.dot} aria-hidden>
          <span className={styles.dotPing} />
          <span className={styles.dotCore} />
        </span>
      ) : null}
      {children}
    </span>
  );
}
