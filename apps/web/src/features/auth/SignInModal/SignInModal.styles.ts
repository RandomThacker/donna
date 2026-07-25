export const signInModalStyles = {
  list: "flex flex-col gap-3",
  provider: [
    "group relative inline-flex w-full items-center justify-center gap-2.5 overflow-hidden",
    "h-12 rounded-full border border-donna-border bg-donna-surface px-7",
    "font-sans text-[0.95rem] font-semibold tracking-wide text-donna-text",
    "transition-[background-color,color,border-color,transform] duration-300 ease-out",
    "hover:-translate-y-0.5 hover:border-donna-accent/40 hover:text-donna-accent",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent active:scale-[0.97]",
    "disabled:pointer-events-none disabled:opacity-50",
  ].join(" "),
  providerLabel: "relative z-10 inline-flex items-center gap-2.5",
  note: "mt-5 text-center text-xs leading-relaxed text-donna-muted",
} as const;
