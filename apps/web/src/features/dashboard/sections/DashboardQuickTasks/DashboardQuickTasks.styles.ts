export const quickTasksStyles = {
  box: "order-4 col-span-12 md:order-5 md:col-span-6",
  addRow: "mb-3 flex gap-2",
  addInput: [
    "h-9 min-w-0 flex-1 rounded-lg border border-donna-border bg-transparent",
    "px-3 text-sm text-donna-text placeholder:text-donna-muted/70",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  addBtn: [
    "inline-flex h-9 shrink-0 items-center justify-center rounded-lg",
    "bg-donna-accent px-3 text-sm font-medium text-donna-on-accent",
    "hover:bg-donna-accent-bright disabled:opacity-60",
  ].join(" "),
  list: "space-y-0.5",
  item: [
    "flex w-full items-center gap-3 rounded-lg px-1.5 py-2 text-left",
    "transition-colors duration-150 hover:bg-donna-accent-soft",
    "disabled:opacity-60",
  ].join(" "),
  check: [
    "grid h-5 w-5 shrink-0 place-items-center rounded border",
    "border-donna-border bg-donna-surface-2 text-transparent",
  ].join(" "),
  checkDone: "border-donna-accent bg-donna-accent text-donna-on-accent",
  labelText: "text-sm text-donna-text",
  labelDone: "text-donna-faint line-through",
  carried: [
    "ml-2 inline-flex translate-y-[-1px] items-center rounded-full",
    "border border-donna-border bg-donna-surface-2 px-2 py-0.5",
    "align-middle text-[10px] font-medium tracking-wide text-donna-muted",
    "no-underline",
  ].join(" "),
  empty: "py-2 text-sm text-donna-muted",
  state: "py-2 text-sm text-donna-muted",
  link: "text-sm text-donna-accent hover:underline",
} as const;
