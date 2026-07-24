export const landingCapabilitiesStyles = {
  section: "relative py-24 sm:py-28",
  heading: "mb-14",
  grid: "grid gap-5 sm:grid-cols-2 lg:grid-cols-4",
  iconWrap: [
    "mb-6 grid h-12 w-12 place-items-center rounded-xl",
    "border border-donna-glass-border bg-donna-bronze-950/60 text-donna-copper-bright",
    "transition-colors duration-300 group-hover:text-donna-copper-bright",
  ].join(" "),
  title: "mb-2.5 font-display text-2xl text-donna-cream",
  body: "text-[0.95rem] leading-relaxed text-donna-muted",
} as const;
