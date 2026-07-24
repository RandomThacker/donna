export const landingNavStyles = {
  header: "fixed inset-x-0 top-0 z-40 pt-4 sm:pt-5",
  bar: [
    "flex h-15 items-center justify-between gap-6 rounded-full px-4 pl-5 sm:px-5",
    "border border-donna-glass-border bg-donna-glass-strong backdrop-blur-xl",
    "shadow-[0_18px_50px_-24px_rgb(0_0_0_/_0.9)]",
  ].join(" "),
  links: "hidden items-center gap-8 md:flex",
  link: [
    "font-sans text-sm text-donna-muted transition-colors duration-200",
    "hover:text-donna-copper-bright",
  ].join(" "),
  actions: "flex items-center gap-1.5 sm:gap-2",
} as const;
