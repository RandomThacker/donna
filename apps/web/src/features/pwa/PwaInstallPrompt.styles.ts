export const installPromptStyles = {
  banner: [
    "fixed inset-x-4 bottom-[calc(4.75rem+env(safe-area-inset-bottom))] z-[70]",
    "rounded-2xl border border-donna-border bg-donna-surface/95 p-4 shadow-donna-card",
    "backdrop-blur-md md:bottom-6 md:left-auto md:right-6 md:max-w-sm md:inset-x-auto",
    "animate-donna-fade-up",
  ].join(" "),
  row: "flex items-start gap-3",
  mark: [
    "mt-0.5 grid h-11 w-11 shrink-0 place-items-center rounded-full",
    "bg-gradient-to-br from-donna-accent-bright to-donna-accent-deep",
    "[mask-image:radial-gradient(farthest-side,transparent_34%,#000_35%)]",
    "[-webkit-mask-image:radial-gradient(farthest-side,transparent_34%,#000_35%)]",
  ].join(" "),
  copy: "min-w-0 flex-1",
  title: "text-sm font-medium text-donna-text",
  body: "mt-1 text-xs leading-relaxed text-donna-muted",
  actions: "mt-3 flex items-center gap-2",
  installBtn: [
    "inline-flex h-9 flex-1 items-center justify-center rounded-full",
    "bg-donna-accent px-4 text-sm font-medium text-donna-on-accent",
    "hover:bg-donna-accent-bright",
  ].join(" "),
  dismissBtn: [
    "inline-flex h-9 items-center justify-center rounded-full px-3",
    "text-sm text-donna-muted hover:text-donna-text",
  ].join(" "),
} as const;
