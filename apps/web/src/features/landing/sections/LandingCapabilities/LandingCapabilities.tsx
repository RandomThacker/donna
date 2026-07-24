import { Container, SectionHeading } from "@/components/common";

import { CapabilityCard } from "./CapabilityCard";
import { landingCapabilitiesStyles as styles } from "./LandingCapabilities.styles";
import type { LandingCapabilitiesProps } from "./LandingCapabilities.types";

export function LandingCapabilities({
  eyebrow,
  title,
  description,
  items,
}: LandingCapabilitiesProps) {
  return (
    <section
      id="capabilities"
      className={styles.section}
      aria-labelledby="landing-capabilities-heading"
    >
      <Container>
        <SectionHeading
          id="landing-capabilities-heading"
          eyebrow={eyebrow}
          title={title}
          description={description}
          className={styles.heading}
        />
        <div className={styles.grid}>
          {items.map((item) => (
            <CapabilityCard key={item.title} item={item} />
          ))}
        </div>
      </Container>
    </section>
  );
}
