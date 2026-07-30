"use client";

import { addMonths } from "date-fns";
import { useEffect, useMemo, useRef, useState } from "react";
import { usePathname, useSearchParams } from "next/navigation";

import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";
import { CreateChooserModal } from "@/features/timeline/sections/CreateChooserModal";
import { EventFormModal } from "@/features/timeline/sections/EventFormModal";
import { ReminderFormModal } from "@/features/timeline/sections/ReminderFormModal";
import type { TimelineItem } from "@/features/timeline/Timeline.types";

import { useCalendarController } from "./Calendar.logic";
import { parseCalendarView } from "./Calendar.routes";
import { calendarStyles as styles } from "./Calendar.styles";
import type { CalendarEvent } from "./Calendar.types";
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
  if (parts.length === 0) return "D";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

function toTimelineEdit(event: CalendarEvent): TimelineItem {
  return {
    id: event.mutation_id || event.id,
    source: "DONNA",
    type: event.timeline_type === "REMINDER" ? "REMINDER" : "EVENT",
    status: "ACTIVE",
    title: event.title,
    description: event.description,
    start_at: event.start_time,
    end_at: event.end_time,
    timezone: event.timezone || "UTC",
    all_day: event.all_day,
    read_only: false,
    is_recurring: Boolean(event.recurrence_rule),
    recurrence_rule: event.recurrence_rule,
    parent_id: event.mutation_id,
    occurrence_id: event.occurrence_id || event.id,
  };
}

export function Calendar() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { user } = useAuth();
  const cal = useCalendarController();
  const appliedEventRef = useRef<string | null>(null);
  const [chooserOpen, setChooserOpen] = useState(false);

  const tz =
    user?.timezone?.trim() ||
    Intl.DateTimeFormat().resolvedOptions().timeZone ||
    "UTC";

  useEffect(() => {
    const view = parseCalendarView(searchParams.get("view"));
    if (view) cal.setView(view);
  }, [searchParams, cal.setView]);

  useEffect(() => {
    const eventId = searchParams.get("event");
    if (!eventId) {
      appliedEventRef.current = null;
      return;
    }
    if (cal.isLoading || appliedEventRef.current === eventId) return;
    const event =
      cal.events.find((item) => item.id === eventId) ??
      cal.events.find((item) => item.occurrence_id === eventId);
    if (event) {
      appliedEventRef.current = eventId;
      cal.openEvent(event);
    }
  }, [searchParams, cal.events, cal.isLoading, cal.openEvent]);

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

  const editingTimeline = cal.editingEvent
    ? toTimelineEdit(cal.editingEvent)
    : null;

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={nav}
          profileName={profileName}
          profileInitials={profileInitials}
          profileEmail={user?.email}
          profileAvatarUrl={user?.avatar_url}
        />

        <main className={styles.workspace}>
          <CalendarToolbar
            title={cal.title}
            cursor={cal.cursor}
            view={cal.view}
            onViewChange={cal.setView}
            onPrev={cal.goPrev}
            onNext={cal.goNext}
            onToday={cal.goToday}
            onOpenSidebar={() => cal.setSidebarOpen(true)}
            onSync={cal.syncNow}
            isSyncing={cal.isSyncing}
            onCreate={() => {
              cal.openCreate(cal.cursor);
              setChooserOpen(true);
            }}
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
                {!cal.isLoading &&
                !cal.isError &&
                !cal.hasAnySource &&
                cal.events.length === 0 ? (
                  <CalendarEmptySources />
                ) : null}
                {!cal.isLoading && !cal.isError && cal.hasAnySource ? (
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
                        fromDate={cal.cursor}
                        timeZone={cal.timeZone}
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
        calendarLabel={
          cal.selectedEvent
            ? cal.calendarLabelFor(cal.selectedEvent.calendar_source_id)
            : undefined
        }
        color={
          cal.selectedEvent
            ? cal.colorFor(
                cal.selectedEvent.calendar_source_id,
                cal.selectedEvent,
              )
            : "#c9a87c"
        }
        timeZone={cal.timeZone}
        onClose={cal.closeEvent}
        onEdit={cal.startEdit}
        onDelete={(event) => void cal.removeEvent(event)}
        deleting={
          cal.deleteEventMutation.isPending ||
          cal.deleteReminderMutation.isPending
        }
      />

      <CreateChooserModal
        open={chooserOpen}
        onClose={() => setChooserOpen(false)}
        onEvent={() => {
          setChooserOpen(false);
          cal.setCreateIntent("event");
        }}
        onReminder={() => {
          setChooserOpen(false);
          cal.setCreateIntent("reminder");
        }}
      />

      <EventFormModal
        open={
          (cal.createIntent === "event" && !cal.editingEvent) ||
          Boolean(
            cal.editingEvent && cal.editingEvent.timeline_type !== "REMINDER",
          )
        }
        onClose={() => {
          cal.setCreateIntent(null);
          cal.setEditingEvent(null);
        }}
        day={cal.createDay}
        editing={
          editingTimeline && editingTimeline.type === "EVENT"
            ? editingTimeline
            : null
        }
        timezone={tz}
        saving={
          cal.createEventMutation.isPending ||
          cal.updateEventMutation.isPending
        }
        onCreate={async (body) => {
          await cal.createEventMutation.mutateAsync(body);
          cal.setCursor(new Date(body.start_at));
          cal.setCreateIntent(null);
          cal.setCreateDay(null);
        }}
        onUpdate={async (id, body) => {
          await cal.updateEventMutation.mutateAsync({ id, body });
        }}
      />

      <ReminderFormModal
        open={
          (cal.createIntent === "reminder" && !cal.editingEvent) ||
          Boolean(
            cal.editingEvent && cal.editingEvent.timeline_type === "REMINDER",
          )
        }
        onClose={() => {
          cal.setCreateIntent(null);
          cal.setEditingEvent(null);
        }}
        day={cal.createDay}
        editing={
          editingTimeline && editingTimeline.type === "REMINDER"
            ? editingTimeline
            : null
        }
        timezone={tz}
        saving={
          cal.createReminderMutation.isPending ||
          cal.updateReminderMutation.isPending
        }
        onCreate={async (body) => {
          await cal.createReminderMutation.mutateAsync(body);
          cal.setCursor(new Date(body.trigger_at));
          cal.setCreateIntent(null);
          cal.setCreateDay(null);
        }}
        onUpdate={async (id, body) => {
          await cal.updateReminderMutation.mutateAsync({ id, body });
        }}
      />
    </div>
  );
}
