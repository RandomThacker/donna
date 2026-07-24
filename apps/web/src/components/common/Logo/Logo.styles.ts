import type { LogoProps } from "./Logo.types";

export const logoStyles = {
  root: "group inline-flex items-center gap-2.5",
  mark: [
    "relative grid h-7 w-7 place-items-center rounded-full",
    "bg-gradient-to-br from-donna-copper-bright to-donna-copper-deep",
    "shadow-[0_0_18px_rgb(203_169_125_/_0.45)]",
  ].join(" "),
  markCore: "h-2 w-2 rounded-full bg-donna-ink/80",
  word: "font-display italic leading-none tracking-tight text-donna-copper transition-colors duration-200 group-hover:text-donna-copper-bright",
  sizes: {
    sm: "text-2xl",
    lg: "text-3xl",
    hero: "text-6xl sm:text-7xl md:text-8xl leading-[0.9]",
  } satisfies Record<NonNullable<LogoProps["size"]>, string>,
} as const;
