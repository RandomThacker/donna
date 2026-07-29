export const journalStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: [
    "h-full min-h-0 min-w-0 overflow-x-hidden overflow-y-auto scrollbar-hidden",
    "bg-donna-bg lg:overflow-y-hidden",
  ].join(" "),
  workspaceInner: [
    "mx-auto flex h-full w-full max-w-6xl flex-col gap-4 overflow-hidden p-5 sm:p-6 lg:p-8",
  ].join(" "),
  main: [
    "flex min-h-0 flex-1 flex-col gap-4",
    "lg:flex-row lg:items-start lg:overflow-hidden",
  ].join(" "),
  // left col: mini calendar + tags (desktop only)
  sidebar: "hidden w-52 shrink-0 flex-col gap-4 overflow-y-auto scrollbar-hidden lg:flex",
  // center col: tasks
  tasksCol: [
    "flex min-h-0 flex-1 flex-col gap-4",
    "lg:h-full lg:overflow-hidden",
  ].join(" "),
  // right col: statistics
  statsCol: "shrink-0 lg:w-52",
  mobileTags: "block lg:hidden",
  header: "flex shrink-0 min-w-0 flex-nowrap items-center justify-between gap-2",
  nav: "flex shrink-0 items-center gap-1.5 sm:gap-2",
  navBtn: [
    "inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
    "border border-donna-border text-donna-muted transition-colors",
    "hover:border-donna-accent/40 hover:text-donna-text",
    "sm:h-9 sm:w-9",
  ].join(" "),
  todayBtn: [
    "inline-flex h-8 shrink-0 items-center justify-center rounded-full",
    "border border-donna-accent/40 bg-donna-accent-soft px-2.5",
    "text-xs font-medium text-donna-accent sm:h-9 sm:px-3.5 sm:text-sm",
  ].join(" "),
  tasksCard: [
    "flex min-h-0 flex-1 flex-col rounded-2xl border border-donna-hairline",
    "bg-donna-surface p-4 sm:p-5",
  ].join(" "),
  filterRow: "mb-3 flex flex-wrap items-center gap-1.5",
  filterHint: "text-xs text-donna-muted",
  addRow: "mb-4 flex shrink-0 gap-2",
  addInput: [
    "h-10 min-w-0 flex-1 rounded-xl border border-donna-border bg-donna-elevated/40",
    "px-3 text-sm text-donna-text placeholder:text-donna-muted/70",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  addBtn: [
    "inline-flex h-10 shrink-0 items-center justify-center rounded-xl",
    "bg-donna-accent px-4 text-sm font-medium text-donna-on-accent",
    "hover:bg-donna-accent-bright disabled:opacity-60",
  ].join(" "),
  tasksBody: "min-h-0 flex-1 overflow-y-auto scrollbar-hidden",
  list: "space-y-1",
  item: [
    "flex items-start gap-3 rounded-xl border border-transparent px-2 py-2.5",
    "transition-[background-color,border-color,opacity,box-shadow] duration-150",
    "hover:border-donna-hairline hover:bg-donna-surface-2/60",
  ].join(" "),
  itemDragging: "opacity-40",
  itemDropTarget: [
    "border-donna-accent/50 bg-donna-accent-soft/60",
    "shadow-[inset_0_0_0_1px_rgba(212,175,122,0.35)]",
  ].join(" "),
  checkbox: [
    "mt-0.5 grid h-5 w-5 shrink-0 place-items-center rounded border",
    "border-donna-border bg-donna-surface-2 cursor-pointer",
  ].join(" "),
  checkboxOn: "border-donna-accent bg-donna-accent text-donna-on-accent",
  itemBody: "min-w-0 flex-1",
  itemRow: "flex min-w-0 flex-wrap items-center gap-1.5",
  itemTitle: "text-sm text-donna-text",
  itemTitleDone: "text-donna-muted line-through",
  carriedPill: [
    "ml-2 inline-flex translate-y-[-1px] items-center rounded-full",
    "border border-donna-border bg-donna-surface-2 px-2 py-0.5",
    "align-middle text-[10px] font-medium tracking-wide text-donna-muted",
    "no-underline",
  ].join(" "),
  itemMeta: "mt-1 flex flex-wrap items-center gap-1.5 text-xs text-donna-faint",
  deleteBtn: [
    "mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-lg",
    "text-donna-faint transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
    "disabled:opacity-50",
  ].join(" "),
  dragHandle: [
    "mt-0.5 grid h-7 w-7 shrink-0 cursor-grab place-items-center rounded-lg",
    "text-donna-faint transition-colors active:cursor-grabbing",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  empty: "py-8 text-center text-sm text-donna-muted",
  statsCard: "shrink-0 self-start rounded-2xl border border-donna-hairline bg-donna-surface p-4",
  statsTitle: "mb-3 text-[11px] font-semibold uppercase tracking-[0.14em] text-donna-faint",
  statsGrid: "space-y-3",
  statRow: "flex items-center justify-between gap-3 text-sm",
  statLabel: "text-donna-muted",
  statValue: "font-medium text-donna-text",
  state: "text-sm text-donna-muted",
} as const;
