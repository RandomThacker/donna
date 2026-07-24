import { Button, Container, Icon, Logo, Pill } from "@/components/common";

import { landingCloseStyles as styles } from "./LandingClose.styles";
import type { LandingCloseProps } from "./LandingClose.types";

export function LandingClose({ eyebrow, title, body, cta, brand }: LandingCloseProps) {
  const year = new Date().getFullYear();

  return (
    <section className={styles.section} aria-labelledby="landing-close-heading">
      <Container>
        <div className={styles.panel}>
          <span className={styles.panelGlow} aria-hidden />
          <div className={styles.inner}>
            <Pill className={styles.eyebrow} withDot>
              {eyebrow}
            </Pill>
            <h2 id="landing-close-heading" className={styles.title}>
              {title}
            </h2>
            <p className={styles.body}>{body}</p>
            <Button
              href={cta.href}
              size="lg"
              className={styles.action}
              iconRight={<Icon name="arrow" className="h-4 w-4" />}
            >
              {cta.label}
            </Button>
          </div>
        </div>
        <footer className={styles.footer}>
          <Logo size="sm" />
          <p>
            © {year} {brand}. Personal AI Operating System.
          </p>
        </footer>
      </Container>
    </section>
  );
}
