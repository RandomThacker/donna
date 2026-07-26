"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";

import {
  listCalendarEvents,
  listCalendarSources,
  syncCalendarSources,
} from "./Calendar.api";
import {
  allDayEventsForDay,
  colorForSource,
  eventsOverlappingDay,
  layoutTimedEvents,
} from "./Calendar.layout";
import type {
  CalendarAccountGroup,
  CalendarConnectedAccount,
  CalendarEvent,
  CalendarSource,
  CalendarView,
} from "./Calendar.types";
import {
  calendarQueryKeys,
  navigateCursor,
  queryRangeForView,
  titleForView,
} from "./Calendar.utils";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function normalizeEvents(raw: CalendarEvent[]): CalendarEvent[] {
  return raw.filter((e) => e.status !== "cancelled");
}

function accountLabel(
  sources: CalendarSource[],
  account?: CalendarConnectedAccount,
): string {
  const primary = sources.find((s) => s.is_primary_on_provider);
  if (primary?.name && EMAIL_RE.test(primary.name.trim())) {
    return primary.name.trim();
  }
  const emailNamed = sources.find((s) => EMAIL_RE.test(s.name.trim()));
  if (emailNamed) {
    return emailNamed.name.trim();
  }
  if (account?.display_name?.trim()) {
    return account.display_name.trim();
  }
  if (primary?.name) {
    return primary.name;
  }
  return sources[0]?.name ?? "Google Calendar";
}

function groupSourcesByAccount(
  sources: CalendarSource[],
  account: CalendarConnectedAccount | undefined,
  colorFor: (sourceId: string) => string,
  visibleSourceIds: Set<string>,
): CalendarAccountGroup[] {
  const byAccount = new Map<string, CalendarSource[]>();
  for (const source of sources) {
    const key = source.connected_account_id;
    const list = byAccount.get(key) ?? [];
    list.push(source);
    byAccount.set(key, list);
  }

  return Array.from(byAccount.entries()).map(([accountId, groupSources]) => {
    const matchedAccount =
      account?.id === accountId ? account : undefined;
    const primary =
      groupSources.find((s) => s.is_primary_on_provider) ?? groupSources[0]!;
    const sourceIds = groupSources.map((s) => s.id);
    return {
      accountId,
      label: accountLabel(groupSources, matchedAccount),
      color: colorFor(primary.id),
      sourceIds,
      visibleCount: sourceIds.filter((id) => visibleSourceIds.has(id)).length,
    };
  });
}

