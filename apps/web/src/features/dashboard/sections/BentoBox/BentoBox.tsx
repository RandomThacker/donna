import { cn } from "@/lib/cn";

import { bentoBoxStyles as styles } from "./BentoBox.styles";
import type { BentoBoxProps } from "./BentoBox.types";

export function BentoBox({ children, className, title }: BentoBoxProps) {
  return (
    <section className={cn(styles.root, className)}>
      {title ? <h2 className={styles.title}>{title}</h2> : null}
      {children}
    </section>
  );
}
