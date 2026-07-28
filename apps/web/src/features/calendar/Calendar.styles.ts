export const calendarStyles = {
  page: "relative flex h-dvh overflow-hidden bg-donna-bg text-donna-text",
  shell: [
    "relative z-10 grid h-full w-full",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: "flex min-h-0 min-w-0 flex-col overflow-hidden bg-donna-bg",
  body: "relative flex min-h-0 flex-1 overflow-hidden",
  main: "flex min-h-0 min-w-0 flex-1 flex-col",
  calSidebar: [
    "hidden w-[17.5rem] shrink-0 flex-col gap-5 overflow-y-auto",
    "border-r border-donna-hairline bg-donna-surface p-4 lg:flex",
  ].join(" "),
  calSidebarMobile: [
    "absolute inset-y-0 left-0 z-30 flex w-[17.5rem] flex-col gap-5",
    "overflow-y-auto border-r border-donna-hairline bg-donna-surface p-4",
    "shadow-donna-card lg:hidden",
    "animate-donna-fade-up lg:hidden",
  ].join(" "),
  mobileBackdrop: "absolute inset-0 z-20 bg-black/40 lg:hidden",
  viewPane: "min-h-0 min-w-0 flex-1 overflow-x-hidden overflow-y-auto",
  syncBar: "h-0.5 w-full overflow-hidden bg-donna-hairline",
  syncBarFill: "h-full w-1/3 animate-pulse bg-donna-accent",
} as const;
