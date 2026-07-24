export const pillStyles = {
  root: [
    "inline-flex items-center gap-2 rounded-full",
    "border border-donna-glass-border bg-donna-glass px-3.5 py-1.5 backdrop-blur-md",
    "font-sans text-[0.7rem] font-medium uppercase tracking-[0.22em] text-donna-copper",
  ].join(" "),
  dot: "relative flex h-1.5 w-1.5",
  dotPing: "absolute inline-flex h-full w-full animate-ping rounded-full bg-donna-copper/60",
  dotCore: "relative inline-flex h-1.5 w-1.5 rounded-full bg-donna-copper-bright",
} as const;
