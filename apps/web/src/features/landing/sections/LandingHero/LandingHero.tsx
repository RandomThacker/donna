"use client";

import { Button, Container, Icon, Pill } from "@/components/common";
import { AuthEntryCta } from "@/features/auth";

import { LandingHeroArc } from "../LandingHeroArc";
import { landingHeroStyles as styles } from "./LandingHero.styles";
import type { LandingHeroProps } from "./LandingHero.types";

export function LandingHero({
  eyebrow,
  headlineLead,
  headlineEmphasis,
  support,
  primaryCta,
  secondaryCta,
  ctaNote,
  chips,
  cards,
}: LandingHeroProps) {
  return (
    <section className={styles.section} aria-labelledby="landing-hero-heading">
      <Container>
        <div className={styles.copy}>
          <Pill className={styles.eyebrow} withDot>
            {eyebrow}
          </Pill>
          <h1 id="landing-hero-heading" className={styles.headline}>
            {headlineLead}
            <span className={styles.emphasis}>{headlineEmphasis}</span>
          </h1>
          <p className={styles.support}>{support}</p>
          <div className={styles.actions}>
            <AuthEntryCta
              label={primaryCta.label}
              size="lg"
              withGoogleIcon
            />
            <Button
              href={secondaryCta.href}
              size="lg"
              variant="outline"
              iconRight={<Icon name="arrow" className="h-4 w-4" />}
            >
              {secondaryCta.label}
            </Button>
          </div>
          <p className={styles.note}>{ctaNote}</p>
        </div>
      </Container>
      <div className={styles.visual}>
        <LandingHeroArc chips={chips} cards={cards} />
      </div>
    </section>
  );
}
