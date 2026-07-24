export const sidebarStyles = {
  aside: [
    "hidden h-full flex-col border-r border-donna-hairline",
    "bg-donna-surface md:flex",
  ].join(" "),
  brand: "flex items-center gap-2.5 px-5 py-5",
  brandMark: [
    "grid h-8 w-8 shrink-0 place-items-center rounded-full",
    "bg-gradient-to-br from-donna-accent-bright to-donna-accent-deep",
  ].join(" "),
  brandCore: "h-2 w-2 rounded-full bg-donna-on-accent/80",
  brandWord: "font-display text-xl italic text-donna-accent",
  nav: "flex flex-1 flex-col gap-1 px-3",
  item: [
    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm",
    "text-donna-muted transition-colors duration-150",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  itemActive: "bg-donna-accent-soft text-donna-text",
  itemIcon: "shrink-0 text-current opacity-80",
  footer: "mt-auto space-y-2 p-3",
  themeToggle: [
    "flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm",
    "text-donna-muted transition-colors duration-150",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  profile: [
    "flex items-center gap-3 rounded-xl border border-donna-border",
    "bg-donna-surface-2 px-3 py-3",
  ].join(" "),
  avatar: [
    "grid h-9 w-9 shrink-0 place-items-center rounded-full",
    "bg-donna-elevated text-xs font-medium text-donna-accent",
  ].join(" "),
  profileMeta: "min-w-0",
  profileName: "truncate text-sm text-donna-text",
  profileHint: "truncate text-xs text-donna-faint",
} as const;
