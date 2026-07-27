import { parseISO } from "date-fns";

import type { CalendarEvent } from "@/features/calendar/Calendar.types";
import { formatZonedTime } from "@/features/calendar/Calendar.timezone";
import { formatEventTime } from "@/features/calendar/Calendar.utils";

import type { DashboardTimelineItem } from "../../Dashboard.types";

function timelineMeta(event: CalendarEvent, timeZone: string): string {
  if (event.location?.trim()) {
    return event.location.trim();
  }
  if (event.all_day) {
    return "All day";
  }
  return formatEventTime(event, timeZone);
}

export function calendarEventsToTimelineItems(
  events: CalendarEvent[],
  timeZone: string,
): DashboardTimelineItem[] {
  return events.map((event) => ({
    id: event.id,
    time: event.all_day
      ? "All day"
      : formatZonedTime(parseISO(event.start_time), timeZone),
    title: event.title?.trim() || "(No title)",
    meta: timelineMeta(event, timeZone),
  }));
}
