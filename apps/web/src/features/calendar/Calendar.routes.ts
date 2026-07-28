import type { CalendarView } from "./Calendar.types";

const CALENDAR_VIEWS: CalendarView[] = ["day", "week", "month", "agenda"];

export function parseCalendarView(value: string | null): CalendarView | null {
  if (!value) {
    return null;
  }
  return CALENDAR_VIEWS.includes(value as CalendarView)
    ? (value as CalendarView)
    : null;
}

/** Deep link into calendar — agenda by default for dashboard timeline. */
export function calendarHref(options?: {
  view?: CalendarView;
  eventId?: string;
}): string {
  const params = new URLSearchParams();
  if (options?.view) {
    params.set("view", options.view);
  }
  if (options?.eventId) {
    params.set("event", options.eventId);
  }
  const query = params.toString();
  return query ? `/dashboard/calendar?${query}` : "/dashboard/calendar";
}

export function calendarAgendaHref(eventId?: string): string {
  return calendarHref({ view: "agenda", eventId });
}
