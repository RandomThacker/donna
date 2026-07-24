import type { TextTone } from "./Text.types";

export const textStyles = {
  tones: {
    cream: "text-donna-cream",
    muted: "text-donna-muted",
    copper: "text-donna-copper",
  } satisfies Record<TextTone, string>,
  body: "font-sans text-base leading-relaxed sm:text-lg",
  display: "font-display text-4xl leading-[1.05] tracking-tight sm:text-5xl md:text-6xl",
  title: "font-display text-3xl leading-tight tracking-tight sm:text-4xl",
  eyebrow: "font-sans text-xs font-medium uppercase tracking-[0.22em] text-donna-copper-dim",
} as const;
