export const dateMarkStyles = {
  root: [
    "inline-flex min-w-0 items-center gap-2 rounded-xl text-left sm:gap-3",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-donna-accent/40",
  ].join(" "),
  icon: [
    "grid h-9 w-9 shrink-0 place-items-center rounded-xl sm:h-10 sm:w-10",
    "border border-donna-border text-donna-muted",
  ].join(" "),
  date: "flex min-w-0 items-baseline gap-1.5 sm:gap-2",
  dayNumber: [
    "font-display text-2xl font-semibold leading-none tracking-tight sm:text-[1.75rem]",
    "text-donna-accent",
  ].join(" "),
  meta: "truncate text-xs text-donna-muted sm:text-sm",
} as const;
