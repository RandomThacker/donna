export const tagsPanelStyles = {
  root: "space-y-3 rounded-2xl border border-donna-hairline bg-donna-surface p-4",
  header: "flex items-center justify-between gap-2",
  title: "text-[11px] font-semibold uppercase tracking-[0.14em] text-donna-faint",
  clearBtn: "text-[11px] font-medium text-donna-accent hover:underline",
  tagList: "space-y-1.5",
  tagRow: "flex items-center justify-between gap-2",
  deleteBtn: [
    "grid h-7 w-7 shrink-0 place-items-center rounded-lg text-donna-faint",
    "transition-colors hover:bg-donna-accent-soft hover:text-donna-text",
    "disabled:opacity-50",
  ].join(" "),
  hint: "text-xs text-donna-muted",
  createForm: "space-y-2 border-t border-donna-hairline pt-3",
  nameInput: [
    "h-9 w-full rounded-xl border border-donna-border bg-donna-elevated/40",
    "px-3 text-sm text-donna-text placeholder:text-donna-muted/70",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  colors: "flex flex-wrap gap-1.5",
  colorBtn: "h-6 w-6 rounded-full border-2 border-transparent transition-transform",
  colorBtnOn: "scale-110 border-donna-text ring-2 ring-donna-accent/40",
  addBtn: [
    "inline-flex h-9 w-full items-center justify-center rounded-xl",
    "bg-donna-accent text-sm font-medium text-donna-on-accent",
    "hover:bg-donna-accent-bright disabled:opacity-60",
  ].join(" "),
} as const;
