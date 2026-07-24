import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { landingRhythmStyles as styles } from "./LandingRhythm.styles";
import type { RhythmStepProps } from "./LandingRhythm.types";

export function RhythmStep({ step, index }: RhythmStepProps) {
  const delay = styles.stepDelays[index] ?? styles.stepDelays[0];

  return (
    <div className={cn(styles.step, delay)}>
      <span className={styles.node}>
        <Icon name={step.icon} className="h-6 w-6" />
      </span>
      <p className={styles.time}>{step.time}</p>
      <h3 className={styles.title}>{step.title}</h3>
      <p className={styles.body}>{step.description}</p>
    </div>
  );
}
