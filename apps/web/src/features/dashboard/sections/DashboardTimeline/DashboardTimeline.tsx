import { BentoBox } from "../BentoBox";
import { timelineStyles as styles } from "./DashboardTimeline.styles";
import type { DashboardTimelineProps } from "./DashboardTimeline.types";

const kindLabel = {
  meeting: "Meeting",
  focus: "Focus",
  personal: "Personal",
} as const;

export function DashboardTimeline({ items }: DashboardTimelineProps) {
  return (
    <BentoBox className={styles.box} title="Today's timeline">
      <ol className={styles.list}>
        {items.map((item) => (
          <li key={`${item.time}-${item.title}`} className={styles.item}>
            <time className={styles.time}>{item.time}</time>
            <div>
              <p className={styles.title}>{item.title}</p>
              <p className={styles.kind}>{kindLabel[item.kind]}</p>
            </div>
          </li>
        ))}
      </ol>
    </BentoBox>
  );
}
