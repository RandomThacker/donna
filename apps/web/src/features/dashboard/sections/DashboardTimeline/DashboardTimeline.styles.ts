export const timelineStyles = {
  box: "col-span-12 md:col-span-6",
  list: "space-y-0",
  item: [
    "grid w-full grid-cols-[3.5rem_1fr] gap-3 border-b border-donna-hairline py-2.5",
    "text-left transition-colors last:border-b-0",
    "hover:bg-donna-accent-soft/40 focus-visible:outline-none",
    "focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-donna-accent/40",
  ].join(" "),
  time: "pt-0.5 text-xs tabular-nums text-donna-faint",
  itemBody: "min-w-0",
  title: "text-sm text-donna-text",
  meta: "mt-0.5 truncate text-[0.65rem] uppercase tracking-[0.14em] text-donna-faint",
  state: "space-y-2 text-sm text-donna-muted",
  link: "font-medium text-donna-accent hover:underline",
} as const;
