export const cardStyles = {
  root: [
    "w-full rounded-xl border border-transparent px-3 py-3 text-left",
    "transition-colors hover:border-donna-hairline hover:bg-donna-surface-2/70",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  top: "flex items-start gap-3",
  iconWrap: [
    "mt-0.5 grid h-8 w-8 shrink-0 place-items-center rounded-lg",
    "bg-donna-accent-soft text-donna-accent",
  ].join(" "),
  main: "min-w-0 flex-1",
  titleRow: "flex items-start justify-between gap-2",
  title: "text-sm font-medium text-donna-text",
  relative: "shrink-0 text-[0.7rem] text-donna-faint",
  body: "mt-0.5 line-clamp-2 text-xs text-donna-muted",
  meta: "mt-2 flex flex-wrap items-center gap-1.5",
  chip: "inline-flex rounded-full px-2 py-0.5 text-[0.65rem] font-medium",
  mutedChip: [
    "inline-flex rounded-full border border-donna-hairline bg-donna-bg/40",
    "px-2 py-0.5 text-[0.65rem] text-donna-faint",
  ].join(" "),
  times: "mt-1.5 space-y-0.5 text-[0.7rem] text-donna-faint",
  fail: "mt-1.5 text-[0.7rem] text-rose-400",
  actions: "mt-2.5 flex flex-wrap gap-1.5",
  actionBtn: [
    "rounded-full border border-donna-border px-2.5 py-1 text-[0.7rem]",
    "text-donna-muted transition-colors hover:border-donna-accent/40 hover:text-donna-text",
    "disabled:opacity-50",
  ].join(" "),
} as const;
