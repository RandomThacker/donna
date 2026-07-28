export const authStyles = {
  page: "relative flex min-h-dvh items-center justify-center overflow-hidden bg-donna-bg px-6 py-16 text-donna-ink",
  glow: "pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,color-mix(in_oklab,var(--donna-accent)_22%,transparent),transparent_55%)]",
  card: "relative w-full max-w-md rounded-[1.75rem] border border-donna-border bg-donna-elevated/80 p-8 shadow-[0_24px_80px_rgba(0,0,0,0.28)] backdrop-blur-xl",
  brand: "mb-4 flex flex-col items-start gap-3",
  title: "font-[family-name:var(--font-instrument-serif)] text-4xl tracking-tight text-donna-text",
  body: "mt-3 text-sm leading-relaxed text-donna-muted",
  actions: "mt-8 flex flex-col gap-3",
  note: "mt-4 text-center text-xs text-donna-muted",
  error: "mt-4 rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200",
  status: "mt-6 text-sm text-donna-muted",
} as const;
