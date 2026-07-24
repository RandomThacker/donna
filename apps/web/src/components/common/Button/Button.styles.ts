import type { ButtonSize, ButtonVariant } from "./Button.types";

export const buttonStyles = {
  base: [
    "group relative inline-flex items-center justify-center gap-2.5 overflow-hidden",
    "rounded-full font-sans font-semibold tracking-wide",
    "transition-[background-color,color,border-color,transform,box-shadow] duration-300 ease-out",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-copper active:scale-[0.97]",
  ].join(" "),
  variants: {
    primary: [
      "bg-gradient-to-b from-donna-copper-bright to-donna-copper text-donna-ink",
      "shadow-donna-cta hover:-translate-y-0.5",
      "hover:shadow-[0_22px_54px_-12px_rgb(203_169_125_/_0.62)]",
    ].join(" "),
    ghost: "bg-transparent text-donna-cream/80 hover:text-donna-copper-bright",
    outline: [
      "border border-donna-glass-border bg-donna-glass text-donna-cream backdrop-blur-md",
      "hover:-translate-y-0.5 hover:border-donna-copper/40 hover:text-donna-copper-bright",
    ].join(" "),
  } satisfies Record<ButtonVariant, string>,
  sizes: {
    md: "h-10 px-5 text-sm",
    lg: "h-13 px-7 text-[0.95rem]",
  } satisfies Record<ButtonSize, string>,
  shine: [
    "pointer-events-none absolute inset-0 -translate-x-full",
    "bg-gradient-to-r from-transparent via-white/35 to-transparent",
    "transition-transform duration-700 ease-out group-hover:translate-x-full",
  ].join(" "),
  label: "relative z-10 inline-flex items-center gap-2.5",
  iconRight: "transition-transform duration-300 group-hover:translate-x-0.5",
} as const;
