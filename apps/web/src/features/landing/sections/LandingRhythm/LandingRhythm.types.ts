import type { LandingRhythmStep } from "../../Landing.types";

export type LandingRhythmProps = {
  eyebrow: string;
  title: string;
  description: string;
  steps: LandingRhythmStep[];
};

export type RhythmStepProps = {
  step: LandingRhythmStep;
  index: number;
};
