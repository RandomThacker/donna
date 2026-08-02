import type { IconName } from "@/components/common";

export type FeatureHighlight = {
  id: string;
  icon: IconName;
  kicker: string;
  title: string;
  body: string;
  points: string[];
};

export type CoreFeature = {
  id: string;
  icon: IconName;
  title: string;
  blurb: string;
};

export type FeaturePillar = {
  id: string;
  title: string;
  body: string;
};

export type FeaturesCopy = {
  eyebrow: string;
  headlineLead: string;
  headlineEmphasis: string;
  subhead: string;
  whoEyebrow: string;
  whoTitle: string;
  whoBody: string[];
  pillars: FeaturePillar[];
  highlightsEyebrow: string;
  highlightsTitle: string;
  highlightsDescription: string;
  highlights: FeatureHighlight[];
  coreEyebrow: string;
  coreTitle: string;
  coreDescription: string;
  core: CoreFeature[];
  horizonLabel: string;
  horizon: string[];
  closeEyebrow: string;
  closeTitle: string;
  closeBody: string;
  closeCtaLabel: string;
  footnote: string;
};
