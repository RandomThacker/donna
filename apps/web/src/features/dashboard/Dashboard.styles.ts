export const dashboardStyles = {
  page: "relative flex h-dvh overflow-hidden bg-donna-bg text-donna-text",
  shell: [
    "relative z-10 grid h-full w-full",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
    "xl:grid-cols-[15rem_minmax(0,1fr)_23.5rem]",
  ].join(" "),
  workspace: "min-h-0 overflow-y-auto bg-donna-bg",
  workspaceInner: "mx-auto w-full max-w-5xl p-5 sm:p-6 lg:p-7",
  bento: "grid grid-cols-12 gap-4",
  phoneMobile: "mt-4 xl:hidden",
  phoneColumn: [
    "hidden min-h-0 items-center justify-center",
    "border-l border-donna-hairline bg-donna-surface xl:flex",
  ].join(" "),
} as const;
