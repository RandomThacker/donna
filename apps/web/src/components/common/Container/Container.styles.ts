import type { ContainerWidth } from "./Container.types";

export const containerStyles = {
  base: "mx-auto w-full px-5 sm:px-8",
  widths: {
    narrow: "max-w-3xl",
    default: "max-w-6xl",
    wide: "max-w-7xl",
  } satisfies Record<ContainerWidth, string>,
} as const;
