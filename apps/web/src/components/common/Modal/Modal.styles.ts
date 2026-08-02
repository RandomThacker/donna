export const modalStyles = {
  root: [
    // Above Donna FAB (z-55) and mobile bottom nav (z-60).
    "fixed inset-0 z-[80] flex items-end justify-center",
    "p-0 sm:items-center sm:p-4",
  ].join(" "),
  backdrop: [
    "absolute inset-0 bg-black/55 backdrop-blur-sm",
    "animate-[donna-modal-fade_180ms_ease-out]",
  ].join(" "),
  panel: [
    "relative z-10 flex w-full max-h-[min(92dvh,44rem)] flex-col overflow-hidden",
    // Mobile: edge-to-edge bottom sheet. Desktop: centered card.
    "rounded-t-[1.75rem] border border-donna-border border-b-0 bg-donna-surface shadow-donna-card",
    "px-5 pt-5 pb-[max(1.25rem,env(safe-area-inset-bottom))]",
    "sm:max-w-md sm:rounded-[1.75rem] sm:border-b sm:p-8",
    "animate-[donna-modal-rise_220ms_ease-out]",
  ].join(" "),
  header: "shrink-0 pr-10",
  title: [
    "font-display text-[1.85rem] leading-tight tracking-tight text-donna-text",
    "sm:text-3xl",
  ].join(" "),
  description: "mt-1.5 text-sm leading-relaxed text-donna-muted sm:mt-2",
  body: [
    "mt-4 min-h-0 flex-1 overflow-y-auto overscroll-contain",
    "-mx-1 px-1 sm:mt-6",
  ].join(" "),
  close: [
    "absolute right-3 top-3 inline-flex h-10 w-10 items-center justify-center rounded-full",
    "text-donna-muted transition-colors duration-200 sm:right-4 sm:top-4 sm:h-9 sm:w-9",
    "hover:bg-donna-surface-2 hover:text-donna-text",
    "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2",
    "focus-visible:outline-donna-accent",
  ].join(" "),
} as const;
