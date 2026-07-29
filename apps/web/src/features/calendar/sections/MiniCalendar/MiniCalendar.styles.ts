export const miniCalendarStyles = {
  root: "rounded-2xl border border-donna-border bg-donna-surface-2 p-3",
  header: "mb-2 flex items-center justify-between",
  month: "text-sm font-medium text-donna-text",
  nav: "flex gap-1",
  navBtn: [
    "grid h-7 w-7 place-items-center rounded-lg text-donna-muted",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  weekdays: "mb-1 grid grid-cols-7 gap-0.5 text-center text-[10px] text-donna-faint",
  grid: "grid grid-cols-7 gap-0.5",
  day: [
    "flex aspect-square flex-col items-center justify-center gap-0.5 rounded-lg",
    "text-xs text-donna-muted transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  dayMuted: "text-donna-faint/70",
  dayToday: "font-semibold text-donna-accent",
  daySelected: [
    "bg-donna-accent text-donna-on-accent",
    "hover:bg-donna-accent-bright hover:text-donna-on-accent",
  ].join(" "),
  dayExtra: "text-[9px] leading-none opacity-80",
} as const;
