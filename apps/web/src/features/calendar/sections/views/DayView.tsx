"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import { HOUR_HEIGHT, HOURS } from "../../Calendar.layout";
import type { CalendarEvent, LaidOutEvent } from "../../Calendar.types";
import { format, isToday } from "../../Calendar.utils";
import { EventCard } from "../EventCard";
import { timelineStyles as styles } from "./view.styles";

type DayViewProps = {
  day: Date;
  allDay: CalendarEvent[];
  timed: LaidOutEvent[];
  colorFor: (sourceId: string) => string;
  onEventClick: (event: CalendarEvent) => void;
};

function nowOffsetPx(date: Date = new Date()): number {
  return ((date.getHours() * 60 + date.getMinutes()) / 60) * HOUR_HEIGHT;
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

  return (
    <div className={styles.nowLine} style={{ top: nowOffsetPx(now) }} aria-hidden>
      <span className={styles.nowDot} />
      <div className={styles.nowRule} />
    </div>
  );
}

export function DayView({
  day,
  allDay,
  timed,
  colorFor,
  onEventClick,
}: DayViewProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const height = HOURS.length * HOUR_HEIGHT;
  const hasEvents = allDay.length > 0 || timed.length > 0;
  const viewingToday = isToday(day);

  const hourLabels = useMemo(
    () =>
      HOURS.map((hour) => {
        const d = new Date(day);
        d.setHours(hour, 0, 0, 0);
        return format(d, "h a");
      }),
    [day],
  );

  useEffect(() => {
    if (!viewingToday) {
      return;
    }

    const el = scrollRef.current;
    if (!el) {
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      const top = nowOffsetPx();
      const target = Math.max(0, top - el.clientHeight * 0.33);
      el.scrollTo({ top: target, behavior: "smooth" });
    });

    return () => window.cancelAnimationFrame(frame);
  }, [viewingToday, day]);

  return (
    <div className={styles.root}>
      {allDay.length > 0 ? (
        <div className={styles.allDay}>
          <p className={styles.allDayLabel}>All day</p>
          <div className={styles.allDayList}>
            {allDay.map((event) => (
              <EventCard
                key={event.id}
                event={event}
                color={colorFor(event.calendar_source_id)}
                compact
                className="max-w-[14rem]"
                onClick={() => onEventClick(event)}
              />
            ))}
          </div>
        </div>
      ) : null}
      <div ref={scrollRef} className={styles.scroll}>
        <div className={styles.grid} style={{ height }}>
          {!hasEvents ? (
            <div className="pointer-events-none absolute inset-0 z-[1] flex items-center justify-center">
              <p className="rounded-full bg-donna-surface/80 px-4 py-2 text-sm text-donna-muted backdrop-blur-sm">
                No events for this day
              </p>
            </div>
          ) : null}
          {HOURS.map((hour, index) => (
            <div
              key={hour}
              className={styles.hourRow}
              style={{ height: HOUR_HEIGHT }}
            >
              <div className={styles.gutter}>
                {hour !== 0 ? (
                  <span className={styles.gutterLabel}>{hourLabels[index]}</span>
                ) : null}
              </div>
              <div className={styles.lane} />
            </div>
          ))}
          <div className={styles.eventsLayer}>
            {timed.map((item) => (
              <div
                key={item.event.id}
                className={styles.eventAbs}
                style={{
                  top: item.top,
                  height: item.height,
                  left: `${item.left}%`,
                  width: `calc(${item.width}% - 0.25rem)`,
                }}
              >
                <EventCard
                  event={item.event}
                  color={colorFor(item.event.calendar_source_id)}
                  className="h-full"
                  onClick={() => onEventClick(item.event)}
                />
              </div>
            ))}
            <NowIndicator day={day} />
          </div>
        </div>
      </div>
    </div>
  );
}
