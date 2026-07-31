export const bottomBarStyles = {
  root: [
    // Above the Donna FAB (z-55) so left tabs like Home stay tappable.
    "fixed inset-x-0 bottom-0 z-[60] overflow-x-hidden md:hidden",
    "border-t border-donna-hairline bg-donna-surface/95 backdrop-blur-xl",
    "shadow-[0_-10px_28px_-20px_rgb(0_0_0_/_0.65)]",
    "pb-[env(safe-area-inset-bottom)]",
  ].join(" "),
  nav: "mx-auto grid h-[4.25rem] w-full max-w-lg grid-cols-4",
  item: [
    "relative z-10 flex h-full min-w-0 w-full flex-col items-center justify-center gap-1",
    "px-1 text-[0.625rem] tracking-wide text-donna-faint",
    "transition-[color,transform] duration-150 active:scale-95",
    // Full-cell hit target for touch.
    "touch-manipulation",
  ].join(" "),
  itemActive: "text-donna-accent",
  // Blinkit/Instamart-style tab marker riding the top edge.
  indicator: [
    "pointer-events-none absolute left-1/2 top-0 h-[3px] w-8 -translate-x-1/2",
    "rounded-b-full bg-donna-accent",
  ].join(" "),
  iconWrap: [
    "relative grid h-9 w-9 place-items-center rounded-2xl",
    "transition-[background-color,color] duration-150",
  ].join(" "),
  iconWrapActive: "bg-donna-accent-soft text-donna-accent",
  icon: "pointer-events-none h-[1.2rem] w-[1.2rem]",
  badge: [
    "absolute -right-0.5 -top-0.5 grid h-4 min-w-4 place-items-center rounded-full",
    "bg-donna-accent px-1 text-[0.55rem] font-semibold leading-none text-donna-on-accent",
  ].join(" "),
  label: "max-w-full truncate font-medium leading-none",
  labelActive: "font-semibold",
} as const;
