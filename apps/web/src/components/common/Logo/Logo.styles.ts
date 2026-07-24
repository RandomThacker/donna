import type { LogoProps } from "./Logo.types";

export const logoStyles = {
  root: "group inline-flex items-center gap-2.5",
  mark: [
    "relative grid h-7 w-7 place-items-center rounded-full",
    "bg-gradient-to-br from-donna-accent-bright to-donna-accent-deep",
  ].join(" "),
  markCore: "h-2 w-2 rounded-full bg-donna-on-accent/80",
  word: "font-display italic leading-none tracking-tight text-donna-accent transition-colors duration-200 group-hover:text-donna-accent-bright",
  sizes: {
    sm: "text-2xl",
    lg: "text-3xl",
    hero: "text-6xl sm:text-7xl md:text-8xl leading-[0.9]",
  } satisfies Record<NonNullable<LogoProps["size"]>, string>,
} as const;
