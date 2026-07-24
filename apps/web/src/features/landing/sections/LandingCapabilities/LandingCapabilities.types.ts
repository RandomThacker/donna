import type { LandingCapability } from "../../Landing.types";

export type LandingCapabilitiesProps = {
  eyebrow: string;
  title: string;
  description: string;
  items: LandingCapability[];
};

export type CapabilityCardProps = {
  item: LandingCapability;
};
