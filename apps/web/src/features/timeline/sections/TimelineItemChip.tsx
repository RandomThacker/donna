"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { LaidOutTimelineItem, TimelineItem } from "../Timeline.types";
import {
  colorForTimelineItem,
  isDonnaEditable,
} from "../Timeline.utils";
import { contrastTextFor } from "../Timeline.layout";

type ChipProps = {
  item: TimelineItem;
  className?: string;
  style?: React.CSSProperties;
  onClick: () => void;
  compact?: boolean;
};

export function TimelineItemChip({
  item,
  className,
  style,
  onClick,
  compact,
}: ChipProps) {
  const bg = colorForTimelineItem(item);
  const fg = contrastTextFor(bg);
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex w-full items-start gap-1 overflow-hidden rounded-md px-1.5 py-0.5 text-left text-[11px] leading-snug",
        "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-donna-accent",
        className,
      )}
      style={{ background: bg, color: fg, ...style }}
      aria-label={`${item.title}${item.read_only ? ", read only" : ""}`}
    >
      <span className="mt-0.5 shrink-0 opacity-90">
        <Icon
          name={
            item.status === "CANCELLED"
              ? "close"
              : item.status === "COMPLETED"
                ? "check"
                : item.type === "REMINDER"
                  ? "bell"
                  : "calendar"
          }
          className="h-3 w-3"
        />
      </span>
      <span className="min-w-0 flex-1">
        <span className={cn("line-clamp-2 font-medium", compact && "line-clamp-1")}>
          {item.title}
        </span>
        {item.is_recurring ? (
          <Icon name="repeat" className="ml-0.5 inline h-2.5 w-2.5 opacity-80" />
        ) : null}
      </span>
      {!isDonnaEditable(item) ? (
        <span className="sr-only">Read only</span>
      ) : null}
    </button>
  );
}

type BlockProps = {
  laid: LaidOutTimelineItem;
  onClick: () => void;
};

export function TimelineTimedBlock({ laid, onClick }: BlockProps) {
  const width = 100 / laid.columns;
  const left = laid.column * width;
  return (
    <div
      className="pointer-events-auto absolute px-0.5"
      style={{
        top: laid.top,
        height: Math.max(18, laid.height),
        left: `${left}%`,
        width: `${width}%`,
      }}
    >
      <TimelineItemChip
        item={laid.item}
        onClick={onClick}
        className="h-full"
      />
    </div>
  );
}
