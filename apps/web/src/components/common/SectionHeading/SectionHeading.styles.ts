import type { SectionHeadingAlign } from "./SectionHeading.types";

export const sectionHeadingStyles = {
  root: {
    left: "flex max-w-2xl flex-col items-start",
    center: "mx-auto flex max-w-2xl flex-col items-center text-center",
  } satisfies Record<SectionHeadingAlign, string>,
  eyebrow: "mb-5",
  title: "font-display text-3xl leading-[1.08] tracking-tight text-donna-cream sm:text-4xl md:text-5xl",
  description: "mt-5 max-w-xl text-base leading-relaxed text-donna-muted sm:text-lg",
} as const;
