export const landingHeroStyles = {
  section: "relative flex min-h-dvh flex-col items-center pt-32 sm:pt-36",
  copy: "relative z-10 mx-auto flex max-w-3xl flex-col items-center text-center",
  eyebrow: "animate-donna-fade-up",
  headline: [
    "mt-7 font-display text-5xl leading-[1.04] tracking-tight text-donna-cream",
    "sm:text-6xl md:text-7xl",
    "animate-donna-fade-up [animation-delay:80ms]",
  ].join(" "),
  emphasis: "block italic text-donna-gradient animate-donna-shimmer",
  support: [
    "mt-6 max-w-xl text-lg leading-relaxed text-donna-muted",
    "animate-donna-fade-up [animation-delay:180ms]",
  ].join(" "),
  actions: [
    "mt-9 flex flex-col items-center gap-3.5 sm:flex-row sm:gap-4",
    "animate-donna-fade-up [animation-delay:260ms]",
  ].join(" "),
  note: "mt-4 text-sm text-donna-faint animate-donna-fade-up [animation-delay:320ms]",
  visual: "relative mt-4 w-full animate-donna-fade [animation-delay:200ms] sm:mt-2",
} as const;
