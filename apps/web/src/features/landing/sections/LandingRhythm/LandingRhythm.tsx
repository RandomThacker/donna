import { Container, SectionHeading } from "@/components/common";

import { landingRhythmStyles as styles } from "./LandingRhythm.styles";
import type { LandingRhythmProps } from "./LandingRhythm.types";
import { RhythmStep } from "./RhythmStep";

export function LandingRhythm({
  eyebrow,
  title,
  description,
  steps,
}: LandingRhythmProps) {
  return (
    <section
      id="rhythm"
      className={styles.section}
      aria-labelledby="landing-rhythm-heading"
    >
      <Container>
        <SectionHeading
          id="landing-rhythm-heading"
          eyebrow={eyebrow}
          title={title}
          description={description}
          className={styles.heading}
        />
        <div className={styles.track}>
          <span className={styles.line} aria-hidden />
          {steps.map((step, index) => (
            <RhythmStep key={step.time} step={step} index={index} />
          ))}
        </div>
      </Container>
    </section>
  );
}
