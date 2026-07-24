import { getLandingContent } from "./Landing.logic";
import { landingStyles as styles } from "./Landing.styles";
import { LandingAtmosphere } from "./sections/LandingAtmosphere";
import { LandingCapabilities } from "./sections/LandingCapabilities";
import { LandingClose } from "./sections/LandingClose";
import { LandingHero } from "./sections/LandingHero";
import { LandingNav } from "./sections/LandingNav";
import { LandingPromise } from "./sections/LandingPromise";
import { LandingRhythm } from "./sections/LandingRhythm";
import { LandingStats } from "./sections/LandingStats";

export function Landing() {
  const { copy } = getLandingContent();

  return (
    <div className={styles.page}>
      <LandingAtmosphere />
      <LandingNav
        navLinks={copy.navLinks}
        getStarted={copy.getStarted}
      />
      <main className={styles.main}>
        <LandingHero {...copy.hero} />
        <LandingStats items={copy.stats} />
        <LandingPromise
          eyebrow={copy.promiseEyebrow}
          title={copy.promiseTitle}
          body={copy.promiseBody}
        />
        <LandingCapabilities
          eyebrow={copy.capabilitiesEyebrow}
          title={copy.capabilitiesTitle}
          description={copy.capabilitiesDescription}
          items={copy.capabilities}
        />
        <LandingRhythm
          eyebrow={copy.rhythmEyebrow}
          title={copy.rhythmTitle}
          description={copy.rhythmDescription}
          steps={copy.rhythm}
        />
        <LandingClose
          eyebrow={copy.closeEyebrow}
          title={copy.closeTitle}
          body={copy.closeBody}
          cta={copy.closeCta}
          brand={copy.brand}
        />
      </main>
    </div>
  );
}
