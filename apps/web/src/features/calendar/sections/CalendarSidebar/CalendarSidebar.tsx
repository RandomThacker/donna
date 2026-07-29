"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { CalendarAccountGroup } from "../../Calendar.types";
import { MiniCalendar } from "../MiniCalendar";
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
  return (
    <>
      <section className={styles.section} aria-label="Mini calendar">
        <h2 className={styles.sectionTitle}>Calendar</h2>
        <MiniCalendar
          month={cursor}
          selected={cursor}
          onSelectDay={onSelectDay}
          onMonthShift={onMonthShift}
        />
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
