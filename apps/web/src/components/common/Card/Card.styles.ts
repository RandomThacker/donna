export const cardStyles = {
  base: [
    "group relative overflow-hidden rounded-2xl",
    "border border-donna-glass-border bg-donna-glass p-7 backdrop-blur-xl",
    "shadow-donna-card",
  ].join(" "),
  interactive: [
    "transition-[transform,border-color,box-shadow] duration-300 ease-out",
    "hover:-translate-y-1 hover:border-donna-copper/35",
    "hover:shadow-[0_30px_70px_-30px_rgb(203_169_125_/_0.35)]",
  ].join(" "),
  sheen: [
    "pointer-events-none absolute inset-x-0 -top-px h-px",
    "bg-gradient-to-r from-transparent via-donna-copper/50 to-transparent",
    "opacity-60",
  ].join(" "),
  glow: [
    "pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full",
    "bg-[radial-gradient(circle,rgb(203_169_125_/_0.18),transparent_70%)]",
    "opacity-0 blur-xl transition-opacity duration-300 group-hover:opacity-100",
  ].join(" "),
} as const;
