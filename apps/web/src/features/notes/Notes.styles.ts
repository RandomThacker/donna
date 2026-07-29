export const noteColorClass: Record<string, string> = {
  default: "bg-donna-surface border-donna-hairline",
  coral: "bg-[#3a2422] border-[#6a403c]",
  sage: "bg-[#243028] border-[#3f5346]",
  sky: "bg-[#222c38] border-[#3a4d63]",
  blush: "bg-[#362430] border-[#5c3f52]",
  sand: "bg-[#342e22] border-[#5a4f38]",
  lilac: "bg-[#2c2438] border-[#4a3d5e]",
};

export const noteColorDot: Record<string, string> = {
  default: "bg-donna-surface-2 ring-1 ring-donna-border",
  coral: "bg-[#c4786a]",
  sage: "bg-[#7aa887]",
  sky: "bg-[#6d92b8]",
  blush: "bg-[#b87a9a]",
  sand: "bg-[#c4a46a]",
  lilac: "bg-[#9a82c4]",
};

export const notesStyles = {
  page: [
    "relative flex h-dvh overflow-x-hidden overflow-hidden bg-donna-bg text-donna-text",
    "pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0",
  ].join(" "),
  shell: [
    "relative z-10 grid h-full w-full min-w-0",
    "grid-cols-1",
    "md:grid-cols-[15rem_minmax(0,1fr)]",
  ].join(" "),
  workspace: [
    "min-h-0 min-w-0 overflow-x-hidden overflow-y-auto scrollbar-hidden",
    "bg-donna-bg",
  ].join(" "),
  workspaceInner: "mx-auto w-full max-w-6xl overflow-x-hidden p-5 sm:p-6 lg:p-8",
  header: "mb-6",
  title: "font-display text-3xl tracking-tight text-donna-text sm:text-4xl",
  subtitle: "mt-1 text-sm text-donna-muted",
  composer: [
    "mx-auto mb-8 w-full max-w-xl rounded-2xl border border-donna-border",
    "bg-donna-surface p-3 shadow-[0_8px_30px_rgba(0,0,0,0.18)]",
  ].join(" "),
  composerTitle: [
    "mb-1 w-full bg-transparent text-sm font-medium text-donna-text",
    "placeholder:text-donna-muted/70 focus:outline-none",
  ].join(" "),
  composerBody: [
    "min-h-[2.5rem] w-full resize-none bg-transparent text-sm text-donna-text",
    "placeholder:text-donna-muted/70 focus:outline-none",
  ].join(" "),
  composerFooter: "mt-2 flex items-center justify-between gap-2",
  colorRow: "flex items-center gap-1.5",
  colorDot: "h-5 w-5 rounded-full transition-transform hover:scale-110",
  colorDotActive: "ring-2 ring-donna-accent ring-offset-1 ring-offset-donna-surface",
  composerActions: "flex items-center gap-2",
  ghostBtn: [
    "inline-flex h-8 items-center justify-center rounded-lg px-2.5",
    "text-sm text-donna-muted transition-colors hover:bg-donna-accent-soft hover:text-donna-text",
  ].join(" "),
  primaryBtn: [
    "inline-flex h-8 items-center justify-center rounded-lg",
    "bg-donna-accent px-3 text-sm font-medium text-donna-on-accent",
    "hover:bg-donna-accent-bright disabled:opacity-60",
  ].join(" "),
  masonry: "flex w-full items-start gap-4",
  masonryColumn: "flex min-w-0 flex-1 flex-col gap-4",
  card: [
    "flex w-full flex-col rounded-2xl border p-4 text-left",
    "transition-[transform,box-shadow] hover:-translate-y-0.5",
    "hover:shadow-[0_10px_28px_rgba(0,0,0,0.22)] cursor-pointer",
  ].join(" "),
  cardTitle: "mb-1.5 text-sm font-semibold leading-snug text-donna-text",
  cardBody: "whitespace-pre-wrap text-sm leading-relaxed text-donna-muted",
  cardPin: "mb-2 inline-flex text-donna-faint",
  section: "mb-8",
  sectionLabel:
    "mb-3 text-[11px] font-semibold uppercase tracking-[0.14em] text-donna-faint",
  empty: "py-16 text-center text-sm text-donna-muted",
  state: "py-8 text-center text-sm text-donna-muted",
  editorRoot: "fixed inset-0 z-50 flex items-center justify-center p-4",
  editorBackdrop: "absolute inset-0 bg-black/55",
  editorPanel: [
    "relative z-10 w-full max-w-lg rounded-2xl border p-4",
    "shadow-[0_20px_50px_rgba(0,0,0,0.45)]",
  ].join(" "),
  editorTitle: [
    "mb-2 w-full bg-transparent text-base font-semibold text-donna-text",
    "placeholder:text-donna-muted/70 focus:outline-none",
  ].join(" "),
  editorBody: [
    "min-h-[10rem] w-full resize-y bg-transparent text-sm leading-relaxed text-donna-text",
    "placeholder:text-donna-muted/70 focus:outline-none",
  ].join(" "),
  editorFooter: "mt-4 flex flex-wrap items-center justify-between gap-3",
} as const;
