export const listStyles = {
  toolbar: "space-y-3 border-b border-donna-hairline px-4 py-3",
  search: [
    "h-9 w-full rounded-xl border border-donna-border bg-donna-surface-2 px-3",
    "text-sm text-donna-text placeholder:text-donna-faint outline-none",
    "focus:border-donna-accent/50",
  ].join(" "),
  filters: "flex flex-wrap gap-1.5",
  filterChip: [
    "rounded-full border border-donna-border px-2.5 py-1 text-[0.7rem]",
    "text-donna-muted transition-colors hover:border-donna-accent/35",
  ].join(" "),
  filterChipOn: "border-donna-accent/50 bg-donna-accent-soft text-donna-accent",
  groups: "space-y-4 px-2 py-3",
  groupLabel: [
    "px-2 pb-1 text-[0.65rem] font-semibold uppercase tracking-[0.14em]",
    "text-donna-faint",
  ].join(" "),
  empty: "px-6 py-16 text-center",
  emptyTitle: "font-display text-xl text-donna-text",
  emptyBody: "mt-2 text-sm text-donna-muted",
  state: "px-4 py-8 text-center text-sm text-donna-muted",
  loadMoreWrap: "px-4 pb-4",
  loadMore: [
    "w-full rounded-xl border border-donna-border py-2.5 text-sm text-donna-muted",
    "transition-colors hover:border-donna-accent/40 hover:text-donna-text",
  ].join(" "),
} as const;
