"use client";

import { addMonths } from "date-fns";
import { useMemo } from "react";
import { usePathname, useRouter } from "next/navigation";

import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useCalendarController } from "./Calendar.logic";
import { calendarStyles as styles } from "./Calendar.styles";
import { weekDays } from "./Calendar.utils";
import { CalendarSidebar } from "./sections/CalendarSidebar";
import { CalendarToolbar } from "./sections/CalendarToolbar";
import { EventDrawer } from "./sections/EventDrawer";
import {
  AgendaView,
  CalendarEmptySources,
  CalendarErrorState,
  CalendarSkeleton,
  DayView,
  MonthView,
  WeekView,
} from "./sections/views";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "D";
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase();
  }
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

export function Calendar() {
  const router = useRouter();
  const pathname = usePathname();
  const { user, signOut } = useAuth();
  const cal = useCalendarController();

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);
  const nav = navItemsForPath(pathname);

  const weekBundles = useMemo(() => {
    return weekDays(cal.cursor).map((day) => ({
      day,
      ...cal.dayLayout(day),
    }));
  }, [cal.cursor, cal.dayLayout]);

  const dayBundle = useMemo(
    () => cal.dayLayout(cal.cursor),
    [cal.cursor, cal.dayLayout],
  );

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={nav}
          profileName={profileName}
          profileInitials={profileInitials}
          profileAvatarUrl={user?.avatar_url}
          onSignOut={() => {
            void (async () => {
              await signOut();
              router.replace("/");
            })();
          }}
        />

        <main className={styles.workspace}>
          <CalendarToolbar
            title={cal.title}
            view={cal.view}
            onViewChange={cal.setView}
            onPrev={cal.goPrev}
            onNext={cal.goNext}
            onToday={cal.goToday}
            onOpenSidebar={() => cal.setSidebarOpen(true)}
            onSync={cal.syncNow}
            isSyncing={cal.isSyncing}
          />
          {cal.isSyncing || cal.isFetching ? (
            <div className={styles.syncBar} aria-hidden>
              <div className={styles.syncBarFill} />
            </div>
          ) : null}

          <div className={styles.body}>
            <aside className={styles.calSidebar} aria-label="Calendar filters">
              <CalendarSidebar
                cursor={cal.cursor}
                onSelectDay={cal.selectDay}
                onMonthShift={(dir) =>
                  cal.setCursor((c) => addMonths(c, dir))
                }
                accounts={cal.accountGroups}
                onToggleAccount={cal.toggleAccount}
              />
            </aside>

            {cal.sidebarOpen ? (
              <>
                <button
                  type="button"
                  className={styles.mobileBackdrop}
                  aria-label="Close calendars panel"
                  onClick={() => cal.setSidebarOpen(false)}
                />
                <aside
                  className={styles.calSidebarMobile}
                  aria-label="Calendar filters"
                >
                  <CalendarSidebar
                    cursor={cal.cursor}
                    onSelectDay={cal.selectDay}
                    onMonthShift={(dir) =>
                      cal.setCursor((c) => addMonths(c, dir))
                    }
                    accounts={cal.accountGroups}
                    onToggleAccount={cal.toggleAccount}
                  />
                </aside>
              </>
            ) : null}

            <div className={styles.main}>
              <div className={styles.viewPane}>
                {cal.isLoading ? <CalendarSkeleton /> : null}
                {!cal.isLoading && cal.isError ? (
                  <CalendarErrorState
                    message={cal.errorMessage || "Something went wrong."}
                    onRetry={cal.refetch}
                  />
                ) : null}
                {!cal.isLoading && !cal.isError && cal.sources.length === 0 ? (
                  <CalendarEmptySources />
                ) : null}
                {!cal.isLoading && !cal.isError && cal.sources.length > 0 ? (
                  <>
                    {cal.view === "day" ? (
                      <DayView
                        day={cal.cursor}
                        allDay={dayBundle.allDay}
                        timed={dayBundle.timed}
                        colorFor={cal.colorFor}
                        onEventClick={cal.openEvent}
                      />
                    ) : null}
                    {cal.view === "week" ? (
                      <WeekView
                        cursor={cal.cursor}
                        days={weekBundles}
                        colorFor={cal.colorFor}
                        onEventClick={cal.openEvent}
                      />
                    ) : null}
                    {cal.view === "month" ? (
                      <MonthView
                        cursor={cal.cursor}
                        events={cal.events}
                        colorFor={cal.colorFor}
                        onSelectDay={cal.selectDay}
                        onEventClick={cal.openEvent}
                      />
                    ) : null}
                    {cal.view === "agenda" ? (
                      <AgendaView
                        events={cal.events}
                        colorFor={cal.colorFor}
                        onEventClick={cal.openEvent}
                        onNearEnd={cal.extendAgenda}
                      />
                    ) : null}
                  </>
                ) : null}
              </div>
            </div>
          </div>
        </main>
      </div>

      <EventDrawer
        event={cal.selectedEvent}
        source={
          cal.selectedEvent
            ? cal.sourceById(cal.selectedEvent.calendar_source_id)
            : undefined
        }
        color={
          cal.selectedEvent
            ? cal.colorFor(cal.selectedEvent.calendar_source_id)
            : "#c9a87c"
        }
        onClose={cal.closeEvent}
      />
    </div>
  );
}
