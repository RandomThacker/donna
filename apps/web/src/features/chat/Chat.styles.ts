export const chatStyles = {
  page: [
    "relative flex h-dvh overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1 md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: "flex min-h-0 min-w-0 flex-col overflow-hidden bg-donna-bg",
  header: "shrink-0 border-b border-donna-hairline px-5 py-4 sm:px-6",
  title: "font-[family-name:var(--font-instrument-serif)] text-2xl tracking-tight",
  subtitle: "mt-1 text-sm text-donna-muted",
  thread: "flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-5 py-5 sm:px-6",
  rowUser: "flex justify-end",
  rowDonna: "flex justify-start",
  bubbleUser: [
    "max-w-[min(36rem,85%)] whitespace-pre-wrap rounded-2xl rounded-br-md",
    "bg-donna-text px-4 py-2.5 text-sm leading-relaxed text-donna-bg",
  ].join(" "),
  bubbleDonna: [
    "max-w-[min(36rem,85%)] whitespace-pre-wrap rounded-2xl rounded-bl-md",
    "bg-donna-surface px-4 py-2.5 text-sm leading-relaxed text-donna-text",
    "ring-1 ring-donna-hairline",
  ].join(" "),
  typing: "text-sm text-donna-muted",
  composer: [
    "shrink-0 border-t border-donna-hairline bg-donna-bg/90 px-5 py-3 backdrop-blur",
    "sm:px-6",
  ].join(" "),
  composerRow: "mx-auto flex w-full max-w-3xl items-end gap-2",
  input: [
    "min-h-11 flex-1 resize-none rounded-xl border border-donna-hairline",
    "bg-donna-surface px-3 py-2.5 text-sm text-donna-text outline-none",
    "placeholder:text-donna-muted focus:border-donna-text/30",
  ].join(" "),
  send: [
    "inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl",
    "bg-donna-text text-donna-bg transition hover:opacity-90 disabled:opacity-40",
  ].join(" "),
} as const;
