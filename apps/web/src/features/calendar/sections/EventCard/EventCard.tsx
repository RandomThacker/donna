import type { CSSProperties } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { contrastTextFor, isRecurring } from "../../Calendar.layout";
import type { CalendarEvent } from "../../Calendar.types";
import { formatEventTime } from "../../Calendar.utils";

export const eventCardStyles = {
  root: [
    "group flex w-full flex-col gap-0.5 overflow-hidden rounded-lg",
    "border border-black/10 px-2 py-1.5 text-left shadow-sm",
    "transition hover:brightness-110 focus-visible:outline-none",
    "focus-visible:ring-2 focus-visible:ring-donna-accent/50",
  ].join(" "),
  titleRow: "flex items-start gap-1.5",
  title: "min-w-0 flex-1 truncate text-[12px] font-medium leading-tight",
  meta: "truncate text-[10px] opacity-80",
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

  return (
    <button
      type="button"
      className={cn(eventCardStyles.root, className)}
      style={{
        backgroundColor: color,
        color: ink,
        ...style,
      }}
      onClick={onClick}
    >
      <div className={eventCardStyles.titleRow}>
        <span className={eventCardStyles.title}>
          {event.title || "(No title)"}
        </span>
        {isRecurring(event) ? (
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
