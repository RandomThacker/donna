import { cn } from "@/lib/cn";

import { textStyles } from "./Text.styles";
import type { TextProps } from "./Text.types";

export function Text({
  children,
  as: Tag = "p",
  tone = "cream",
  className,
  id,
}: TextProps) {
  return (
    <Tag id={id} className={cn(textStyles.tones[tone], className)}>
      {children}
    </Tag>
  );
}

export function BodyText({ children, className, tone = "muted", id }: TextProps) {
  return (
    <Text as="p" tone={tone} id={id} className={cn(textStyles.body, className)}>
      {children}
    </Text>
  );
}

export function DisplayText({
  children,
  className,
  tone = "cream",
  as = "h1",
  id,
}: TextProps) {
  return (
    <Text as={as} tone={tone} id={id} className={cn(textStyles.display, className)}>
      {children}
    </Text>
  );
}

export function TitleText({
  children,
  className,
  tone = "cream",
  as = "h2",
  id,
}: TextProps) {
  return (
    <Text as={as} tone={tone} id={id} className={cn(textStyles.title, className)}>
      {children}
    </Text>
  );
}

export function EyebrowText({ children, className }: Omit<TextProps, "tone" | "as">) {
  return <p className={cn(textStyles.eyebrow, className)}>{children}</p>;
}