export function useCalendarController() {
  const queryClient = useQueryClient();
  const [view, setView] = useState<CalendarView>("day");
  const [cursor, setCursor] = useState(() => new Date());
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [hiddenSourceIds, setHiddenSourceIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [agendaDays, setAgendaDays] = useState(60);

  const range = useMemo(
    () => queryRangeForView(view, cursor, agendaDays),
    [view, cursor, agendaDays],
  );
  const fromIso = range.from.toISOString();
  const toIso = range.to.toISOString();

  const sourcesQuery = useQuery({
    queryKey: calendarQueryKeys.sources,
    queryFn: ({ signal }) => listCalendarSources(signal),
  });

  const eventsQuery = useQuery({
    queryKey: calendarQueryKeys.events(fromIso, toIso),
    queryFn: ({ signal }) =>
      listCalendarEvents({ from: fromIso, to: toIso, signal }),
  });

  const syncMutation = useMutation({
    mutationFn: () => syncCalendarSources(),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: calendarQueryKeys.all });
    },
  });

  const enabledSources = useMemo(() => {
    const sources = sourcesQuery.data?.sources ?? [];
    return sources.filter((s) => s.sync_enabled);
  }, [sourcesQuery.data?.sources]);

  const sourceColorMap = useMemo(() => {
    const map = new Map<string, string>();
    enabledSources.forEach((source, index) => {
      map.set(source.id, colorForSource(source.id, source.color, index));
    });
    return map;
  }, [enabledSources]);

  const visibleSourceIds = useMemo(() => {
    return new Set(
      enabledSources
        .filter((s) => !hiddenSourceIds.has(s.id))
        .map((s) => s.id),
    );
  }, [enabledSources, hiddenSourceIds]);

  const events = useMemo(() => {
    const all = normalizeEvents(eventsQuery.data?.events ?? []);
    return all.filter((e) => visibleSourceIds.has(e.calendar_source_id));
  }, [eventsQuery.data?.events, visibleSourceIds]);

  const selectedEvent = useMemo(
    () => events.find((e) => e.id === selectedEventId) ?? null,
    [events, selectedEventId],
  );

  const title = useMemo(() => titleForView(view, cursor), [view, cursor]);

  const goToday = useCallback(() => setCursor(new Date()), []);
  const goPrev = useCallback(
    () => setCursor((c) => navigateCursor(view, c, -1)),
    [view],
  );
  const goNext = useCallback(
    () => setCursor((c) => navigateCursor(view, c, 1)),
    [view],
  );

  const selectDay = useCallback((day: Date) => {
    setCursor(day);
    setView("day");
    setSidebarOpen(false);
  }, []);

  const toggleSource = useCallback((sourceId: string) => {
    setHiddenSourceIds((prev) => {
      const next = new Set(prev);
      if (next.has(sourceId)) {
        next.delete(sourceId);
      } else {
        next.add(sourceId);
      }
      return next;
    });
  }, []);

  const toggleAccount = useCallback(
    (sourceIds: string[]) => {
      setHiddenSourceIds((prev) => {
        const next = new Set(prev);
        const allVisible = sourceIds.every((id) => !next.has(id));
        if (allVisible) {
          for (const id of sourceIds) {
            next.add(id);
          }
        } else {
          for (const id of sourceIds) {
            next.delete(id);
          }
        }
        return next;
      });
    },
    [],
  );

  const accountGroups = useMemo(
    () =>
      groupSourcesByAccount(
        enabledSources,
        sourcesQuery.data?.account,
        (id) => sourceColorMap.get(id) ?? "#c9a87c",
        visibleSourceIds,
      ),
    [enabledSources, sourcesQuery.data?.account, sourceColorMap, visibleSourceIds],
  );
  const openEvent = useCallback((event: CalendarEvent) => {
    setSelectedEventId(event.id);
  }, []);

  const closeEvent = useCallback(() => setSelectedEventId(null), []);

  const extendAgenda = useCallback(() => {
    setAgendaDays((d) => d + 30);
  }, []);

  const dayLayout = useCallback(
    (day: Date) => ({
      allDay: allDayEventsForDay(events, day),
      timed: layoutTimedEvents(eventsOverlappingDay(events, day), day),
    }),
    [events],
  );

  const sourceById = useCallback(
    (id: string): CalendarSource | undefined =>
      enabledSources.find((s) => s.id === id),
    [enabledSources],
  );

  return {
    view,
    setView,
    cursor,
    setCursor,
    title,
    goToday,
    goPrev,
    goNext,
    selectDay,
    events,
    dayLayout,
    sources: enabledSources,
    accountGroups,
    sourceColorMap,
    visibleSourceIds,
    toggleSource,
    toggleAccount,
    sync: sourcesQuery.data?.sync,
    isLoading: sourcesQuery.isLoading || eventsQuery.isLoading,
    isFetching: eventsQuery.isFetching || sourcesQuery.isFetching,
    isError: sourcesQuery.isError || eventsQuery.isError,
    errorMessage:
      (sourcesQuery.error as Error | null)?.message ||
      (eventsQuery.error as Error | null)?.message ||
      null,
    refetch: () => {
      void sourcesQuery.refetch();
      void eventsQuery.refetch();
    },
    syncNow: () => syncMutation.mutate(),
    isSyncing: syncMutation.isPending,
    syncError: (syncMutation.error as Error | null)?.message ?? null,
    lastSyncResult: syncMutation.data ?? null,
    selectedEvent,
    openEvent,
    closeEvent,
    sidebarOpen,
    setSidebarOpen,
    extendAgenda,
    sourceById,
    colorFor: (sourceId: string) =>
      sourceColorMap.get(sourceId) ?? "#c9a87c",
  };
}

export type CalendarController = ReturnType<typeof useCalendarController>;
