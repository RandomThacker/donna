"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { CalendarView } from "../../Calendar.types";
import { toolbarStyles as styles } from "./CalendarToolbar.styles";

const VIEWS: Array<{ id: CalendarView; label: string }> = [
  { id: "day", label: "Day" },
  { id: "week", label: "Week" },
  { id: "month", label: "Month" },
  { id: "agenda", label: "Agenda" },
];

type CalendarToolbarProps = {
  title: string;
  view: CalendarView;
  onViewChange: (view: CalendarView) => void;
  onPrev: () => void;
  onNext: () => void;
  onToday: () => void;
  onOpenSidebar: () => void;
  onSync: () => void;
  isSyncing: boolean;
};

export function CalendarToolbar({
  title,
  view,
  onViewChange,
  onPrev,
  onNext,
  onToday,
  onOpenSidebar,
  onSync,
  isSyncing,
}: CalendarToolbarProps) {
  return (
    <header className={styles.root}>
      <div className={styles.left}>
        <div className={styles.navGroup}>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label="Previous"
            onClick={onPrev}
          >
            <Icon name="chevronLeft" className="h-4 w-4" />
          </button>
          <button type="button" className={styles.todayBtn} onClick={onToday}>
            Today
          </button>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label="Next"
            onClick={onNext}
          >
            <Icon name="chevronRight" className="h-4 w-4" />
          </button>
        </div>
        <h1 className={styles.title}>{title}</h1>
      </div>

      <div className={styles.right}>
        <button
          type="button"
          className={styles.calendarsBtn}
          onClick={onOpenSidebar}
        >
          Calendars
        </button>

        <div className={styles.controls}>
          <div
            className={styles.viewSwitch}
            role="tablist"
            aria-label="Calendar view"
          >
            {VIEWS.map((item) => (
              <button
                key={item.id}
                type="button"
                role="tab"
                aria-selected={view === item.id}
                className={cn(
                  styles.viewBtn,
                  view === item.id && styles.viewBtnActive,
                )}
                onClick={() => onViewChange(item.id)}
              >
                {item.label}
              </button>
            ))}
          </div>
          <span className={styles.divider} aria-hidden />
          <button
            type="button"
            className={cn(styles.syncBtn, isSyncing && styles.syncBtnActive)}
            disabled={isSyncing}
            aria-label={isSyncing ? "Syncing calendars" : "Sync calendars"}
            title={isSyncing ? "Syncing…" : "Sync"}
            onClick={onSync}
          >
            <Icon
              name="refresh"
              className={cn("h-3.5 w-3.5", isSyncing && styles.spin)}
            />
          </button>
        </div>
      </div>
    </header>
  );
}
