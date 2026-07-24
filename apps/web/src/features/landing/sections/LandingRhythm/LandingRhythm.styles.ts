export const landingRhythmStyles = {
  section: "relative py-24 sm:py-28",
  heading: "mb-16",
  track: "relative grid gap-10 md:grid-cols-3 md:gap-8",
  line: [
    "pointer-events-none absolute left-0 right-0 top-6 hidden h-px md:block",
    "bg-gradient-to-r from-transparent via-donna-copper/30 to-transparent",
  ].join(" "),
  step: "relative animate-donna-fade-up",
  stepDelays: [
    "[animation-delay:0ms]",
    "[animation-delay:120ms]",
    "[animation-delay:240ms]",
  ],
  node: [
    "relative z-10 mb-6 grid h-12 w-12 place-items-center rounded-full",
    "border border-donna-copper/30 bg-donna-bronze-950 text-donna-copper-bright",
    "shadow-[0_0_24px_-6px_rgb(203_169_125_/_0.6)]",
  ].join(" "),
  time: "mb-2 font-sans text-xs font-medium uppercase tracking-[0.22em] text-donna-copper-dim",
  title: "mb-2.5 font-display text-2xl text-donna-cream",
  body: "max-w-xs text-[0.95rem] leading-relaxed text-donna-muted",
} as const;
