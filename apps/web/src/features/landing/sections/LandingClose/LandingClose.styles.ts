export const landingCloseStyles = {
  section: "relative pb-16 pt-24 sm:pt-28",
  panel: [
    "relative overflow-hidden rounded-3xl border border-donna-glass-border",
    "bg-donna-glass px-8 py-16 text-center backdrop-blur-xl sm:px-16 sm:py-20",
    "shadow-donna-card",
  ].join(" "),
  panelGlow: [
    "pointer-events-none absolute left-1/2 top-[-40%] h-[60%] w-[70%] -translate-x-1/2 rounded-full",
    "bg-[radial-gradient(circle,rgb(203_169_125_/_0.22),transparent_70%)] blur-2xl",
  ].join(" "),
  inner: "relative z-10 mx-auto flex max-w-2xl flex-col items-center",
  eyebrow: "mb-6",
  title: "font-display text-4xl leading-[1.06] tracking-tight text-donna-cream sm:text-5xl md:text-6xl",
  body: "mt-5 max-w-xl text-lg leading-relaxed text-donna-muted",
  action: "mt-10",
  footer: [
    "mt-16 flex flex-col items-center justify-between gap-4 border-t border-donna-hairline pt-8",
    "text-sm text-donna-faint sm:flex-row",
  ].join(" "),
} as const;
