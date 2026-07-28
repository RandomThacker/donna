"use client";

import { format } from "date-fns";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useSyncExternalStore } from "react";

import { useCalendarDayEvents } from "@/features/calendar/useCalendarDayEvents";
import { fetchTaskDay } from "@/features/tasks/Tasks.api";
import { taskQueryKeys } from "@/features/tasks/Tasks.logic";

import type { DashboardGreeting } from "./Dashboard.types";

/** Local civil-hour greeting. Late night (before 5) stays evening. */
export function salutationForHour(hour: number): string {
  if (hour >= 17 || hour < 5) {
    return "Good evening";
  }
  if (hour >= 12) {
    return "Good afternoon";
  }
  return "Good morning";
}

export function emojiForHour(hour: number): string {
  if (hour >= 17 || hour < 5) {
    return "🌙";
  }
  if (hour >= 12) {
    return "";
  }
  return "☀️";
}

function plural(count: number, singular: string, pluralForm: string): string {
  return `${count} ${count === 1 ? singular : pluralForm}`;
}

/** Meetings = remaining calendar events today. Tasks = incomplete journal tasks today. */
export function buildGreetingSummary(
  meetingCount: number,
  taskCount: number,
): string {
  return `${plural(meetingCount, "meeting", "meetings")} · ${plural(taskCount, "task", "tasks")}`;
}

function subscribeLocalHour(onStoreChange: () => void): () => void {
  const id = window.setInterval(onStoreChange, 60_000);
  return () => window.clearInterval(id);
}

function getLocalHour(): number {
  return new Date().getHours();
}

/** Sentinel so SSR never freezes a server timezone into the greeting. */
function getServerHour(): number {
  return -1;
}

/** Browser-local hour; null until the client store is active. */
function useLocalHour(): number | null {
  const hour = useSyncExternalStore(
    subscribeLocalHour,
    getLocalHour,
    getServerHour,
  );
  return hour < 0 ? null : hour;
}

/** Live greeting strip: salutation + meeting/task counts. */
export function useDashboardGreeting(name: string): {
  greeting: DashboardGreeting;
  isLoading: boolean;
} {
  const today = useMemo(() => new Date(), []);
  const dateKey = format(today, "yyyy-MM-dd");
  const localHour = useLocalHour();
  const { events, isLoading: eventsLoading } = useCalendarDayEvents(today);

  const tasksQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const greeting = useMemo(() => {
    const now = new Date();
    const hour = localHour;
    const salutation =
      hour === null ? "Hello" : salutationForHour(hour);
    const emoji = hour === null ? "" : emojiForHour(hour);

    const upcomingMeetings = events.filter((event) => {
      if (event.all_day) {
        return true;
      }
      return new Date(event.end_time).getTime() > now.getTime();
    });

    const pendingTasks =
      tasksQuery.data?.occurrences.filter((occurrence) => !occurrence.completed) ??
      [];

    return {
      salutation,
      name,
      emoji,
      summary: buildGreetingSummary(upcomingMeetings.length, pendingTasks.length),
      nudge: "",
    } satisfies DashboardGreeting;
  }, [events, localHour, name, tasksQuery.data?.occurrences]);

  return {
    greeting,
    isLoading: eventsLoading || tasksQuery.isLoading,
  };
}
