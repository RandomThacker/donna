export const bentoBoxStyles = {
  root: [
    "rounded-[1.35rem] border border-donna-border bg-donna-surface p-4",
    "shadow-donna-card sm:rounded-2xl sm:p-5",
    "transition-[border-color,transform] duration-200",
    "md:hover:border-donna-accent/30",
  ].join(" "),
  title: [
    "mb-3 shrink-0 text-[0.7rem] font-medium uppercase tracking-[0.18em]",
    "text-donna-faint sm:mb-4",
  ].join(" "),
  fixedPanel: "flex h-80 min-h-0 flex-col",
  scrollBody: "min-h-0 flex-1 overflow-y-auto scrollbar-hidden",
} as const;
