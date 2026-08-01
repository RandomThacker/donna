"use client";

import { Fragment, type ReactNode } from "react";

import { cn } from "@/lib/cn";

const VIEW_EVENT_LINK =
  /\[View Event\]\((\/dashboard\/calendar\?event=[^)\s]+)\)/g;

type ChatBubbleTextProps = {
  text: string;
  className?: string;
  linkClassName?: string;
  onViewEvent?: (href: string) => void;
};

function renderSegments(
  text: string,
  linkClassName: string | undefined,
  onViewEvent: ((href: string) => void) | undefined,
): ReactNode[] {
  const nodes: ReactNode[] = [];
  let lastIndex = 0;
  const pattern = new RegExp(VIEW_EVENT_LINK.source, "g");
  let match: RegExpExecArray | null;
  let key = 0;

  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(
        <Fragment key={`t-${key++}`}>
          {text.slice(lastIndex, match.index)}
        </Fragment>,
      );
    }
    const href = match[1]!;
    nodes.push(
      <button
        key={`l-${key++}`}
        type="button"
        className={cn(
          "underline underline-offset-2",
          linkClassName ?? "text-[var(--im-blue)]",
        )}
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onViewEvent?.(href);
        }}
      >
        View Event
      </button>,
    );
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    nodes.push(
      <Fragment key={`t-${key++}`}>{text.slice(lastIndex)}</Fragment>,
    );
  }

  return nodes;
}

/** Renders chat bubble text and turns `[View Event](/dashboard/calendar?event=…)` into a link. */
export function ChatBubbleText({
  text,
  className,
  linkClassName,
  onViewEvent,
}: ChatBubbleTextProps) {
  if (!onViewEvent || !/\[View Event\]\(/.test(text)) {
    return <span className={className}>{text}</span>;
  }

  return (
    <span className={className}>
      {renderSegments(text, linkClassName, onViewEvent)}
    </span>
  );
}
