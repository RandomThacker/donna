export const notificationsPageStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1 md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: [
    "relative flex min-h-0 min-w-0 flex-col overflow-hidden bg-donna-bg",
  ].join(" "),
  header: [
    "flex shrink-0 items-center justify-between gap-3 border-b border-donna-hairline",
    "bg-donna-surface/80 px-5 py-4 backdrop-blur-md sm:px-6",
  ].join(" "),
  title: "font-display text-2xl tracking-tight text-donna-text sm:text-3xl",
  subtitle: "mt-0.5 text-sm text-donna-muted",
  badge: [
    "inline-flex h-6 min-w-6 items-center justify-center rounded-full",
    "bg-donna-accent px-2 text-xs font-semibold text-donna-on-accent",
  ].join(" "),
  main: [
    "grid min-h-0 flex-1 grid-cols-1 overflow-hidden",
    "lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]",
  ].join(" "),
  listPane: [
    "min-h-0 overflow-hidden border-donna-hairline",
    "lg:border-r",
    "bg-donna-surface",
  ].join(" "),
  detailsPane: [
    "hidden min-h-0 overflow-y-auto bg-donna-bg lg:block",
  ].join(" "),
  detailsEmpty: [
    "flex h-full items-center justify-center px-6 text-center text-sm text-donna-muted",
  ].join(" "),
  mobileDetails: [
    "fixed inset-0 z-40 flex flex-col bg-donna-surface lg:hidden",
    "pb-[calc(4rem+env(safe-area-inset-bottom))]",
  ].join(" "),
  mobileDetailsHeader: [
    "flex shrink-0 items-center gap-2 border-b border-donna-hairline px-3 py-3",
  ].join(" "),
  iconBtn: [
    "grid h-9 w-9 shrink-0 place-items-center rounded-xl",
    "text-donna-muted transition-colors",
    "hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  mobileDetailsTitle: "min-w-0 flex-1 font-display text-xl text-donna-text",
  mobileDetailsBody: "min-h-0 flex-1 overflow-y-auto",
} as const;
