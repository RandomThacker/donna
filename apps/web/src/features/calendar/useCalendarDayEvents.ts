"use client";

import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { useAuth } from "@/features/auth";

import { agendaEventsForDay } from "./Calendar.agenda";
import { listCalendarEvents, listCalendarSources } from "./Calendar.api";
import { dedupeHolidayEvents } from "./Calendar.layout";
import type { CalendarEvent } from "./Calendar.types";
import { endOfZonedDay, resolveCalendarTimeZone, startOfZonedDay } from "./Calendar.timezone";
import { calendarQueryKeys } from "./Calendar.utils";

function normalizeEvents(raw: CalendarEvent[]): CalendarEvent[] {
  return raw.filter((event) => event.status !== "cancelled");
}

/** Donna DB events for one civil day — same filtering as agenda. */
export function useCalendarDayEvents(day = new Date()) {
  const { user } = useAuth();
  const timeZone = resolveCalendarTimeZone(user?.timezone);

  const range = useMemo(() => {
    const from = startOfZonedDay(day, timeZone);
    const to = endOfZonedDay(day, timeZone);
    return { from, to };
  }, [day, timeZone]);

  const fromIso = range.from.toISOString();
  const toIso = range.to.toISOString();

  const sourcesQuery = useQuery({
    queryKey: calendarQueryKeys.sources,
    queryFn: ({ signal }) => listCalendarSources(signal),
    staleTime: 60_000,
  });

  const eventsQuery = useQuery({
    queryKey: calendarQueryKeys.events(fromIso, toIso),
    queryFn: ({ signal }) =>
      listCalendarEvents({ from: fromIso, to: toIso, signal }),
    staleTime: 60_000,
  });

  const enabledSources = useMemo(() => {
    const sources = sourcesQuery.data?.sources ?? [];
    return sources.filter((source) => source.sync_enabled);
  }, [sourcesQuery.data?.sources]);

  const events = useMemo(() => {
    const all = normalizeEvents(eventsQuery.data?.events ?? []);
    const visibleSourceIds = new Set(enabledSources.map((source) => source.id));
    const visible = all.filter((event) =>
      visibleSourceIds.has(event.calendar_source_id),
    );
    const sourcesById = new Map(
      enabledSources.map((source) => [source.id, source] as const),
    );
    const deduped = dedupeHolidayEvents(visible, sourcesById);
    return agendaEventsForDay(deduped, day, timeZone);
  }, [eventsQuery.data?.events, enabledSources, day, timeZone]);

  return {
    events,
    timeZone,
    isLoading: sourcesQuery.isLoading || eventsQuery.isLoading,
    isError: sourcesQuery.isError || eventsQuery.isError,
  };
}
