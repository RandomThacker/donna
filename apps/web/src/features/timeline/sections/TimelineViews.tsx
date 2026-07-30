"use client";

import { format, isSameDay, isSameMonth, isToday } from "date-fns";

import { cn } from "@/lib/cn";

import { HOURS, HOUR_HEIGHT, monthCells, weekDays } from "../Timeline.layout";
import type { LaidOutTimelineItem, TimelineItem } from "../Timeline.types";
import { groupAgendaItems } from "../Timeline.layout";
import { TimelineItemChip, TimelineTimedBlock } from "./TimelineItemChip";
import { colorForTimelineItem } from "../Timeline.utils";

type DayBundle = {
  allDay: TimelineItem[];
  timed: LaidOutTimelineItem[];
};

type Shared = {
  onOpen: (item: TimelineItem) => void;
  onCreateDay: (day: Date) => void;
};

export function TimelineSkeleton() {
  return (
    <div className="animate-pulse space-y-3 p-4" aria-busy="true">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="h-16 rounded-xl bg-donna-surface" />
      ))}
    </div>
  );
}

export function TimelineError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <div className="grid place-items-center p-10 text-center">
      <p className="text-sm text-donna-muted">{message}</p>
      <button
        type="button"
        className="mt-3 rounded-full border border-donna-border px-4 py-2 text-sm"
        onClick={onRetry}
      >
        Try again
      </button>
    </div>
  );
}

export function TimelineEmpty() {
  return (
    <div className="grid place-items-center p-10 text-center">
      <p className="font-display text-xl">Nothing on the timeline</p>
      <p className="mt-2 max-w-sm text-sm text-donna-muted">
        Click a day to create a Donna event or reminder.
      </p>
    </div>
  );
}

export function MonthView({
  cursor,
  items,
  onOpen,
  onCreateDay,
}: Shared & { cursor: Date; items: TimelineItem[] }) {
  const cells = monthCells(cursor);
  return (
    <div className="grid h-full min-h-[28rem] grid-cols-7 grid-rows-[auto_repeat(6,minmax(0,1fr))]">
      {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
        <div
          key={d}
          className="border-b border-donna-hairline px-2 py-2 text-center text-[11px] font-medium text-donna-muted"
        >
          {d}
        </div>
      ))}
      {cells.map((day) => {
        const dayItems = items.filter((item) => {
          const s = new Date(item.start_at);
          return isSameDay(s, day);
        });
        return (
          <button
            key={day.toISOString()}
            type="button"
            className={cn(
              "flex min-h-[5.5rem] flex-col gap-0.5 overflow-hidden border-b border-r border-donna-hairline p-1 text-left",
              "hover:bg-donna-surface/50",
              !isSameMonth(day, cursor) && "bg-donna-bg/40 text-donna-faint",
            )}
            onClick={() => onCreateDay(day)}
          >
            <span
              className={cn(
                "mb-0.5 inline-flex h-6 w-6 items-center justify-center rounded-full text-xs",
                isToday(day) && "bg-donna-accent text-donna-on-accent",
              )}
            >
              {format(day, "d")}
            </span>
            <div className="flex min-h-0 flex-1 flex-col gap-0.5 overflow-hidden">
              {dayItems.slice(0, 3).map((item) => (
                <div
                  key={item.occurrence_id}
                  onClick={(e) => e.stopPropagation()}
                  onKeyDown={(e) => e.stopPropagation()}
                >
                  <TimelineItemChip
                    item={item}
                    compact
                    onClick={() => onOpen(item)}
                  />
                </div>
              ))}
              {dayItems.length > 3 ? (
                <span className="px-1 text-[10px] text-donna-muted">
                  +{dayItems.length - 3} more
                </span>
              ) : null}
            </div>
          </button>
        );
      })}
    </div>
  );
}

