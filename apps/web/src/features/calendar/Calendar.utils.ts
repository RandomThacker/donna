import {
  addDays,
  addMonths,
  addWeeks,
  endOfDay,
  endOfMonth,
  endOfWeek,
  format,
  formatDistanceToNow,
  isSameMonth,
  isToday,
  parseISO,
  startOfDay,
  startOfMonth,
  startOfWeek,
  subDays,
  subMonths,
  subWeeks,
} from "date-fns";

import type { CalendarEvent, CalendarView } from "./Calendar.types";
import {
  endOfZonedDay,
  formatZonedTime,
  resolveCalendarTimeZone,
  startOfZonedDay,
} from "./Calendar.timezone";

export const calendarQueryKeys = {
  all: ["calendar"] as const,
  freshness: ["calendar", "freshness"] as const,
  // v2: bust stale caches after multi-account calendar connect fixes
  sources: ["calendar", "sources", "v2"] as const,
  events: (from: string, to: string) =>
    ["calendar", "events", "v2", from, to] as const,
};

export function startOfViewWeek(date: Date): Date {
  return startOfWeek(date, { weekStartsOn: 0 });
}

export function endOfViewWeek(date: Date): Date {
  return endOfWeek(date, { weekStartsOn: 0 });
}

/** Inclusive query window for the active view, with a small buffer. */
export function queryRangeForView(
  view: CalendarView,
  cursor: Date,
  agendaHorizonDays = 60,
  timeZone = resolveCalendarTimeZone(null),
): { from: Date; to: Date } {
  switch (view) {
    case "day": {
      const from = subDays(startOfDay(cursor), 1);
      const to = addDays(endOfDay(cursor), 1);
      return { from, to };
    }
    case "week": {
      const from = subDays(startOfViewWeek(cursor), 1);
      const to = addDays(endOfViewWeek(cursor), 1);
      return { from, to };
    }
    case "month": {
      const gridStart = startOfViewWeek(startOfMonth(cursor));
      const gridEnd = endOfViewWeek(endOfMonth(cursor));
      return { from: subDays(gridStart, 1), to: addDays(gridEnd, 1) };
    }
    case "agenda": {
      // Agenda starts at the cursor's civil day in the calendar timezone (IST by default).
      const from = startOfZonedDay(cursor, timeZone);
      const to = endOfZonedDay(addDays(cursor, agendaHorizonDays), timeZone);
      return { from, to };
    }
  }
}

export function navigateCursor(
  view: CalendarView,
  cursor: Date,
  direction: -1 | 1,
): Date {
  switch (view) {
    case "day":
    case "agenda":
      return direction === 1 ? addDays(cursor, 1) : subDays(cursor, 1);
    case "week":
      return direction === 1 ? addWeeks(cursor, 1) : subWeeks(cursor, 1);
    case "month":
      return direction === 1 ? addMonths(cursor, 1) : subMonths(cursor, 1);
  }
}

export function titleForView(view: CalendarView, cursor: Date): string {
  switch (view) {
    case "day":
      return format(cursor, "EEEE, MMMM d, yyyy");
    case "week": {
      const start = startOfViewWeek(cursor);
      const end = endOfViewWeek(cursor);
      if (isSameMonth(start, end)) {
        return `${format(start, "MMM d")} – ${format(end, "d, yyyy")}`;
      }
      return `${format(start, "MMM d")} – ${format(end, "MMM d, yyyy")}`;
    }
    case "month":
      return format(cursor, "MMMM yyyy");
    case "agenda":
      return format(cursor, "MMMM yyyy");
  }
}

export function formatEventTime(
  event: CalendarEvent,
  timeZone = resolveCalendarTimeZone(null),
): string {
  if (event.all_day) {
    return "All day";
  }
  const start = parseISO(event.start_time);
  const end = parseISO(event.end_time);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return "Time unavailable";
  }
  return `${formatZonedTime(start, timeZone)} – ${formatZonedTime(end, timeZone)}`;
}

export function formatEventDate(event: CalendarEvent): string {
  const start = parseISO(event.start_time);
  if (event.all_day) {
    return format(start, "EEEE, MMMM d, yyyy");
  }
  return format(start, "EEEE, MMMM d, yyyy");
}

export function relativeSyncLabel(iso?: string): string {
  if (!iso) {
    return "Never synced";
  }
  try {
    return `Synced ${formatDistanceToNow(parseISO(iso), { addSuffix: true })}`;
  } catch {
    return "Synced recently";
  }
}

export function parseOrganizer(
  value: unknown,
): { email?: string; displayName?: string } | null {
  if (!value || typeof value !== "object") {
    return null;
  }
  const obj = value as Record<string, unknown>;
  const email = typeof obj.email === "string" ? obj.email : undefined;
  const displayName =
    typeof obj.displayName === "string"
      ? obj.displayName
      : typeof obj.display_name === "string"
        ? obj.display_name
        : undefined;
  if (!email && !displayName) {
    return null;
  }
  return { email, displayName };
}

export function parseAttendees(
  value: unknown,
): Array<{ email?: string; displayName?: string; responseStatus?: string }> {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.map((item) => {
    if (!item || typeof item !== "object") {
      return {};
    }
    const obj = item as Record<string, unknown>;
    return {
      email: typeof obj.email === "string" ? obj.email : undefined,
      displayName:
        typeof obj.displayName === "string"
          ? obj.displayName
          : typeof obj.display_name === "string"
            ? obj.display_name
            : undefined,
      responseStatus:
        typeof obj.responseStatus === "string"
          ? obj.responseStatus
          : typeof obj.response_status === "string"
            ? obj.response_status
            : undefined,
    };
  });
}

export function buildMonthGrid(cursor: Date): Date[] {
  const start = startOfViewWeek(startOfMonth(cursor));
  const end = endOfViewWeek(endOfMonth(cursor));
  const days: Date[] = [];
  let current = start;
  while (current <= end) {
    days.push(current);
    current = addDays(current, 1);
  }
  return days;
}

export function weekDays(cursor: Date): Date[] {
  const start = startOfViewWeek(cursor);
  return Array.from({ length: 7 }, (_, i) => addDays(start, i));
}

export { isToday, format, parseISO, startOfDay, endOfDay, addDays };
