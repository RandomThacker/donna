export const drawerStyles = {
  root: [
    "fixed inset-0 z-50 flex justify-end overflow-hidden",
    "max-md:items-end max-md:justify-center",
  ].join(" "),
  backdrop: "absolute inset-0 bg-black/45 backdrop-blur-[2px]",
  panel: [
    "relative flex h-full w-full max-w-[26rem] min-w-0 flex-col overflow-hidden",
    "border-l border-donna-hairline bg-donna-surface shadow-donna-card",
    "max-md:h-[min(90dvh,40rem)] max-md:max-w-none max-md:rounded-t-[1.5rem]",
    "max-md:border-l-0 max-md:border-t",
  ].join(" "),
  header: [
    "flex shrink-0 items-center gap-2 border-b border-donna-hairline px-4 py-3.5",
  ].join(" "),
  headerTitle: "min-w-0 flex-1 font-display text-2xl tracking-tight text-donna-text",
  iconBtn: [
    "grid h-9 w-9 shrink-0 place-items-center rounded-xl",
    "text-donna-muted transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  body: "min-h-0 flex-1 overflow-y-auto scrollbar-hidden",
} as const;
