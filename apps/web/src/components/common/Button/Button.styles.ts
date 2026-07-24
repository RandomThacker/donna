import type { ButtonSize, ButtonVariant } from "./Button.types";

export const buttonStyles = {
  base: [
    "group relative inline-flex items-center justify-center gap-2.5 overflow-hidden",
    "rounded-full font-sans font-semibold tracking-wide",
    "transition-[background-color,color,border-color,transform,box-shadow] duration-300 ease-out",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent active:scale-[0.97]",
  ].join(" "),
  variants: {
    primary: [
      "bg-gradient-to-b from-donna-accent-bright to-donna-accent text-donna-on-accent",
      "shadow-donna-cta hover:-translate-y-0.5",
    ].join(" "),
    ghost: "bg-transparent text-donna-muted hover:text-donna-accent-bright",
    outline: [
      "border border-donna-border bg-donna-surface text-donna-text",
      "hover:-translate-y-0.5 hover:border-donna-accent/40 hover:text-donna-accent",
    ].join(" "),
  } satisfies Record<ButtonVariant, string>,
  sizes: {
    md: "h-10 px-5 text-sm",
    lg: "h-12 px-7 text-[0.95rem]",
  } satisfies Record<ButtonSize, string>,
  shine: [
    "pointer-events-none absolute inset-0 -translate-x-full",
    "bg-gradient-to-r from-transparent via-white/30 to-transparent",
    "transition-transform duration-700 ease-out group-hover:translate-x-full",
  ].join(" "),
  label: "relative z-10 inline-flex items-center gap-2.5",
  iconRight: "transition-transform duration-300 group-hover:translate-x-0.5",
} as const;
