export const tagPickerStyles = {
  root: "relative",
  trigger: [
    "mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-lg",
    "text-donna-faint transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
    "disabled:opacity-40",
  ].join(" "),
  menu: [
    "absolute right-0 top-full z-20 mt-1 w-44 rounded-xl",
    "border border-donna-border bg-donna-surface p-1 shadow-lg",
  ].join(" "),
  empty: "px-2 py-2 text-xs text-donna-muted",
  list: "max-h-48 space-y-0.5 overflow-y-auto scrollbar-hidden",
  option: [
    "flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-sm",
    "text-donna-text transition-colors hover:bg-donna-accent-soft",
  ].join(" "),
  optionOn: "bg-donna-accent-soft/60",
  swatch: "h-2.5 w-2.5 shrink-0 rounded-full",
  optionName: "min-w-0 flex-1 truncate",
} as const;
