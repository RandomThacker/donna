import { cn } from "@/lib/cn";

import { Pill } from "../Pill";
import { sectionHeadingStyles as styles } from "./SectionHeading.styles";
import type { SectionHeadingProps } from "./SectionHeading.types";

export function SectionHeading({
  eyebrow,
  title,
  description,
  align = "left",
  id,
  className,
}: SectionHeadingProps) {
  return (
    <div className={cn(styles.root[align], className)}>
      {eyebrow ? <Pill className={styles.eyebrow}>{eyebrow}</Pill> : null}
      <h2 id={id} className={styles.title}>
        {title}
      </h2>
      {description ? <p className={styles.description}>{description}</p> : null}
    </div>
  );
}
