export const phoneStyles = {
  wrap: "flex h-full w-full items-center justify-center px-0 py-4 xl:p-5",
  frame: [
    "relative mx-auto w-full max-w-[26.5rem]",
    "h-[min(52rem,78dvh)]",
    "rounded-[2.85rem] p-[0.58rem]",
    "bg-[var(--phone-frame)]",
    "shadow-[0_28px_70px_-24px_rgb(0_0_0_/_0.65),inset_0_1px_0_var(--phone-frame-shine)]",
    "ring-1 ring-[var(--phone-frame-edge)]",
    "xl:h-[42rem] xl:w-[20.25rem] xl:max-w-none",
  ].join(" "),
  sideBtnLeft: [
    "pointer-events-none absolute -left-[0.15rem] top-[7.2rem] h-7 w-[0.15rem]",
    "rounded-l-sm bg-[var(--phone-frame)] ring-1 ring-[var(--phone-frame-edge)]",
  ].join(" "),
  sideBtnVol: [
    "pointer-events-none absolute -left-[0.15rem] top-[9.4rem] h-12 w-[0.15rem]",
    "rounded-l-sm bg-[var(--phone-frame)] ring-1 ring-[var(--phone-frame-edge)]",
  ].join(" "),
  sideBtnRight: [
    "pointer-events-none absolute -right-[0.15rem] top-[9.6rem] h-16 w-[0.15rem]",
    "rounded-r-sm bg-[var(--phone-frame)] ring-1 ring-[var(--phone-frame-edge)]",
  ].join(" "),
  screen: [
    "relative flex h-full w-full flex-col overflow-hidden rounded-[2.35rem]",
    "bg-[var(--im-bg)]",
  ].join(" "),
  statusBar: [
    "absolute inset-x-0 top-0 z-30 flex h-11 items-end justify-between px-7 pb-1.5",
    "text-[0.72rem] font-semibold tracking-tight text-[var(--phone-status)]",
  ].join(" "),
  island: [
    "absolute left-1/2 top-[0.55rem] z-40 h-[1.65rem] w-[5.9rem] -translate-x-1/2",
    "rounded-full bg-[var(--phone-island)]",
  ].join(" "),
  homeIndicator: [
    "absolute bottom-[0.35rem] left-1/2 z-40 h-[0.28rem] w-[7.75rem] -translate-x-1/2",
    "rounded-full bg-[var(--phone-home)]",
  ].join(" "),
  content: "relative z-10 flex min-h-0 flex-1 flex-col pt-12",
} as const;

export const phoneFullscreenStyles = {
  root: [
    // Above the bottom bar (z-60) so chat owns the full screen.
    "fixed inset-0 z-[70] flex flex-col overflow-hidden md:hidden",
    "origin-bottom-right bg-[var(--im-bg)]",
    "pt-[env(safe-area-inset-top)]",
  ].join(" "),
  rootEnter: "animate-donna-phone-in",
  rootExit: "animate-donna-phone-out",
  body: "flex min-h-0 flex-1 flex-col overflow-hidden",
} as const;

export const phoneFabStyles = {
  button: [
    "fixed z-[55] grid h-12 w-12 touch-none place-items-center rounded-full md:hidden",
    "bg-transparent animate-donna-fab-glow",
    "cursor-grab active:cursor-grabbing",
    "select-none",
  ].join(" "),
  buttonDragging: "scale-105",
  mark: [
    "h-full w-full rounded-full",
    "bg-gradient-to-br from-donna-accent-bright to-donna-accent-deep",
    "[mask-image:radial-gradient(farthest-side,transparent_34%,#000_35%)]",
    "[-webkit-mask-image:radial-gradient(farthest-side,transparent_34%,#000_35%)]",
  ].join(" "),
  badge: [
    "pointer-events-none absolute right-0.5 top-0.5 z-10",
    "h-2.5 w-2.5 rounded-full bg-[#ff3b30]",
    "ring-2 ring-donna-bg",
  ].join(" "),
} as const;

