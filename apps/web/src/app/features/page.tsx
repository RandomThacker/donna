import type { Metadata } from "next";

import { FeaturesPage } from "@/features/landing/FeaturesPage";

export const metadata: Metadata = {
  title: "Features — Donna",
  description:
    "What Donna is and what she can do: dashboard, calendar, command chat, automations, personality, and more.",
};

export default function FeaturesRoutePage() {
  return <FeaturesPage />;
}
