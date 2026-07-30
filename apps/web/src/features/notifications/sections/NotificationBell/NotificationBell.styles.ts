export const bellStyles = {
  root: [
    "fixed z-[45] inline-flex h-10 w-10 items-center justify-center rounded-full",
    "border border-donna-border bg-donna-surface text-donna-text shadow-donna-card",
    "transition-colors hover:border-donna-accent/40 hover:bg-donna-surface-2",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
    "right-4 top-[max(1rem,env(safe-area-inset-top))]",
    "md:right-5 md:top-5",
  ].join(" "),
  badge: [
    "absolute -right-1 -top-1 grid min-w-[1.15rem] place-items-center rounded-full",
    "bg-donna-accent px-1 text-[0.65rem] font-semibold leading-none text-donna-on-accent",
    "h-[1.15rem]",
  ].join(" "),
} as const;