export const iMessageStyles = {
  listRoot: "flex min-h-0 flex-1 flex-col bg-[var(--im-bg)]",
  listHeader: "bg-[var(--im-bg)] px-4 pb-3 pt-1",
  listTopRow: "mb-2 flex items-center justify-between",
  edit: "text-[14px] text-[var(--im-blue)]",
  listTitle: "text-[26px] font-bold leading-none tracking-tight text-[var(--im-text)]",
  compose: "grid h-7 w-7 place-items-center text-[var(--im-blue)]",
  search: [
    "mt-3.5 flex items-center gap-2 rounded-[10px] px-2.5 py-2",
    "bg-[var(--im-search)] text-[13px] text-[var(--im-muted)]",
  ].join(" "),
  listBody: "min-h-0 flex-1 overflow-y-auto bg-[var(--im-bg)] pt-1",
  row: [
    "flex w-full items-start gap-3 px-4 py-3 text-left",
    "active:bg-[var(--im-surface-2)]",
  ].join(" "),
  avatar: [
    "mt-0.5 grid h-[44px] w-[44px] shrink-0 place-items-center rounded-full",
    "bg-[#c7a57a] text-[15px] font-semibold text-white",
  ].join(" "),
  rowMain: [
    "min-w-0 flex-1 border-b border-[var(--im-separator)] pb-3",
  ].join(" "),
  rowTop: "flex items-center gap-2",
  rowName: "min-w-0 flex-1 truncate text-[14px] font-semibold text-[var(--im-text)]",
  rowMeta: "flex shrink-0 items-center gap-0.5",
  rowTime: "text-[12px] text-[var(--im-muted)]",
  chevron: "h-[12px] w-[12px] text-[var(--im-muted)] opacity-60",
  rowBottom: "mt-1 flex items-start gap-2",
  rowPreview: [
    "min-w-0 flex-1 text-[12px] leading-[1.35] text-[var(--im-muted)]",
    "line-clamp-2",
  ].join(" "),
  unread: [
    "mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-[#ff3b30]",
  ].join(" "),

  chatRoot: "relative flex min-h-0 flex-1 flex-col bg-[var(--im-bg)]",
  chatNav: [
    "relative z-20 flex shrink-0 items-center justify-center",
    "min-h-[3.25rem] border-b border-[var(--im-separator)]/50 bg-[var(--im-nav)]",
    "px-2.5 pb-2.5 pt-1.5 backdrop-blur-xl",
  ].join(" "),
  back: [
    "absolute left-2 top-1/2 flex -translate-y-1/2 items-center gap-0.5",
    "text-[13px] leading-none text-[var(--im-blue)]",
  ].join(" "),
  chatTitleWrap: "flex flex-col items-center justify-center gap-0.5 text-center",
  chatAvatar: [
    "grid h-8 w-8 place-items-center rounded-full",
    "bg-[#c7a57a] text-[11px] font-semibold leading-none text-white",
  ].join(" "),
  chatName: "w-full text-center text-[11px] font-medium leading-none text-[var(--im-text)]",
  chatInfo: [
    "absolute right-2 top-1/2 grid h-7 w-7 -translate-y-1/2 place-items-center",
    "text-[var(--im-blue)]",
  ].join(" "),
  chatBody: [
    "flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain px-3.5 pb-3 pt-3",
    "[-webkit-overflow-scrolling:touch]",
  ].join(" "),
  chatBodyInner: "mt-auto flex w-full flex-col",
  stamp: "mb-3 text-center text-[10px] font-medium text-[var(--im-muted)]",
  newMessageRule: "my-3 flex w-full items-center gap-3",
  newMessageLine: "h-px min-w-0 flex-1 bg-[var(--im-blue)]",
  newMessageLabel: [
    "shrink-0 text-[10px] font-medium tracking-wide text-[var(--im-blue)]",
  ].join(" "),
  cluster: "flex flex-col",
  bubbleWrap: "relative max-w-[78%] px-0",
  bubbleWrapIn: "self-start",
  bubbleWrapOut: "self-end",
  bubble: [
    "relative flex flex-col gap-1 px-[11px] py-[7px]",
    "text-[13px] leading-[1.3]",
  ].join(" "),
  bubbleIn: [
    "rounded-[16px] bg-[var(--im-bubble-in)] text-[var(--im-bubble-in-text)]",
  ].join(" "),
  bubbleOut: "rounded-[16px] bg-[var(--im-bubble-out)] text-white",
  bubbleGrouped: "mt-[3px]",
  bubbleSpaced: "mt-3",
  bubbleInLast: "rounded-bl-[4px]",
  bubbleOutLast: "rounded-br-[4px]",
  bubbleInMiddle: "rounded-bl-[16px]",
  bubbleOutMiddle: "rounded-br-[16px]",
  bubbleText: "whitespace-pre-wrap",
  bubbleTime: "self-end text-[9px] leading-none opacity-70",
  bubbleTimeIn: "text-[var(--im-muted)]",
  bubbleTimeOut: "text-white/75",
  bubbleEnterIn: "animate-donna-im-in",
  bubbleEnterOut: "animate-donna-im-out",
  typingWrap: "relative mt-3 max-w-[78%] self-start",
  typingEnter: "animate-donna-im-typing-in",
  typingExit: "animate-donna-im-typing-out",
  typingBubble: [
    "flex items-center gap-1 rounded-[16px] rounded-bl-[4px]",
    "bg-[var(--im-bubble-in)] px-3.5 py-2.5",
  ].join(" "),
  typingDot: "h-1.5 w-1.5 rounded-full bg-[var(--im-muted)] animate-donna-im-dot",
  tailIn: [
    "pointer-events-none absolute bottom-0 -left-[4px] h-[14px] w-[10px]",
    "text-[var(--im-bubble-in)]",
  ].join(" "),
  tailOut: [
    "pointer-events-none absolute bottom-0 -right-[4px] h-[14px] w-[10px]",
    "text-[var(--im-bubble-out)]",
  ].join(" "),
  composerDock: [
    "relative z-20 flex shrink-0 flex-col",
    "border-t border-[var(--im-separator)]/40 bg-[var(--im-nav)]",
    "backdrop-blur-xl",
    "pb-[max(1.5rem,env(safe-area-inset-bottom))]",
  ].join(" "),
  suggestionRow: [
    "flex gap-2 overflow-x-auto overscroll-x-contain",
    "border-b border-[var(--im-separator)]/40",
    "bg-black px-2.5 py-2",
    "scrollbar-none [-ms-overflow-style:none] [scrollbar-width:none]",
    "[&::-webkit-scrollbar]:hidden",
  ].join(" "),
  suggestionPill: [
    "shrink-0 rounded-full border border-dashed border-[var(--im-muted)]/50",
    "bg-black px-3.5 py-2",
    "text-[11px] font-medium leading-none tracking-wide",
    "text-[var(--im-muted)] whitespace-nowrap",
    "transition-colors",
    "hover:border-[var(--im-blue)]/55 hover:text-[var(--im-blue)]",
    "active:border-[var(--im-blue)]/40",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--im-blue)]/40",
  ].join(" "),
  composer: [
    "relative flex items-end gap-2",
    "px-2.5 pt-2",
  ].join(" "),
  plus: [
    "mb-0.5 grid h-[28px] w-[28px] shrink-0 place-items-center rounded-full",
    "bg-[var(--im-surface-2)] text-[var(--im-muted)]",
  ].join(" "),
  inputShell: [
    "flex min-w-0 flex-1 items-center rounded-full",
    "border border-[var(--im-input-border)] bg-[var(--im-input-bg)]",
    "pl-3.5 pr-1",
  ].join(" "),
  input: [
    "min-w-0 flex-1 bg-transparent py-1.5 text-[13px] text-[var(--im-text)]",
    "placeholder:text-[var(--im-muted)] focus:outline-none",
  ].join(" "),
  mic: "grid h-7 w-7 shrink-0 place-items-center text-[var(--im-blue)]",
  send: [
    "grid h-7 w-7 shrink-0 place-items-center rounded-full",
    "bg-[var(--im-blue)] text-white",
  ].join(" "),
} as const;
