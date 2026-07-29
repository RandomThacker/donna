"use client";

import type { ReactNode } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { buildMonthGrid, format, isToday } from "../../Calendar.utils";
import { miniCalendarStyles as styles } from "./MiniCalendar.styles";

const WEEKDAYS = ["S", "M", "T", "W", "T", "F", "S"] as const;

type MiniCalendarProps = {
  month: Date;
  selected: Date;
  onSelectDay: (day: Date) => void;
  onMonthShift: (direction: -1 | 1) => void;
  dayExtra?: (day: Date) => ReactNode;
  className?: string;
  "aria-label"?: string;
};

export function MiniCalendar({
  month,
  selected,
  onSelectDay,
  onMonthShift,
  dayExtra,
  className,
  "aria-label": ariaLabel = "Mini calendar",
}: MiniCalendarProps) {
  const days = buildMonthGrid(month);

  return (
    <div className={cn(styles.root, className)} aria-label={ariaLabel}>
      <div className={styles.header}>
        <p className={styles.month}>{format(month, "MMMM yyyy")}</p>
        <div className={styles.nav}>
          <button
            type="button"
            className={styles.navBtn}
            aria-label="Previous month"
            onClick={() => onMonthShift(-1)}
          >
            <Icon name="chevronLeft" className="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            className={styles.navBtn}
            aria-label="Next month"
            onClick={() => onMonthShift(1)}
          >
            <Icon name="chevronRight" className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
      <div className={styles.weekdays}>
        {WEEKDAYS.map((d, i) => (
          <span key={`${d}-${i}`}>{d}</span>
        ))}
      </div>
      <div className={styles.grid}>
        {days.map((day) => {
          const inMonth = day.getMonth() === month.getMonth();
          const isSelected =
            day.getDate() === selected.getDate() &&
            day.getMonth() === selected.getMonth() &&
            day.getFullYear() === selected.getFullYear();
          const extra = dayExtra?.(day);
          return (
            <button
              key={day.toISOString()}
              type="button"
              className={cn(
                styles.day,
                !inMonth && styles.dayMuted,
                isToday(day) && !isSelected && styles.dayToday,
                isSelected && styles.daySelected,
              )}
              onClick={() => onSelectDay(day)}
            >
              <span>{day.getDate()}</span>
              {extra ? <span className={styles.dayExtra}>{extra}</span> : null}
            </button>
          );
        })}
      </div>
    </div>
  );
}
