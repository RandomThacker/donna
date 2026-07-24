import type { IconName } from "@/components/common";

export type LandingLink = {
  label: string;
  href: string;
};

export type LandingStat = {
  value: string;
  label: string;
};

export type LandingCapability = {
  icon: IconName;
  title: string;
  description: string;
};

export type LandingRhythmStep = {
  icon: IconName;
  time: string;
  title: string;
  description: string;
};

export type HeroChartKind = "line" | "bars";

export type LandingHeroChip = {
  icon: IconName;
  label: string;
};

export type LandingHeroCard = {
  icon: IconName;
  title: string;
  value: string;
  sub: string;
  chart: HeroChartKind;
};

export type LandingHeroCopy = {
  eyebrow: string;
  headlineLead: string;
  headlineEmphasis: string;
  support: string;
  primaryCta: LandingLink;
  secondaryCta: LandingLink;
  ctaNote: string;
  chips: LandingHeroChip[];
  cards: LandingHeroCard[];
};

export type LandingCopy = {
  brand: string;
  navLinks: LandingLink[];
  getStarted: LandingLink;
  hero: LandingHeroCopy;
  stats: LandingStat[];
  promiseEyebrow: string;
  promiseTitle: string;
  promiseBody: string;
  capabilitiesEyebrow: string;
  capabilitiesTitle: string;
  capabilitiesDescription: string;
  capabilities: LandingCapability[];
  rhythmEyebrow: string;
  rhythmTitle: string;
  rhythmDescription: string;
  rhythm: LandingRhythmStep[];
  closeEyebrow: string;
  closeTitle: string;
  closeBody: string;
  closeCta: LandingLink;
};
