export const bentoBoxStyles = {
  root: [
    "rounded-2xl border border-donna-border bg-donna-surface p-5",
    "shadow-donna-card",
    "transition-[border-color,transform] duration-200 hover:border-donna-accent/30",
  ].join(" "),
  title: "mb-4 text-[0.7rem] font-medium uppercase tracking-[0.18em] text-donna-faint",
} as const;
