"use client";

import { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/cn";

import { HOUR_HEIGHT, HOURS } from "../../Calendar.layout";
import type { CalendarEvent, LaidOutEvent } from "../../Calendar.types";
import { format, isToday, weekDays } from "../../Calendar.utils";
import { EventCard } from "../EventCard";
import { timelineStyles as styles } from "./view.styles";

type DayBundle = {
  day: Date;
  allDay: CalendarEvent[];
  timed: LaidOutEvent[];
};

type WeekViewProps = {
  cursor: Date;
  days: DayBundle[];
  colorFor: (sourceId: string, event?: CalendarEvent) => string;
  onEventClick: (event: CalendarEvent) => void;
};

function hourLabel(hour: number): string {
  const d = new Date();
  d.setHours(hour, 0, 0, 0);
  return format(d, "h a");
}

function NowIndicator({ day }: { day: Date }) {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const id = window.setInterval(() => setNow(new Date()), 60_000);
    return () => window.clearInterval(id);
  }, []);
  if (!isToday(day)) {
    return null;
  }
  const minutes = now.getHours() * 60 + now.getMinutes();
  const top = (minutes / 60) * HOUR_HEIGHT;
  return (
    <div
      className="pointer-events-none absolute inset-x-0 z-20"
      style={{ top }}
      aria-hidden
    >
      <div className="h-px w-full bg-rose-500/90" />
    </div>
  );
}

export function WeekView({ cursor, days, colorFor, onEventClick }: WeekViewProps) {
  const headers = useMemo(() => weekDays(cursor), [cursor]);
  const height = HOURS.length * HOUR_HEIGHT;

  return (
    <div className={styles.root}>
      <div
        className={styles.weekHeader}
        style={{ gridTemplateColumns: "3.5rem repeat(7, minmax(0, 1fr))" }}
      >
        <div className={styles.weekHeaderGutter} />
        {headers.map((day) => (
          <div key={day.toISOString()} className={styles.weekDayHead}>
            <p className={styles.weekDayName}>{format(day, "EEE")}</p>
            <p
              className={cn(
                styles.weekDayNum,
                isToday(day) && styles.weekDayToday,
              )}
            >
              {format(day, "d")}
            </p>
          </div>
        ))}
      </div>

      <div
        className={styles.weekAllDay}
        style={{ gridTemplateColumns: "3.5rem repeat(7, minmax(0, 1fr))" }}
      >
        <div className="border-r border-donna-hairline px-1 py-2 text-[10px] text-donna-faint">
          All day
        </div>
        {days.map(({ day, allDay }) => (
          <div
            key={`all-${day.toISOString()}`}
            className="space-y-1 border-l border-donna-hairline p-1"
          >
            {allDay.map((event) => (
              <EventCard
                key={event.id}
                event={event}
                color={colorFor(event.calendar_source_id, event)}
                compact
                onClick={() => onEventClick(event)}
              />
            ))}
          </div>
        ))}
      </div>

      <div className={styles.scroll}>
        <div className={styles.weekCols} style={{ height }}>
          <div
            className="sticky left-0 z-10 w-14 shrink-0 border-r border-donna-hairline bg-donna-bg"
            style={{ height }}
          >
            {HOURS.map((hour) => (
              <div
                key={hour}
                className="relative border-b border-donna-hairline/70 pr-2 text-right text-[11px] text-donna-faint"
                style={{ height: HOUR_HEIGHT }}
              >
                {hour !== 0 ? (
                  <span className="-translate-y-1/2">{hourLabel(hour)}</span>
                ) : null}
              </div>
            ))}
          </div>
          {days.map(({ day, timed }) => (
            <div key={day.toISOString()} className={styles.weekCol} style={{ height }}>
              {HOURS.map((hour) => (
                <div
                  key={hour}
                  className="border-b border-donna-hairline/70"
                  style={{ height: HOUR_HEIGHT }}
                />
              ))}
              {timed.map((item) => (
                <div
                  key={item.event.id}
                  className="absolute px-0.5"
                  style={{
                    top: item.top,
                    height: item.height,
                    left: `${item.left}%`,
                    width: `calc(${item.width}% - 2px)`,
                  }}
                >
                  <EventCard
                    event={item.event}
                    color={colorFor(item.event.calendar_source_id, item.event)}
                    compact
                    className="h-full"
                    onClick={() => onEventClick(item.event)}
                  />
                </div>
              ))}
              <NowIndicator day={day} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
