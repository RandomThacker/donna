import type { AgendaGroup, CalendarEvent } from "./Calendar.types";
import {
  eventAgendaDateKey,
  isZonedToday,
  zonedDateKey,
} from "./Calendar.timezone";
import { format, parseISO } from "./Calendar.utils";

/** Events on a single civil day (agenda semantics), sorted by start time. */
export function agendaEventsForDay(
  events: CalendarEvent[],
  day: Date,
  timeZone: string,
): CalendarEvent[] {
  const dayKey = zonedDateKey(day, timeZone);
  return [...events]
    .filter((event) => eventAgendaDateKey(event, timeZone) === dayKey)
    .sort(
      (a, b) =>
        new Date(a.start_time).getTime() - new Date(b.start_time).getTime(),
    );
}

/** Group events by civil day from `fromDate` onward (agenda view). */
export function groupAgendaEvents(
  events: CalendarEvent[],
  fromDate: Date,
  timeZone: string,
): AgendaGroup[] {
  const map = new Map<string, AgendaGroup>();
  const fromKey = zonedDateKey(fromDate, timeZone);
  const sorted = [...events].sort(
    (a, b) =>
      new Date(a.start_time).getTime() - new Date(b.start_time).getTime(),
  );

  for (const event of sorted) {
    const key = eventAgendaDateKey(event, timeZone);
    if (!key || key < fromKey) {
      continue;
    }
    let group = map.get(key);
    if (!group) {
      const date = parseISO(`${key}T12:00:00.000Z`);
      const label = isZonedToday(date, timeZone)
        ? `Today · ${format(date, "MMMM d")}`
        : format(date, "EEEE, MMMM d");
      group = { key, date, label, events: [] };
      map.set(key, group);
    }
    group.events.push(event);
  }

  return Array.from(map.values());
}
