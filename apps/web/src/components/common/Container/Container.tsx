import { cn } from "@/lib/cn";

import { containerStyles } from "./Container.styles";
import type { ContainerProps } from "./Container.types";

export function Container({
  children,
  width = "default",
  className,
  as: Tag = "div",
}: ContainerProps) {
  return (
    <Tag className={cn(containerStyles.base, containerStyles.widths[width], className)}>
      {children}
    </Tag>
  );
}