function TimeGrid({
  days,
  dayLayout,
  onOpen,
  onCreateDay,
}: Shared & {
  days: Date[];
  dayLayout: (day: Date) => DayBundle;
}) {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div
        className="grid shrink-0 border-b border-donna-hairline"
        style={{ gridTemplateColumns: `3.5rem repeat(${days.length}, minmax(0,1fr))` }}
      >
        <div />
        {days.map((day) => (
          <button
            key={day.toISOString()}
            type="button"
            className="border-l border-donna-hairline px-2 py-2 text-center hover:bg-donna-surface/40"
            onClick={() => onCreateDay(day)}
          >
            <div className="text-[11px] text-donna-muted">{format(day, "EEE")}</div>
            <div
              className={cn(
                "mx-auto mt-1 grid h-7 w-7 place-items-center rounded-full text-sm",
                isToday(day) && "bg-donna-accent text-donna-on-accent",
              )}
            >
              {format(day, "d")}
            </div>
          </button>
        ))}
      </div>
      <div
        className="grid shrink-0 border-b border-donna-hairline"
        style={{ gridTemplateColumns: `3.5rem repeat(${days.length}, minmax(0,1fr))` }}
      >
        <div className="px-1 py-1 text-[10px] text-donna-faint">all-day</div>
        {days.map((day) => {
          const { allDay } = dayLayout(day);
          return (
            <div
              key={`all-${day.toISOString()}`}
              className="min-h-10 space-y-0.5 border-l border-donna-hairline p-1"
            >
              {allDay.map((item) => (
                <TimelineItemChip
                  key={item.occurrence_id}
                  item={item}
                  compact
                  onClick={() => onOpen(item)}
                />
              ))}
            </div>
          );
        })}
      </div>
      <div className="relative min-h-0 flex-1 overflow-auto">
        <div
          className="grid"
          style={{
            gridTemplateColumns: `3.5rem repeat(${days.length}, minmax(0,1fr))`,
            height: HOUR_HEIGHT * 24,
          }}
        >
          <div className="relative">
            {HOURS.map((h) => (
              <div
                key={h}
                className="absolute right-1 -translate-y-1/2 text-[10px] text-donna-faint"
                style={{ top: h * HOUR_HEIGHT }}
              >
                {format(new Date(2000, 0, 1, h), "ha")}
              </div>
            ))}
          </div>
          {days.map((day) => {
            const { timed } = dayLayout(day);
            return (
              <div
                key={day.toISOString()}
                className="relative border-l border-donna-hairline"
                onDoubleClick={() => onCreateDay(day)}
              >
                {HOURS.map((h) => (
                  <div
                    key={h}
                    className="absolute inset-x-0 border-t border-donna-hairline/60"
                    style={{ top: h * HOUR_HEIGHT, height: HOUR_HEIGHT }}
                  />
                ))}
                <div className="pointer-events-none absolute inset-0">
                  {timed.map((laid) => (
                    <TimelineTimedBlock
                      key={laid.item.occurrence_id}
                      laid={laid}
                      onClick={() => onOpen(laid.item)}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export function WeekView({
  cursor,
  dayLayout,
  onOpen,
  onCreateDay,
}: Shared & {
  cursor: Date;
  dayLayout: (day: Date) => DayBundle;
}) {
  return (
    <TimeGrid
      days={weekDays(cursor)}
      dayLayout={dayLayout}
      onOpen={onOpen}
      onCreateDay={onCreateDay}
    />
  );
}

export function DayView({
  cursor,
  dayLayout,
  onOpen,
  onCreateDay,
}: Shared & {
  cursor: Date;
  dayLayout: (day: Date) => DayBundle;
}) {
  return (
    <TimeGrid
      days={[cursor]}
      dayLayout={dayLayout}
      onOpen={onOpen}
      onCreateDay={onCreateDay}
    />
  );
}

export function AgendaView({
  cursor,
  items,
  onOpen,
}: {
  cursor: Date;
  items: TimelineItem[];
  onOpen: (item: TimelineItem) => void;
}) {
  const groups = groupAgendaItems(items, cursor);
  if (groups.length === 0) return <TimelineEmpty />;
  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-6 p-4 sm:p-6">
      {groups.map((group) => (
        <section key={group.key}>
          <h2 className="mb-2 font-display text-lg">{group.label}</h2>
          <ul className="space-y-2">
            {group.items.map((item) => (
              <li key={item.occurrence_id}>
                <button
                  type="button"
                  className="flex w-full items-stretch gap-3 rounded-xl border border-donna-hairline bg-donna-surface p-3 text-left hover:border-donna-accent/30"
                  onClick={() => onOpen(item)}
                >
                  <span
                    className="w-1 shrink-0 rounded-full"
                    style={{ background: colorForTimelineItem(item) }}
                    aria-hidden
                  />
                  <div className="min-w-0 flex-1">
                    <p className="font-medium">{item.title}</p>
                    <p className="mt-0.5 text-xs text-donna-muted">
                      {item.all_day
                        ? "All day"
                        : `${format(new Date(item.start_at), "h:mm a")} – ${format(new Date(item.end_at), "h:mm a")}`}
                      {item.type === "REMINDER" ? " · Reminder" : ""}
                      {item.is_recurring ? " · Repeats" : ""}
                    </p>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}

