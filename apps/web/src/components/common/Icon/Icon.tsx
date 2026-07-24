import { cn } from "@/lib/cn";

import { iconPaths } from "./Icon.paths";
import type { IconProps } from "./Icon.types";

const filledIcons = new Set(["google"]);

export function Icon({ name, className }: IconProps) {
  const isFilled = filledIcons.has(name);

  return (
    <svg
      viewBox="0 0 24 24"
      fill={isFilled ? "currentColor" : "none"}
      stroke={isFilled ? "none" : "currentColor"}
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      className={cn("h-5 w-5", className)}
    >
      {iconPaths[name].map((d) => (
        <path key={d} d={d} />
      ))}
    </svg>
  );
}
