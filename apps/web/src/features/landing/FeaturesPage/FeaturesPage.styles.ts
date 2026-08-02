export const featuresPageStyles = {
  page: "relative min-h-dvh overflow-x-hidden bg-donna-void text-donna-cream",
  main: "relative z-10 pb-20 pt-28 sm:pb-28 sm:pt-36",
  stage: "relative mx-auto w-full max-w-6xl px-5 sm:px-8",

  grid: [
    "pointer-events-none absolute inset-x-0 top-0 h-[60rem] opacity-[0.25]",
    "bg-[linear-gradient(rgb(255_255_255/0.035)_1px,transparent_1px),linear-gradient(90deg,rgb(255_255_255/0.035)_1px,transparent_1px)]",
    "bg-[size:3.5rem_3.5rem]",
    "[mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,black,transparent_78%)]",
  ].join(" "),
  glowA: [
    "pointer-events-none absolute left-1/2 top-[-6rem] h-[26rem] w-[26rem] -translate-x-1/2",
    "rounded-full bg-donna-accent/16 blur-[130px] motion-safe:animate-donna-float",
  ].join(" "),
  glowB: [
    "pointer-events-none absolute right-[-6%] top-[38%] h-[22rem] w-[22rem] rounded-full",
    "bg-donna-accent-deep/20 blur-[130px] motion-safe:animate-donna-drift",
  ].join(" "),

  hero: "relative mx-auto max-w-3xl text-center motion-safe:animate-donna-fade-up",
  badge: [
    "inline-flex items-center gap-2 rounded-full border border-donna-accent/35",
    "bg-donna-accent-soft px-3.5 py-1.5 text-[10px] font-medium uppercase tracking-[0.24em]",
    "text-donna-accent",
  ].join(" "),
  badgeDot: "h-1.5 w-1.5 rounded-full bg-donna-accent motion-safe:animate-pulse",
  headline: [
    "mt-7 font-display text-[2.75rem] leading-[1.02] tracking-tight text-donna-text",
    "sm:text-6xl lg:text-[4.25rem]",
  ].join(" "),
  emphasis: "block italic text-donna-gradient motion-safe:animate-donna-shimmer",
  subhead: [
    "mx-auto mt-6 max-w-2xl text-base leading-relaxed text-donna-muted sm:text-lg",
  ].join(" "),

  pillars: [
    "relative mx-auto mt-14 grid max-w-4xl gap-px overflow-hidden rounded-3xl",
    "border border-donna-hairline bg-donna-hairline/60 sm:mt-16 sm:grid-cols-3",
  ].join(" "),
  pillar: "bg-donna-void/80 px-6 py-6 backdrop-blur-sm",
  pillarTitle: "font-display text-lg tracking-tight text-donna-text",
  pillarBody: "mt-2 text-sm leading-relaxed text-donna-muted",

  section: "relative mt-24 sm:mt-32",
  sectionHead: "max-w-2xl",
  sectionHeadCentered: "mx-auto max-w-2xl text-center",
  sectionEyebrow: [
    "inline-flex items-center gap-2 text-[10px] font-semibold uppercase",
    "tracking-[0.24em] text-donna-accent",
  ].join(" "),
  sectionRule: "h-px w-8 bg-donna-accent/45",
  sectionTitle: [
    "mt-4 font-display text-3xl leading-tight tracking-tight text-donna-text",
    "sm:text-[2.6rem]",
  ].join(" "),
  sectionDesc: "mt-4 text-sm leading-relaxed text-donna-muted sm:text-base",

  whoPanel: [
    "relative mt-10 overflow-hidden rounded-[2rem] border border-donna-border",
    "bg-donna-surface/70 p-7 shadow-donna-card backdrop-blur-md sm:p-10",
  ].join(" "),
  whoGlow: [
    "pointer-events-none absolute -inset-px rounded-[2rem]",
    "bg-[linear-gradient(135deg,rgb(201_168_124/0.28),transparent_42%,rgb(201_168_124/0.1))]",
  ].join(" "),
  whoBody: [
    "relative space-y-5 text-base leading-relaxed text-donna-muted",
    "sm:text-lg",
  ].join(" "),
  whoLead: "text-donna-text",

  highlights: "mt-12 grid gap-4 lg:grid-cols-2",
  highlight: [
    "group relative flex h-full flex-col overflow-hidden rounded-[1.75rem]",
    "border border-donna-border bg-donna-surface/60 p-6 backdrop-blur-md sm:p-8",
    "shadow-[0_24px_70px_-40px_rgb(0_0_0/0.95)]",
    "transition-[transform,border-color] duration-500",
    "hover:border-donna-accent/35 motion-safe:hover:-translate-y-1",
    "motion-safe:animate-donna-fade-up",
  ].join(" "),
  highlightGlow: [
    "pointer-events-none absolute -right-16 -top-16 h-48 w-48 rounded-full",
    "bg-donna-accent/12 blur-3xl opacity-0 transition-opacity duration-500",
    "group-hover:opacity-100",
  ].join(" "),
  highlightIndex: [
    "pointer-events-none absolute right-6 top-5 font-display text-5xl leading-none",
    "text-donna-text/[0.06] transition-colors duration-500",
    "group-hover:text-donna-accent/15 sm:text-6xl",
  ].join(" "),
  highlightIcon: [
    "relative grid h-11 w-11 place-items-center rounded-2xl",
    "bg-donna-accent-soft text-donna-accent ring-1 ring-donna-accent/25",
  ].join(" "),
  highlightKicker: [
    "relative mt-5 text-[10px] font-semibold uppercase tracking-[0.2em]",
    "text-donna-faint",
  ].join(" "),
  highlightTitle: [
    "relative mt-2 font-display text-2xl leading-snug tracking-tight text-donna-text",
    "sm:text-[1.7rem]",
  ].join(" "),
  highlightBody: "relative mt-3 text-sm leading-relaxed text-donna-muted sm:text-[0.95rem]",
  highlightPoints: "relative mt-6 space-y-2.5 border-t border-donna-hairline pt-5",
  highlightPoint: "flex items-start gap-2.5 text-sm text-donna-muted",
  highlightBullet: [
    "mt-[0.45rem] h-1.5 w-1.5 shrink-0 rounded-full bg-donna-accent/70",
  ].join(" "),

  coreGrid: "mt-12 grid gap-px overflow-hidden rounded-[1.75rem] border border-donna-hairline bg-donna-hairline/60 sm:grid-cols-2 lg:grid-cols-3",
  coreCard: [
    "group relative bg-donna-void/85 p-6 transition-colors duration-300",
    "hover:bg-donna-surface/70",
  ].join(" "),
  coreTop: "flex items-center gap-3",
  coreIcon: [
    "grid h-9 w-9 shrink-0 place-items-center rounded-xl",
    "bg-donna-accent-soft/70 text-donna-accent ring-1 ring-donna-accent/15",
    "transition-colors duration-300 group-hover:bg-donna-accent-soft",
  ].join(" "),
  coreTitle: "font-medium text-donna-text",
  coreBlurb: "mt-3 text-sm leading-relaxed text-donna-muted",

  horizon: [
    "mt-12 flex flex-col items-center gap-4 rounded-3xl border border-donna-hairline",
    "bg-donna-surface/40 px-6 py-6 text-center backdrop-blur-sm sm:flex-row sm:justify-center sm:gap-5 sm:text-left",
  ].join(" "),
  horizonLabel: [
    "text-[10px] font-semibold uppercase tracking-[0.22em] text-donna-faint",
  ].join(" "),
  horizonList: "flex flex-wrap items-center justify-center gap-2",
  horizonPill: [
    "inline-flex items-center gap-2 rounded-full border border-donna-hairline",
    "bg-donna-surface-2/70 px-3 py-1.5 text-xs text-donna-muted",
  ].join(" "),
  horizonDot: "h-1.5 w-1.5 rounded-full bg-donna-accent/60",

  close: [
    "relative mt-24 overflow-hidden rounded-[2rem] border border-donna-border",
    "bg-donna-surface/75 px-6 py-12 text-center shadow-donna-card backdrop-blur-md",
    "sm:mt-32 sm:px-10 sm:py-16",
  ].join(" "),
  closeGlow: [
    "pointer-events-none absolute -inset-px rounded-[2rem]",
    "bg-[radial-gradient(ellipse_70%_100%_at_50%_0%,rgb(201_168_124/0.22),transparent_65%)]",
  ].join(" "),
  closeTitle: [
    "relative mt-4 font-display text-3xl tracking-tight text-donna-text sm:text-[2.75rem]",
  ].join(" "),
  closeBody: [
    "relative mx-auto mt-4 max-w-lg text-sm leading-relaxed text-donna-muted sm:text-base",
  ].join(" "),
  closeActions: "relative mt-9 flex flex-wrap items-center justify-center gap-3",
  secondaryBtn: [
    "inline-flex h-11 items-center justify-center rounded-full",
    "border border-donna-border px-6 text-sm text-donna-muted",
    "transition hover:border-donna-accent/40 hover:text-donna-text",
  ].join(" "),
  footnote: "relative mt-7 text-xs text-donna-faint",

  footer: [
    "relative mt-16 flex flex-col items-center gap-3 border-t border-donna-hairline",
    "pt-8 text-center text-xs text-donna-faint sm:flex-row sm:justify-between sm:text-left",
  ].join(" "),
} as const;
