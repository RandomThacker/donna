"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";

import { useAuth } from "@/features/auth";

import {
  listCalendarEvents,
  listCalendarSources,
  syncCalendarSources,
} from "./Calendar.api";
import {
  allDayEventsForDay,
  colorForSource,
  dedupeHolidayEvents,
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
import { resolveCalendarTimeZone } from "./Calendar.timezone";
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
  if (account?.email?.trim()) {
    return account.email.trim();
  }
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
  if (primary?.name?.trim()) {
    return primary.name.trim();
  }
  return sources[0]?.name?.trim() || "Calendar";
}

function buildAccountGroups(
  sources: CalendarSource[],
  accounts: CalendarConnectedAccount[],
  colorFor: (sourceId: string) => string,
  visibleSourceIds: Set<string>,
): CalendarAccountGroup[] {
  const byAccount = new Map<string, CalendarSource[]>();
  for (const source of sources) {
    const key = source.connected_account_id;
    if (!key) {
      continue;
    }
    const list = byAccount.get(key) ?? [];
    list.push(source);
    byAccount.set(key, list);
  }

  const accountById = new Map(accounts.map((a) => [a.id, a]));
  const orderedIds: string[] = [];
  for (const account of accounts) {
    orderedIds.push(account.id);
  }
  for (const accountId of byAccount.keys()) {
    if (!accountById.has(accountId)) {
      orderedIds.push(accountId);
    }
  }

  return orderedIds.map((accountId) => {
    const matchedAccount = accountById.get(accountId);
    const groupSources = byAccount.get(accountId) ?? [];
    const primary =
      groupSources.find((s) => s.is_primary_on_provider) ?? groupSources[0];
    const sourceIds = groupSources.map((s) => s.id);
    const label =
      matchedAccount?.email?.trim() ||
      accountLabel(groupSources, matchedAccount);
    const email =
      matchedAccount?.email?.trim() ||
      (EMAIL_RE.test(label) ? label : null) ||
      groupSources.find((s) => EMAIL_RE.test(s.provider_calendar_id.trim()))
        ?.provider_calendar_id ||
      null;
    return {
      accountId,
      label,
      email,
      color: primary ? colorFor(primary.id) : "#c9a87c",
      sourceIds,
      visibleCount: sourceIds.filter((id) => visibleSourceIds.has(id)).length,
    };
  });
}

export function useCalendarController() {
  const { user } = useAuth();
  const timeZone = resolveCalendarTimeZone(user?.timezone);
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
    () => queryRangeForView(view, cursor, agendaDays, timeZone),
    [view, cursor, agendaDays, timeZone],
  );
  const fromIso = range.from.toISOString();
  const toIso = range.to.toISOString();

  const sourcesQuery = useQuery({
    queryKey: calendarQueryKeys.sources,
    queryFn: ({ signal }) => listCalendarSources(signal),
    staleTime: 0,
    refetchOnMount: "always",
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
    const visible = all.filter((e) => visibleSourceIds.has(e.calendar_source_id));
    const sourcesById = new Map(
      enabledSources.map((source) => [source.id, source] as const),
    );
    return dedupeHolidayEvents(visible, sourcesById);
  }, [eventsQuery.data?.events, visibleSourceIds, enabledSources]);

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

  const connectedAccounts = useMemo(() => {
    if (sourcesQuery.data?.accounts?.length) {
      return sourcesQuery.data.accounts;
    }
    if (sourcesQuery.data?.account) {
      return [sourcesQuery.data.account];
    }
    return [];
  }, [sourcesQuery.data?.accounts, sourcesQuery.data?.account]);

  const accountGroups = useMemo(
    () =>
      buildAccountGroups(
        enabledSources,
        connectedAccounts,
        (id) => sourceColorMap.get(id) ?? "#c9a87c",
        visibleSourceIds,
      ),
    [enabledSources, connectedAccounts, sourceColorMap, visibleSourceIds],
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

  const calendarLabelFor = useCallback(
    (sourceId: string): string => {
      const group = accountGroups.find((g) => g.sourceIds.includes(sourceId));
      if (group?.label?.trim()) {
        return group.label.trim();
      }
      return sourceById(sourceId)?.name?.trim() || "Unknown calendar";
    },
    [accountGroups, sourceById],
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
    timeZone,
    sourceById,
    calendarLabelFor,
    colorFor: (sourceId: string) =>
      sourceColorMap.get(sourceId) ?? "#c9a87c",
  };
}

export type CalendarController = ReturnType<typeof useCalendarController>;
