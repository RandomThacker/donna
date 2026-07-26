"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { CalendarAccountGroup } from "../../Calendar.types";
import { buildMonthGrid, format, isToday } from "../../Calendar.utils";
import { sidebarPanelStyles as styles } from "./CalendarSidebar.styles";

type CalendarSidebarProps = {
  cursor: Date;
  onSelectDay: (day: Date) => void;
  onMonthShift: (direction: -1 | 1) => void;
  accounts: CalendarAccountGroup[];
  onToggleAccount: (sourceIds: string[]) => void;
};

export function CalendarSidebar({
  cursor,
  onSelectDay,
  onMonthShift,
  accounts,
  onToggleAccount,
}: CalendarSidebarProps) {
  const days = buildMonthGrid(cursor);
  const weekdays = ["S", "M", "T", "W", "T", "F", "S"];

  return (
    <>
      <section className={styles.section} aria-label="Mini calendar">
        <h2 className={styles.sectionTitle}>Calendar</h2>
        <div className={styles.miniRoot}>
          <div className={styles.miniHeader}>
            <p className={styles.miniMonth}>{format(cursor, "MMMM yyyy")}</p>
            <div className={styles.miniNav}>
              <button
                type="button"
                className={styles.miniNavBtn}
                aria-label="Previous month"
                onClick={() => onMonthShift(-1)}
              >
                <Icon name="chevronLeft" className="h-3.5 w-3.5" />
              </button>
              <button
                type="button"
                className={styles.miniNavBtn}
                aria-label="Next month"
                onClick={() => onMonthShift(1)}
              >
                <Icon name="chevronRight" className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
          <div className={styles.miniWeekdays}>
            {weekdays.map((d, i) => (
              <span key={`${d}-${i}`}>{d}</span>
            ))}
          </div>
          <div className={styles.miniGrid}>
            {days.map((day) => {
              const inMonth = day.getMonth() === cursor.getMonth();
              const selected =
                day.getDate() === cursor.getDate() &&
                day.getMonth() === cursor.getMonth() &&
                day.getFullYear() === cursor.getFullYear();
              return (
                <button
                  key={day.toISOString()}
                  type="button"
                  className={cn(
                    styles.miniDay,
                    !inMonth && styles.miniDayMuted,
                    isToday(day) && !selected && styles.miniDayToday,
                    selected && styles.miniDaySelected,
                  )}
                  onClick={() => onSelectDay(day)}
                >
                  {day.getDate()}
                </button>
              );
            })}
          </div>
        </div>
      </section>

      <section className={styles.section} aria-label="Connected calendars">
        <h2 className={styles.sectionTitle}>Calendars</h2>
        {accounts.length === 0 ? (
          <p className="text-sm text-donna-muted">No calendars connected yet.</p>
        ) : (
          <ul className={styles.sourceList}>
            {accounts.map((account) => {
              const visible =
                account.visibleCount > 0 &&
                account.visibleCount === account.sourceIds.length;
              const partial =
                account.visibleCount > 0 &&
                account.visibleCount < account.sourceIds.length;
              return (
                <li key={account.accountId}>
                  <button
                    type="button"
                    className={styles.sourceRow}
                    aria-pressed={visible}
                    aria-label={`${account.label}, ${account.sourceIds.length} calendars`}
                    onClick={() => onToggleAccount(account.sourceIds)}
                  >
                    <span
                      className={cn(
                        styles.checkbox,
                        (visible || partial) && styles.checkboxOn,
                      )}
                      aria-hidden
                    >
                      {visible ? (
                        <Icon name="check" className="h-3 w-3" />
                      ) : partial ? (
                        <span className="h-0.5 w-2 rounded-full bg-donna-on-accent" />
                      ) : null}
                    </span>
                    <span
                      className={styles.colorDot}
                      style={{ backgroundColor: account.color }}
                      aria-hidden
                    />
                    <span className={styles.sourceName}>{account.label}</span>
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </section>
    </>
  );
}
