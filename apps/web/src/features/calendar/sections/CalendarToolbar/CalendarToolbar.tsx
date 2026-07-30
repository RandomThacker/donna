"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { CalendarView } from "../../Calendar.types";
import { isToday, mobileTitleParts } from "../../Calendar.utils";
import { toolbarStyles as styles } from "./CalendarToolbar.styles";

const VIEWS: Array<{ id: CalendarView; label: string }> = [
  { id: "day", label: "Day" },
  { id: "week", label: "Week" },
  { id: "month", label: "Month" },
  { id: "agenda", label: "Agenda" },
];

type CalendarToolbarProps = {
  title: string;
  cursor: Date;
  view: CalendarView;
  onViewChange: (view: CalendarView) => void;
  onPrev: () => void;
  onNext: () => void;
  onToday: () => void;
  onOpenSidebar: () => void;
  onSync: () => void;
  isSyncing: boolean;
  onCreate?: () => void;
};

export function CalendarToolbar({
  title,
  cursor,
  view,
  onViewChange,
  onPrev,
  onNext,
  onToday,
  onOpenSidebar,
  onSync,
  isSyncing,
  onCreate,
}: CalendarToolbarProps) {
  const mobile = mobileTitleParts(view, cursor);
  const viewingToday = isToday(cursor);

  return (
    <header className={styles.root}>
      <div className={styles.topRow}>
        <button
          type="button"
          className={styles.calendarsBtn}
          aria-label="Calendars"
          onClick={onOpenSidebar}
        >
          <Icon name="calendar" className="h-4 w-4" />
        </button>

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

        <button
          type="button"
          className={styles.dateBtn}
          onClick={onToday}
          aria-label={viewingToday ? title : `Go to today · ${title}`}
        >
          <span className={styles.mobileDate}>
            <span
              className={cn(
                styles.dayNumber,
                viewingToday && view === "day" && styles.dayNumberToday,
              )}
            >
              {mobile.primary}
            </span>
            {mobile.secondary ? (
              <span className={styles.dayMeta}>{mobile.secondary}</span>
            ) : null}
          </span>
          <h1 className={styles.desktopTitle}>{title}</h1>
        </button>

        <button
          type="button"
          className={cn(
            styles.syncBtn,
            "sm:hidden",
            isSyncing && styles.syncBtnActive,
          )}
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
        {onCreate ? (
          <button
            type="button"
            className={cn(styles.syncBtn, "sm:hidden")}
            aria-label="Create"
            onClick={onCreate}
          >
            <Icon name="plus" className="h-3.5 w-3.5" />
          </button>
        ) : null}
      </div>

      <div className={styles.bottomRow}>
        <div className={styles.mobileNav}>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label="Previous"
            onClick={onPrev}
          >
            <Icon name="chevronLeft" className="h-4 w-4" />
          </button>
        </div>

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
          <span className={cn(styles.divider, "hidden sm:block")} aria-hidden />
          {onCreate ? (
            <button
              type="button"
              className={cn(styles.syncBtn, "hidden sm:inline-flex sm:w-auto sm:gap-1.5 sm:px-3")}
              aria-label="Create"
              onClick={onCreate}
            >
              <Icon name="plus" className="h-3.5 w-3.5" />
              <span className="text-xs font-medium">Create</span>
            </button>
          ) : null}
          <button
            type="button"
            className={cn(
              styles.syncBtn,
              "hidden sm:grid",
              isSyncing && styles.syncBtnActive,
            )}
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

        <div className={styles.mobileNav}>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label="Next"
            onClick={onNext}
          >
            <Icon name="chevronRight" className="h-4 w-4" />
          </button>
        </div>
      </div>
    </header>
  );
}
