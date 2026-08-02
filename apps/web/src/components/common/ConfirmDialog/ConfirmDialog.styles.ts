export const confirmDialogStyles = {
  root: "z-[80]",
  actions: "flex flex-wrap justify-end gap-2",
  cancel: [
    "rounded-full px-4 py-2 text-sm text-donna-muted",
    "transition-colors hover:text-donna-text",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
    "disabled:opacity-50",
  ].join(" "),
  confirm: [
    "rounded-full px-4 py-2 text-sm font-medium",
    "transition-colors disabled:opacity-50",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  confirmDefault:
    "bg-donna-accent text-donna-on-accent hover:bg-donna-accent-bright",
  confirmDanger:
    "bg-rose-500/90 text-white hover:bg-rose-500",
} as const;
