import { BentoBox } from "../BentoBox";
import { greetingStyles as styles } from "./DashboardGreeting.styles";
import type { DashboardGreetingProps } from "./DashboardGreeting.types";

export function DashboardGreeting({ greeting }: DashboardGreetingProps) {
  return (
    <BentoBox className={styles.box}>
      <h1 className={styles.title}>
        {greeting.salutation},{" "}
        <span className={styles.name}>{greeting.name}</span>
      </h1>
      <div className={styles.row}>
        <p className={styles.summary}>{greeting.summary}</p>
        <p className={styles.nudge}>{greeting.nudge}</p>
      </div>
    </BentoBox>
  );
}
