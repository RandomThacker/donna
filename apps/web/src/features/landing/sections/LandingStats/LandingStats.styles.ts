export const landingStatsStyles = {
  section: "relative",
  grid: [
    "grid gap-px overflow-hidden rounded-2xl border border-donna-glass-border",
    "bg-donna-hairline sm:grid-cols-3",
  ].join(" "),
  item: "bg-donna-glass px-7 py-8 backdrop-blur-md",
  value: "font-display text-4xl italic text-donna-copper-bright sm:text-5xl",
  label: "mt-2 text-sm leading-snug text-donna-muted",
} as const;
