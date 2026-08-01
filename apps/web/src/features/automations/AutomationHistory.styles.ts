export const automationHistoryStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4.25rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: [
    "min-h-0 min-w-0 overflow-x-hidden overflow-y-auto scrollbar-hidden",
    "bg-donna-bg",
  ].join(" "),
  workspaceInner: [
    "mx-auto flex w-full max-w-3xl flex-col gap-6 overflow-x-hidden p-5 sm:p-6 lg:p-8",
  ].join(" "),
  title: "font-display text-3xl tracking-tight text-donna-text sm:text-4xl",
  subtitle: "mt-2 text-sm leading-relaxed text-donna-muted",
  stats: [
    "grid gap-3 sm:grid-cols-2 lg:grid-cols-4",
    "rounded-2xl border border-donna-hairline bg-donna-surface p-4",
  ].join(" "),
  stat: "flex flex-col gap-1",
  statLabel: "text-xs text-donna-muted",
  statValue: "text-sm font-medium text-donna-text",
  list: [
    "rounded-2xl border border-donna-hairline bg-donna-surface",
    "divide-y divide-donna-hairline",
  ].join(" "),
  card: "flex flex-col gap-3 px-5 py-4",
  cardHeader: "flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between",
  cardTitle: "text-sm font-medium text-donna-text",
  cardMeta: "text-xs text-donna-muted",
  expandBtn: [
    "inline-flex items-center gap-1 rounded-full border border-donna-border",
    "bg-donna-surface-2 px-3 py-1.5 text-xs text-donna-text",
    "hover:border-donna-accent/40 hover:bg-donna-accent-soft",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  commands: "flex flex-col gap-2 border-t border-donna-hairline pt-3",
  commandRow: [
    "grid gap-1 rounded-xl border border-donna-hairline bg-donna-surface-2 px-3 py-2",
    "sm:grid-cols-[1fr_auto_auto] sm:items-center",
  ].join(" "),
  commandText: "text-sm text-donna-text",
  commandStatus: "text-xs text-donna-muted",
  empty: "px-5 py-8 text-sm text-donna-muted",
  error: "text-sm text-rose-300",
  debug: [
    "mt-2 overflow-x-auto rounded-xl border border-donna-hairline",
    "bg-donna-elevated p-3 text-[11px] leading-relaxed text-donna-muted",
  ].join(" "),
} as const;
