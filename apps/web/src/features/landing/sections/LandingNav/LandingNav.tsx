"use client";

import { Container, Logo } from "@/components/common";
import { AuthEntryCta } from "@/features/auth";

import { landingNavStyles as styles } from "./LandingNav.styles";
import type { LandingNavProps } from "./LandingNav.types";

export function LandingNav({ navLinks, getStarted }: LandingNavProps) {
  return (
    <header className={styles.header}>
      <Container>
        <div className={styles.bar}>
          <Logo size="sm" />
          <nav className={styles.links} aria-label="Primary">
            {navLinks.map((link) => (
              <a key={link.href} href={link.href} className={styles.link}>
                {link.label}
              </a>
            ))}
          </nav>
          <div className={styles.actions}>
            <AuthEntryCta label={getStarted.label} />
          </div>
        </div>
      </Container>
    </header>
  );
}
