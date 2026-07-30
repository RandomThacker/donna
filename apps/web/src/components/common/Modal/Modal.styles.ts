export const modalStyles = {
  root: "fixed inset-0 z-50 flex items-end justify-center p-4 sm:items-center",
  backdrop: [
    "absolute inset-0 bg-black/55 backdrop-blur-sm",
    "animate-[donna-modal-fade_180ms_ease-out]",
  ].join(" "),
  panel: [
    "relative z-10 w-full max-w-md overflow-hidden rounded-[1.75rem]",
    "border border-donna-border bg-donna-surface p-6 shadow-donna-card",
    "sm:p-8",
    "animate-[donna-modal-rise_220ms_ease-out]",
  ].join(" "),
  header: "pr-10",
  title: "font-display text-3xl tracking-tight text-donna-text",
  description: "mt-2 text-sm leading-relaxed text-donna-muted",
  body: "mt-6",
  close: [
    "absolute right-4 top-4 inline-flex h-9 w-9 items-center justify-center rounded-full",
    "text-donna-muted transition-colors duration-200",
    "hover:bg-donna-surface-2 hover:text-donna-text",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
} as const;
