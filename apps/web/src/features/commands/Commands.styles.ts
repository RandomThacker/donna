export const commandsStyles = {
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
  atmosphere: [
    "pointer-events-none absolute inset-0 overflow-hidden",
    "bg-[radial-gradient(ellipse_80%_50%_at_20%_-10%,rgb(201_168_124/0.14),transparent_55%),",
    "radial-gradient(ellipse_60%_40%_at_90%_10%,rgb(226_199_154/0.08),transparent_50%)]",
  ].join(""),
  inner: [
    "relative mx-auto flex w-full max-w-3xl flex-col gap-10",
    "overflow-x-hidden px-5 py-6 sm:px-6 sm:py-8 lg:px-8 lg:py-10",
  ].join(" "),
  hero: "max-w-xl",
  eyebrow: [
    "mb-3 inline-flex items-center gap-2 text-[0.7rem] font-semibold uppercase tracking-[0.22em]",
    "text-donna-accent",
  ].join(" "),
  title: [
    "font-display text-[2.35rem] leading-[1.05] tracking-tight text-donna-text",
    "sm:text-5xl",
  ].join(" "),
  subtitle: "mt-3 max-w-md text-sm leading-relaxed text-donna-muted sm:text-[0.95rem]",
  ctaRow: "mt-5 flex flex-wrap items-center gap-3",
  primaryCta: [
    "inline-flex items-center gap-2 rounded-full bg-donna-accent px-4 py-2.5",
    "text-sm font-medium text-donna-on-accent",
    "transition-[transform,opacity] duration-200 hover:opacity-95 active:scale-[0.98]",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  secondaryHint: "text-xs text-donna-faint",
  list: "flex flex-col gap-4",
  guide: [
    "group relative overflow-hidden rounded-2xl",
    "border border-donna-hairline bg-donna-surface/80",
    "transition-[border-color,transform,background-color] duration-300",
    "hover:border-donna-accent/35 hover:bg-donna-surface",
    "motion-safe:hover:-translate-y-0.5",
    "motion-safe:animate-[donna-fade-up_0.5s_ease-out_both]",
  ].join(" "),
  guideGlow: [
    "pointer-events-none absolute -right-8 -top-10 h-28 w-28 rounded-full",
    "bg-donna-accent/10 blur-2xl opacity-0 transition-opacity duration-500",
    "group-hover:opacity-100",
  ].join(" "),
  guideHead: "relative flex items-start gap-3 px-5 pt-5 sm:px-6 sm:pt-6",
  iconWrap: [
    "mt-0.5 grid h-10 w-10 shrink-0 place-items-center rounded-xl",
    "bg-donna-accent-soft text-donna-accent ring-1 ring-donna-accent/20",
  ].join(" "),
  guideCopy: "min-w-0 flex-1",
  guideTitle: "font-display text-xl tracking-tight text-donna-text",
  guideBlurb: "mt-1 text-sm text-donna-muted",
  intentPill: [
    "shrink-0 rounded-full border border-donna-hairline bg-donna-bg/60",
    "px-2.5 py-1 font-mono text-[0.65rem] tracking-wide text-donna-faint",
  ].join(" "),
  examples: "relative mt-4 space-y-2 px-5 pb-5 sm:px-6 sm:pb-6",
  exampleRow: [
    "flex items-center gap-2 rounded-xl border border-transparent",
    "bg-donna-bg/50 px-3 py-2.5",
    "transition-[border-color,background-color] duration-150",
    "hover:border-donna-accent/25 hover:bg-donna-elevated/60",
  ].join(" "),
  phrase: [
    "min-w-0 flex-1 font-mono text-[0.8rem] leading-snug text-donna-text",
    "sm:text-[0.85rem]",
  ].join(" "),
  actions: "flex shrink-0 items-center gap-1",
  iconBtn: [
    "grid h-8 w-8 place-items-center rounded-full text-donna-muted",
    "transition-[color,background-color] duration-150",
    "hover:bg-donna-accent-soft hover:text-donna-accent",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
  footer: [
    "border-t border-donna-hairline pt-6 text-sm leading-relaxed text-donna-muted",
  ].join(" "),
  footerStrong: "text-donna-text",
} as const;
