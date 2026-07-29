export const bottomBarStyles = {
  root: [
    "fixed inset-x-0 bottom-0 z-50 overflow-x-hidden md:hidden",
    "border-t border-donna-hairline bg-donna-surface",
    "pb-[env(safe-area-inset-bottom)]",
  ].join(" "),
  nav: "mx-auto grid h-[3.75rem] w-full max-w-lg grid-cols-5",
  item: [
    "flex min-w-0 flex-col items-center justify-center gap-0.5 px-1",
    "text-[10px] font-medium tracking-wide text-donna-faint",
    "transition-colors duration-150",
    "active:opacity-80",
  ].join(" "),
  itemActive: "text-donna-accent",
  iconWrap: "grid h-8 w-8 place-items-center rounded-xl",
  iconWrapActive: "bg-donna-accent-soft text-donna-accent",
  icon: "h-[1.15rem] w-[1.15rem]",
  label: "truncate",
} as const;
