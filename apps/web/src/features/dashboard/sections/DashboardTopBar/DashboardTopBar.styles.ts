export const topBarStyles = {
  root: [
    "sticky top-0 z-40 flex items-center gap-3 md:hidden",
    "border-b border-donna-hairline bg-donna-bg/85 backdrop-blur-xl",
    "px-4 pb-3 pt-[calc(0.75rem+env(safe-area-inset-top))]",
  ].join(" "),
  brand: "flex min-w-0 items-center gap-2",
  brandMark: [
    "grid h-8 w-8 shrink-0 place-items-center rounded-full",
    "bg-gradient-to-br from-donna-accent-bright to-donna-accent-deep",
  ].join(" "),
  brandCore: "h-2 w-2 rounded-full bg-donna-on-accent/80",
  brandWord: "font-display text-lg italic leading-none text-donna-accent",
  spacer: "flex-1",
  action: [
    "relative grid h-9 w-9 shrink-0 place-items-center rounded-full",
    "border border-donna-hairline bg-donna-surface text-donna-muted",
    "transition-transform duration-150 active:scale-95",
  ].join(" "),
  actionIcon: "h-[1.05rem] w-[1.05rem]",
  badge: [
    "absolute -right-0.5 -top-0.5 grid h-4 min-w-4 place-items-center rounded-full",
    "bg-donna-accent px-1 text-[0.55rem] font-semibold leading-none text-donna-on-accent",
  ].join(" "),
  avatar: [
    "grid h-9 w-9 shrink-0 place-items-center overflow-hidden rounded-full",
    "bg-donna-elevated text-[0.7rem] font-medium text-donna-accent",
    "transition-transform duration-150 active:scale-95",
  ].join(" "),
  avatarImage: "h-full w-full object-cover",
} as const;
