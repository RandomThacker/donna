export const memoriesStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1 md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: [
    "relative min-h-0 min-w-0 overflow-x-hidden overflow-y-auto scrollbar-hidden",
    "bg-donna-bg",
  ].join(" "),
  stage: [
    "relative mx-auto flex min-h-[calc(100dvh-5rem)] w-full max-w-4xl",
    "flex-col items-center justify-center px-5 py-10 sm:px-8 sm:py-14",
  ].join(" "),
  glowA: [
    "pointer-events-none absolute left-[8%] top-[12%] h-56 w-56 rounded-full",
    "bg-donna-accent/20 blur-[90px] animate-donna-drift",
  ].join(" "),
  glowB: [
    "pointer-events-none absolute bottom-[10%] right-[6%] h-72 w-72 rounded-full",
    "bg-donna-accent-deep/25 blur-[110px] animate-donna-float",
  ].join(" "),
  grid: [
    "pointer-events-none absolute inset-0 opacity-[0.35]",
    "bg-[linear-gradient(rgb(255_255_255/0.03)_1px,transparent_1px),linear-gradient(90deg,rgb(255_255_255/0.03)_1px,transparent_1px)]",
    "bg-[size:3rem_3rem]",
    "[mask-image:radial-gradient(ellipse_at_center,black,transparent_72%)]",
  ].join(" "),
  panel: [
    "relative w-full overflow-hidden rounded-[2rem] border border-donna-border",
    "bg-donna-surface/80 p-6 shadow-donna-card backdrop-blur-md sm:p-10",
    "animate-donna-fade-up",
  ].join(" "),
  panelGlow: [
    "pointer-events-none absolute -inset-px rounded-[2rem]",
    "bg-[linear-gradient(135deg,rgb(201_168_124/0.35),transparent_45%,rgb(201_168_124/0.12))]",
  ].join(" "),
  badge: [
    "mb-5 inline-flex items-center gap-2 rounded-full border border-donna-accent/35",
    "bg-donna-accent-soft px-3 py-1 text-[11px] font-medium uppercase tracking-[0.2em]",
    "text-donna-accent",
  ].join(" "),
  badgeDot: "h-1.5 w-1.5 rounded-full bg-donna-accent animate-pulse",
  eyebrow: "text-[11px] font-semibold uppercase tracking-[0.22em] text-donna-faint",
  headline: [
    "mt-3 font-display text-4xl leading-[1.05] tracking-tight text-donna-text",
    "sm:text-5xl lg:text-6xl",
  ].join(" "),
  emphasis: "italic text-donna-gradient animate-donna-shimmer",
  subhead: "mt-4 max-w-2xl text-base leading-relaxed text-donna-muted sm:text-lg",
  status: [
    "mt-6 inline-flex items-center gap-2 rounded-xl border border-donna-hairline",
    "bg-donna-surface-2/80 px-4 py-2.5 text-sm text-donna-muted",
  ].join(" "),
  cards: "relative mt-10 grid gap-3 sm:grid-cols-2",
  card: [
    "rounded-2xl border border-donna-border bg-donna-elevated/70 p-4",
    "shadow-[0_12px_40px_-24px_rgb(0_0_0/0.9)] transition-transform duration-300",
    "hover:-translate-y-0.5 hover:border-donna-accent/30",
  ].join(" "),
  cardTag: "text-[10px] font-medium uppercase tracking-[0.16em] text-donna-faint",
  cardTitle: "mt-1 font-medium text-donna-text",
  cardSnippet: "mt-1 text-sm leading-relaxed text-donna-muted",
  actions: "mt-10 flex flex-wrap items-center justify-center gap-3",
  primaryBtn: [
    "inline-flex h-11 items-center justify-center rounded-full",
    "bg-donna-accent px-6 text-sm font-medium text-donna-on-accent",
    "transition hover:bg-donna-accent-bright",
  ].join(" "),
  secondaryBtn: [
    "inline-flex h-11 items-center justify-center rounded-full",
    "border border-donna-border px-6 text-sm text-donna-muted",
    "transition hover:border-donna-accent/40 hover:text-donna-text",
  ].join(" "),
  footnote: "mt-6 text-center text-xs text-donna-faint",
} as const;
