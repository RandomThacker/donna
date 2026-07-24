export const quickTasksStyles = {
  box: "col-span-12 md:col-span-6",
  list: "space-y-0.5",
  item: [
    "flex w-full items-center gap-3 rounded-lg px-1.5 py-2 text-left",
    "transition-colors duration-150 hover:bg-donna-accent-soft",
  ].join(" "),
  check: [
    "grid h-5 w-5 shrink-0 place-items-center rounded-full border",
    "border-donna-accent/40 text-transparent",
  ].join(" "),
  checkDone: "border-donna-accent bg-donna-accent-soft text-donna-accent",
  labelText: "text-sm text-donna-text",
  labelDone: "text-donna-faint line-through",
} as const;
