export const greetingStyles = {
  box: "order-1 col-span-12 md:order-1 md:mb-1",
  title: [
    "font-display text-[1.6rem] leading-[1.15] tracking-tight text-donna-text",
    "sm:text-4xl sm:leading-tight",
  ].join(" "),
  name: "italic text-donna-accent",
  row: "mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-2 sm:mt-3",
  summary: "text-[0.8rem] text-donna-muted sm:text-sm",
  nudge: "text-sm text-donna-accent",
} as const;
