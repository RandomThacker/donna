export const bentoBoxStyles = {
  root: [
    "rounded-2xl border border-donna-border bg-donna-surface p-5",
    "shadow-donna-card",
    "transition-[border-color,transform] duration-200 hover:border-donna-accent/30",
  ].join(" "),
  title: "mb-4 shrink-0 text-[0.7rem] font-medium uppercase tracking-[0.18em] text-donna-faint",
  fixedPanel: "flex h-80 min-h-0 flex-col",
  scrollBody: "min-h-0 flex-1 overflow-y-auto scrollbar-hidden",
} as const;
