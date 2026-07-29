export const dashboardStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
    "xl:grid-cols-[15rem_minmax(0,1fr)_23.5rem]",
  ].join(" "),
  workspace: [
    "min-h-0 min-w-0 overflow-x-hidden overflow-y-auto scrollbar-hidden",
    "bg-donna-bg",
  ].join(" "),
  workspaceInner: "mx-auto w-full max-w-5xl overflow-x-hidden p-5 sm:p-6 lg:p-7",
  bento: "grid grid-cols-12 gap-4",
  phoneColumn: [
    "hidden min-h-0 min-w-0 items-center justify-center overflow-x-hidden",
    "border-l border-donna-hairline bg-donna-surface xl:flex",
  ].join(" "),
} as const;
