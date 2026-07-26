"use client";

import { useMemo } from "react";

import { cn } from "@/lib/cn";

import { contrastTextFor, eventsOverlappingDay } from "../../Calendar.layout";
import type { CalendarEvent } from "../../Calendar.types";
import { buildMonthGrid, format, isToday } from "../../Calendar.utils";
import { timelineStyles as styles } from "./view.styles";

type MonthViewProps = {
  cursor: Date;
  events: CalendarEvent[];
  colorFor: (sourceId: string) => string;
  onSelectDay: (day: Date) => void;
  onEventClick: (event: CalendarEvent) => void;
};

const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const MAX_CHIPS = 3;

export function MonthView({
  cursor,
  events,
  colorFor,
  onSelectDay,
  onEventClick,
}: MonthViewProps) {
  const days = useMemo(() => buildMonthGrid(cursor), [cursor]);

  return (
    <div className={styles.monthRoot}>
      <div className={styles.monthGrid}>
        {WEEKDAYS.map((d) => (
          <div key={d} className={styles.monthWeekday}>
            {d}
          </div>
        ))}
        {days.map((day) => {
          const inMonth = day.getMonth() === cursor.getMonth();
          const dayEvents = eventsOverlappingDay(events, day);
          const visible = dayEvents.slice(0, MAX_CHIPS);
          const overflow = dayEvents.length - visible.length;

          return (
            <div
              key={day.toISOString()}
              className={cn(
                styles.monthCell,
                !inMonth && styles.monthCellMuted,
                isToday(day) && styles.monthCellToday,
              )}
            >
              <button
                type="button"
                className={cn(
                  styles.monthDateBtn,
                  isToday(day) && styles.monthDateToday,
                )}
                onClick={() => onSelectDay(day)}
              >
                {format(day, "d")}
              </button>
              <div className={styles.monthChips}>
                {visible.map((event) => (
                  <button
                    key={event.id}
                    type="button"
                    className={styles.monthChip}
                    style={{
                      backgroundColor: colorFor(event.calendar_source_id),
                      color: contrastTextFor(
                        colorFor(event.calendar_source_id),
                      ),
                    }}
                    onClick={() => onEventClick(event)}
                  >
                    {event.all_day
                      ? event.title
                      : `${format(new Date(event.start_time), "h:mma")} ${event.title}`}
                  </button>
                ))}
                {overflow > 0 ? (
                  <button
                    type="button"
                    className={styles.monthMore}
                    onClick={() => onSelectDay(day)}
                  >
                    +{overflow} more
                  </button>
                ) : null}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
