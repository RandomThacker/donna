export const heroArcStyles = {
  root: "relative mx-auto h-[20rem] w-full max-w-5xl overflow-hidden sm:h-[24rem] md:h-[30rem]",
  glow: [
    "absolute bottom-0 left-1/2 h-[65%] w-[75%] -translate-x-1/2 rounded-[50%] blur-2xl",
    "bg-[radial-gradient(circle_at_center,rgb(203_169_125_/_0.28)_0%,rgb(169_127_79_/_0.12)_40%,transparent_72%)]",
  ].join(" "),
  grid: [
    "absolute inset-0 opacity-[0.5]",
    "[mask-image:radial-gradient(circle_at_50%_120%,black,transparent_70%)]",
    "bg-[radial-gradient(rgb(203_169_125_/_0.16)_1px,transparent_1px)] [background-size:26px_26px]",
  ].join(" "),
  rings: "absolute inset-0",
  ringBase: "absolute bottom-0 left-1/2 aspect-square -translate-x-1/2 translate-y-1/2 rounded-full",
  ringItems: [
    "w-[16rem] border border-donna-copper/25 bg-[radial-gradient(circle_at_50%_0%,rgb(203_169_125_/_0.1),transparent_60%)] sm:w-[20rem]",
    "w-[26rem] border border-donna-copper/18 sm:w-[32rem]",
    "w-[38rem] border border-donna-copper/12 sm:w-[46rem]",
    "w-[52rem] border border-donna-copper/[0.08] sm:w-[62rem]",
  ],
  chipBase: [
    "absolute z-20 inline-flex items-center gap-2 rounded-full px-3.5 py-2",
    "border border-donna-glass-border bg-donna-glass-strong text-sm text-donna-cream backdrop-blur-md",
    "shadow-[0_12px_34px_-14px_rgb(0_0_0_/_0.85)] animate-donna-float",
  ].join(" "),
  chipIcon: "text-donna-copper-bright",
  chipPositions: [
    "left-[2%] top-[42%] hidden md:inline-flex [animation-delay:-2s] sm:left-[6%]",
    "left-1/2 top-[8%] -translate-x-1/2",
    "right-[2%] top-[42%] hidden md:inline-flex [animation-delay:-4s] sm:right-[6%]",
  ],
  cardBase: [
    "absolute z-20 hidden w-52 rounded-2xl p-4 md:block",
    "border border-donna-glass-border bg-donna-glass-strong backdrop-blur-xl shadow-donna-card",
    "animate-donna-float",
  ].join(" "),
  cardPositions: [
    "bottom-[8%] left-[1%] lg:left-[6%]",
    "bottom-[20%] right-[1%] [animation-delay:-3s] lg:right-[6%]",
  ],
  cardHead: "mb-3 flex items-center gap-2 text-donna-copper-bright",
  cardTitle: "text-[0.7rem] font-medium uppercase tracking-[0.16em] text-donna-muted",
  cardValue: "font-display text-2xl italic leading-none text-donna-cream",
  cardSub: "mt-1 text-xs text-donna-faint",
  cardChart: "mt-3",
} as const;
