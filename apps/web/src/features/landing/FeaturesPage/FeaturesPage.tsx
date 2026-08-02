"use client";

import Link from "next/link";

import { AuthEntryCta } from "@/features/auth";
import { Icon, Logo } from "@/components/common";

import { getLandingContent } from "../Landing.logic";
import { LandingAtmosphere } from "../sections/LandingAtmosphere";
import { LandingNav } from "../sections/LandingNav";
import { getFeaturesContent } from "./FeaturesPage.logic";
import { featuresPageStyles as styles } from "./FeaturesPage.styles";

export function FeaturesPage() {
  const { copy: landing } = getLandingContent();
  const { copy } = getFeaturesContent();
  const year = new Date().getFullYear();

  return (
    <div className={styles.page}>
      <LandingAtmosphere />
      <LandingNav navLinks={landing.navLinks} getStarted={landing.getStarted} />

      <main className={styles.main}>
        <div className={styles.stage}>
          <div className={styles.grid} aria-hidden />
          <div className={styles.glowA} aria-hidden />
          <div className={styles.glowB} aria-hidden />

          <header className={styles.hero}>
            <span className={styles.badge}>
              <span className={styles.badgeDot} aria-hidden />
              {copy.eyebrow}
            </span>
            <h1 className={styles.headline}>
              {copy.headlineLead}
              <span className={styles.emphasis}>{copy.headlineEmphasis}</span>
            </h1>
            <p className={styles.subhead}>{copy.subhead}</p>
          </header>

          <div className={styles.pillars}>
            {copy.pillars.map((pillar) => (
              <article key={pillar.id} className={styles.pillar}>
                <h2 className={styles.pillarTitle}>{pillar.title}</h2>
                <p className={styles.pillarBody}>{pillar.body}</p>
              </article>
            ))}
          </div>

          <section className={styles.section} aria-labelledby="who-donna">
            <div className={styles.sectionHead}>
              <p className={styles.sectionEyebrow}>
                <span className={styles.sectionRule} aria-hidden />
                {copy.whoEyebrow}
              </p>
              <h2 id="who-donna" className={styles.sectionTitle}>
                {copy.whoTitle}
              </h2>
            </div>
            <div className={styles.whoPanel}>
              <div className={styles.whoGlow} aria-hidden />
              <div className={styles.whoBody}>
                {copy.whoBody.map((paragraph, index) => (
                  <p
                    key={paragraph.slice(0, 32)}
                    className={index === 0 ? styles.whoLead : undefined}
                  >
                    {paragraph}
                  </p>
                ))}
              </div>
            </div>
          </section>

          <section className={styles.section} aria-labelledby="differentiators">
            <div className={styles.sectionHead}>
              <p className={styles.sectionEyebrow}>
                <span className={styles.sectionRule} aria-hidden />
                {copy.highlightsEyebrow}
              </p>
              <h2 id="differentiators" className={styles.sectionTitle}>
                {copy.highlightsTitle}
              </h2>
              <p className={styles.sectionDesc}>{copy.highlightsDescription}</p>
            </div>

            <div className={styles.highlights}>
              {copy.highlights.map((highlight, index) => (
                <article
                  key={highlight.id}
                  className={styles.highlight}
                  style={{ animationDelay: `${index * 90}ms` }}
                >
                  <span className={styles.highlightGlow} aria-hidden />
                  <span className={styles.highlightIndex} aria-hidden>
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  <span className={styles.highlightIcon} aria-hidden>
                    <Icon name={highlight.icon} className="h-5 w-5" />
                  </span>
                  <p className={styles.highlightKicker}>{highlight.kicker}</p>
                  <h3 className={styles.highlightTitle}>{highlight.title}</h3>
                  <p className={styles.highlightBody}>{highlight.body}</p>
                  <ul className={styles.highlightPoints}>
                    {highlight.points.map((point) => (
                      <li key={point} className={styles.highlightPoint}>
                        <span className={styles.highlightBullet} aria-hidden />
                        {point}
                      </li>
                    ))}
                  </ul>
                </article>
              ))}
            </div>
          </section>

          <section className={styles.section} aria-labelledby="core-features">
            <div className={styles.sectionHeadCentered}>
              <p className={styles.sectionEyebrow}>{copy.coreEyebrow}</p>
              <h2 id="core-features" className={styles.sectionTitle}>
                {copy.coreTitle}
              </h2>
              <p className={styles.sectionDesc}>{copy.coreDescription}</p>
            </div>

            <div className={styles.coreGrid}>
              {copy.core.map((feature) => (
                <article key={feature.id} className={styles.coreCard}>
                  <div className={styles.coreTop}>
                    <span className={styles.coreIcon} aria-hidden>
                      <Icon name={feature.icon} className="h-4 w-4" />
                    </span>
                    <h3 className={styles.coreTitle}>{feature.title}</h3>
                  </div>
                  <p className={styles.coreBlurb}>{feature.blurb}</p>
                </article>
              ))}
            </div>

            <div className={styles.horizon}>
              <p className={styles.horizonLabel}>{copy.horizonLabel}</p>
              <ul className={styles.horizonList}>
                {copy.horizon.map((item) => (
                  <li key={item} className={styles.horizonPill}>
                    <span className={styles.horizonDot} aria-hidden />
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          </section>

          <section className={styles.close} aria-labelledby="features-close">
            <div className={styles.closeGlow} aria-hidden />
            <p className={styles.sectionEyebrow}>{copy.closeEyebrow}</p>
            <h2 id="features-close" className={styles.closeTitle}>
              {copy.closeTitle}
            </h2>
            <p className={styles.closeBody}>{copy.closeBody}</p>
            <div className={styles.closeActions}>
              <AuthEntryCta label={copy.closeCtaLabel} size="lg" withArrowIcon />
              <Link href="/" className={styles.secondaryBtn}>
                Back to home
              </Link>
            </div>
            <p className={styles.footnote}>{copy.footnote}</p>
          </section>

          <footer className={styles.footer}>
            <Logo size="sm" />
            <p>© {year} Donna. Personal AI Operating System.</p>
          </footer>
        </div>
      </main>
    </div>
  );
}
