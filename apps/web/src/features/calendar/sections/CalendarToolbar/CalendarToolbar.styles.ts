export const toolbarStyles = {
  root: [
    "flex items-center gap-3 border-b border-donna-hairline",
    "bg-donna-surface/80 px-4 py-2.5 backdrop-blur-md sm:px-5",
  ].join(" "),
  left: "flex min-w-0 flex-1 items-center gap-2 sm:gap-3",
  navGroup: "flex shrink-0 items-center gap-0.5",
  iconBtn: [
    "grid h-8 w-8 place-items-center rounded-lg",
    "text-donna-muted transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-donna-accent/40",
  ].join(" "),
  todayBtn: [
    "h-8 rounded-lg border border-donna-border bg-donna-surface-2 px-2.5",
    "text-xs font-medium text-donna-text transition-colors sm:text-sm",
    "hover:bg-donna-accent-soft",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-donna-accent/40",
  ].join(" "),
  title: [
    "min-w-0 truncate font-display text-lg italic leading-none text-donna-text",
    "sm:text-xl",
  ].join(" "),
  right: "flex shrink-0 items-center gap-2",
  controls: [
    "flex items-center rounded-xl border border-donna-border",
    "bg-donna-surface-2 p-0.5",
  ].join(" "),
  viewSwitch: "flex items-center",
  viewBtn: [
    "rounded-lg px-2 py-1.5 text-xs font-medium text-donna-muted",
    "transition-colors sm:px-2.5 sm:text-[13px]",
    "hover:text-donna-text",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-donna-accent/40",
  ].join(" "),
  viewBtnActive: "bg-donna-elevated text-donna-text shadow-sm",
  divider: "mx-0.5 h-4 w-px shrink-0 bg-donna-border",
  syncBtn: [
    "grid h-8 w-8 place-items-center rounded-lg",
    "text-donna-muted transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
    "disabled:opacity-50",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-donna-accent/40",
  ].join(" "),
  syncBtnActive: "text-donna-accent",
  calendarsBtn: [
    "grid h-8 place-items-center rounded-lg border border-donna-border",
    "px-2.5 text-xs text-donna-muted lg:hidden",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  spin: "animate-spin",
} as const;
