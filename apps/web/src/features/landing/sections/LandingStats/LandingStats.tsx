import { Container } from "@/components/common";

import { landingStatsStyles as styles } from "./LandingStats.styles";
import type { LandingStatsProps } from "./LandingStats.types";

export function LandingStats({ items }: LandingStatsProps) {
  return (
    <section className={styles.section} aria-label="Donna at a glance">
      <Container>
        <div className={styles.grid}>
          {items.map((item) => (
            <div key={item.label} className={styles.item}>
              <p className={styles.value}>{item.value}</p>
              <p className={styles.label}>{item.label}</p>
            </div>
          ))}
        </div>
      </Container>
    </section>
  );
}
