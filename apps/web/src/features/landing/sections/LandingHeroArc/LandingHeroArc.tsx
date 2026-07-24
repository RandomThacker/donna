import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { HeroChart } from "./HeroChart";
import { heroArcStyles as styles } from "./LandingHeroArc.styles";
import type { LandingHeroArcProps } from "./LandingHeroArc.types";

export function LandingHeroArc({ chips, cards }: LandingHeroArcProps) {
  return (
    <div className={styles.root} aria-hidden>
      <div className={styles.glow} />
      <div className={styles.grid} />
      <div className={styles.rings}>
        {styles.ringItems.map((ring, index) => (
          <span key={index} className={cn(styles.ringBase, ring)} />
        ))}
      </div>

      {chips.map((chip, index) => (
        <span
          key={chip.label}
          className={cn(styles.chipBase, styles.chipPositions[index])}
        >
          <Icon name={chip.icon} className={cn("h-4 w-4", styles.chipIcon)} />
          {chip.label}
        </span>
      ))}

      {cards.map((card, index) => (
        <div
          key={card.title}
          className={cn(styles.cardBase, styles.cardPositions[index])}
        >
          <div className={styles.cardHead}>
            <Icon name={card.icon} className="h-4 w-4" />
            <span className={styles.cardTitle}>{card.title}</span>
          </div>
          <p className={styles.cardValue}>{card.value}</p>
          <p className={styles.cardSub}>{card.sub}</p>
          <div className={styles.cardChart}>
            <HeroChart kind={card.chart} />
          </div>
        </div>
      ))}
    </div>
  );
}
