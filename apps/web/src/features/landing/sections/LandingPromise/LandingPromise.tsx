import { Container, Pill } from "@/components/common";

import { landingPromiseStyles as styles } from "./LandingPromise.styles";
import type { LandingPromiseProps } from "./LandingPromise.types";

export function LandingPromise({ eyebrow, title, body }: LandingPromiseProps) {
  return (
    <section className={styles.section} aria-labelledby="landing-promise-heading">
      <Container className={styles.grid}>
        <div>
          <Pill className={styles.eyebrow}>{eyebrow}</Pill>
          <h2 id="landing-promise-heading" className={styles.title}>
            {title}
          </h2>
        </div>
        <p className={styles.body}>{body}</p>
      </Container>
    </section>
  );
}
