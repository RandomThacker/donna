import { BentoBox } from "../BentoBox";
import { insightsStyles as styles } from "./DashboardInsights.styles";
import type { DashboardInsightsProps } from "./DashboardInsights.types";

export function DashboardInsights({ insights }: DashboardInsightsProps) {
  return (
    <BentoBox className={styles.box} title="Donna noticed">
      <ul className={styles.list}>
        {insights.map((insight) => (
          <li key={insight.id} className={styles.item}>
            {insight.text}
          </li>
        ))}
      </ul>
    </BentoBox>
  );
}
