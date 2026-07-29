import { greetingStyles as styles } from "./DashboardGreeting.styles";
import type { DashboardGreetingProps } from "./DashboardGreeting.types";

export function DashboardGreeting({ greeting }: DashboardGreetingProps) {
  return (
    <section className={styles.box}>
      <h1 className={styles.title}>
        {greeting.salutation},{" "}
        <span className={styles.name}>{greeting.name}</span>
        {greeting.emoji ? ` ${greeting.emoji}` : null}
      </h1>
      <div className={styles.row}>
        <p className={styles.summary}>{greeting.summary}</p>
      </div>
    </section>
  );
}
