import type { CSSProperties } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { contrastTextFor, isRecurring } from "../../Calendar.layout";
import type { CalendarEvent } from "../../Calendar.types";
import { formatEventTime } from "../../Calendar.utils";

export const eventCardStyles = {
  root: [
    "group box-border flex h-full w-full min-h-0 cursor-pointer flex-col",
    "justify-center overflow-hidden rounded-lg border border-black/10",
    "px-2.5 py-1 text-left shadow-sm",
    "transition hover:brightness-110 focus-visible:outline-none",
    "focus-visible:ring-2 focus-visible:ring-donna-accent/50",
  ].join(" "),
  titleRow: "flex min-w-0 items-center gap-1",
  title: "min-w-0 truncate text-[12px] font-medium leading-none",
  titleGrow: "flex-1",
  meta: "truncate text-[10px] leading-none opacity-80",
  badges: "mt-0.5 flex flex-wrap items-center gap-1",
  badge:
    "rounded px-1 py-px text-[9px] font-medium uppercase tracking-wide opacity-90 bg-black/15",
  icon: "h-3 w-3 shrink-0 opacity-80",
} as const;

type EventCardProps = {
  event: CalendarEvent;
  color: string;
  compact?: boolean;
  className?: string;
  style?: CSSProperties;
  onClick?: () => void;
};

export function EventCard({
  event,
  color,
  compact = false,
  className,
  style,
  onClick,
}: EventCardProps) {
  const ink = contrastTextFor(color);
  const recurring = isRecurring(event);

  return (
    <button
      type="button"
      className={cn(eventCardStyles.root, className)}
      style={{
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        backgroundColor: color,
        color: ink,
        ...style,
      }}
      onClick={onClick}
    >
      <div className={eventCardStyles.titleRow}>
        <span
          className={cn(
            eventCardStyles.title,
            !compact && eventCardStyles.titleGrow,
          )}
        >
          {event.title || "(No title)"}
        </span>
        {recurring ? (
          <Icon name="repeat" className={eventCardStyles.icon} />
        ) : null}
      </div>
      {!compact ? (
        <>
          <p className={eventCardStyles.meta}>{formatEventTime(event)}</p>
          {event.location ? (
            <p className={eventCardStyles.meta}>
              <span className="inline-flex items-center gap-0.5">
                <Icon name="mapPin" className="h-2.5 w-2.5" />
                {event.location}
              </span>
            </p>
          ) : null}
          <div className={eventCardStyles.badges}>
            {event.all_day ? (
              <span className={eventCardStyles.badge}>All day</span>
            ) : null}
          </div>
        </>
      ) : null}
    </button>
  );
}
