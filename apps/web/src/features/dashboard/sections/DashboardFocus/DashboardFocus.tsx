import { Button, Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { BentoBox } from "../BentoBox";
import { focusStyles as styles } from "./DashboardFocus.styles";
import type { DashboardFocusProps } from "./DashboardFocus.types";

export function DashboardFocus({ focus }: DashboardFocusProps) {
  const percent = Math.round(focus.progress * 100);

  return (
    <BentoBox className={styles.box} title="Today's focus">
      <h3 className={styles.goal}>{focus.goal}</h3>
      <div className={styles.meta}>
        <div className={styles.progressWrap}>
          <p className={styles.progressLabel}>{percent}% complete</p>
          <div
            className={styles.progressTrack}
            role="progressbar"
            aria-valuenow={percent}
            aria-valuemin={0}
            aria-valuemax={100}
          >
            <div
              className={cn(
                styles.progressFill,
                `[transform:scaleX(${focus.progress})]`,
              )}
            />
          </div>
        </div>
        <p className={styles.time}>{focus.timeRemaining}</p>
      </div>
      <Button
        href={focus.ctaHref}
        size="md"
        className={styles.cta}
        iconRight={<Icon name="arrow" className="h-4 w-4" />}
      >
        {focus.ctaLabel}
      </Button>
    </BentoBox>
  );
}
