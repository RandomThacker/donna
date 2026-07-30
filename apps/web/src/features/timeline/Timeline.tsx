"use client";

import { Suspense, useEffect, useState } from "react";
import { usePathname, useSearchParams } from "next/navigation";

import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useTimelineController } from "./Timeline.logic";
import { timelineStyles as styles } from "./Timeline.styles";
import { CreateChooserModal } from "./sections/CreateChooserModal";
import { EventFormModal } from "./sections/EventFormModal";
import { ItemDetailsModal } from "./sections/ItemDetailsModal";
import { ReminderFormModal } from "./sections/ReminderFormModal";
import { TimelineFilters } from "./sections/TimelineFilters";
import { TimelineToolbar } from "./sections/TimelineToolbar";
import {
  AgendaView,
  DayView,
  MonthView,
  TimelineError,
  TimelineSkeleton,
  WeekView,
} from "./sections/TimelineViews";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "D";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

function TimelineApp() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { user } = useAuth();
  const tl = useTimelineController();
  const nav = navItemsForPath(pathname);
  const [chooserOpen, setChooserOpen] = useState(false);
  const tz =
    user?.timezone?.trim() ||
    Intl.DateTimeFormat().resolvedOptions().timeZone ||
    "UTC";

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  useEffect(() => {
    const occurrence = searchParams.get("occurrence");
    if (occurrence) {
      tl.openOccurrence(occurrence);
    }
  }, [searchParams, tl.openOccurrence]);

  const editingEvent =
    tl.editing && tl.editing.type === "EVENT" ? tl.editing : null;
  const editingReminder =
    tl.editing && tl.editing.type === "REMINDER" ? tl.editing : null;

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
          <TimelineToolbar
            title={tl.title}
            view={tl.view}
            onView={tl.setView}
            onToday={tl.goToday}
            onPrev={tl.goPrev}
            onNext={tl.goNext}
            onMenu={() => tl.setSidebarOpen(true)}
            onCreate={() => {
              tl.openCreate(tl.cursor, null);
              setChooserOpen(true);
            }}
            search={tl.search}
            onSearch={tl.setSearch}
            fetching={tl.isFetching}
          />
          <div className={styles.body}>
            {tl.sidebarOpen ? (
              <button
                type="button"
                className={styles.backdrop}
                aria-label="Close filters"
                onClick={() => tl.setSidebarOpen(false)}
              />
            ) : null}
            <aside
              className={tl.sidebarOpen ? styles.sideMobile : styles.side}
            >
              <TimelineFilters
                cursor={tl.cursor}
                onSelectDay={(day) => {
                  tl.setCursor(day);
                  tl.setView("day");
                  tl.setSidebarOpen(false);
                }}
                filters={tl.filters}
                onToggle={tl.toggleFilter}
              />
            </aside>
            <div className={styles.main}>
              <div className={styles.viewPane}>
                {tl.isLoading ? <TimelineSkeleton /> : null}
                {tl.isError ? (
                  <TimelineError
                    message={tl.errorMessage || "Couldn’t load timeline"}
                    onRetry={tl.refetch}
                  />
                ) : null}
                {!tl.isLoading && !tl.isError ? (
                  <>
                    {tl.view === "month" ? (
                      <MonthView
                        cursor={tl.cursor}
                        items={tl.items}
                        onOpen={tl.openItem}
                        onCreateDay={(day) => {
                          tl.openCreate(day, null);
                          setChooserOpen(true);
                        }}
                      />
                    ) : null}
                    {tl.view === "week" ? (
                      <WeekView
                        cursor={tl.cursor}
                        dayLayout={tl.dayLayout}
                        onOpen={tl.openItem}
                        onCreateDay={(day) => {
                          tl.openCreate(day, null);
                          setChooserOpen(true);
                        }}
                      />
                    ) : null}
                    {tl.view === "day" ? (
                      <DayView
                        cursor={tl.cursor}
                        dayLayout={tl.dayLayout}
                        onOpen={tl.openItem}
                        onCreateDay={(day) => {
                          tl.openCreate(day, null);
                          setChooserOpen(true);
                        }}
                      />
                    ) : null}
                    {tl.view === "agenda" ? (
                      <AgendaView
                        cursor={tl.cursor}
                        items={tl.items}
                        onOpen={tl.openItem}
                      />
                    ) : null}
                  </>
                ) : null}
              </div>
            </div>
          </div>
        </main>
      </div>

      <CreateChooserModal
        open={chooserOpen}
        onClose={() => setChooserOpen(false)}
        onEvent={() => {
          setChooserOpen(false);
          tl.setCreateIntent("event");
        }}
        onReminder={() => {
          setChooserOpen(false);
          tl.setCreateIntent("reminder");
        }}
      />

      <EventFormModal
        open={tl.createIntent === "event" || Boolean(editingEvent)}
        onClose={() => {
          tl.setCreateIntent(null);
          tl.setEditing(null);
        }}
        day={tl.createDay}
        editing={editingEvent}
        timezone={tz}
        saving={
          tl.createEventMutation.isPending || tl.updateEventMutation.isPending
        }
        onCreate={async (body) => {
          await tl.createEventMutation.mutateAsync(body);
        }}
        onUpdate={async (id, body) => {
          await tl.updateEventMutation.mutateAsync({ id, body });
        }}
      />

      <ReminderFormModal
        open={tl.createIntent === "reminder" || Boolean(editingReminder)}
        onClose={() => {
          tl.setCreateIntent(null);
          tl.setEditing(null);
        }}
        day={tl.createDay}
        editing={editingReminder}
        timezone={tz}
        saving={
          tl.createReminderMutation.isPending ||
          tl.updateReminderMutation.isPending
        }
        onCreate={async (body) => {
          await tl.createReminderMutation.mutateAsync(body);
        }}
        onUpdate={async (id, body) => {
          await tl.updateReminderMutation.mutateAsync({ id, body });
        }}
      />

      <ItemDetailsModal
        item={tl.selected}
        onClose={() => tl.setSelected(null)}
        onEdit={(item) => tl.startEdit(item)}
        onDelete={(item) => void tl.removeItem(item)}
        deleting={
          tl.deleteEventMutation.isPending ||
          tl.deleteReminderMutation.isPending
        }
      />
    </div>
  );
}

export function Timeline() {
  return (
    <Suspense fallback={<div className={styles.page} />}>
      <TimelineApp />
    </Suspense>
  );
}
