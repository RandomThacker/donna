"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";

import { useAuth } from "@/features/auth";
import {
  createDonnaEvent,
  createDonnaReminder,
  deleteDonnaEvent,
  deleteDonnaReminder,
  fetchTimeline,
  updateDonnaEvent,
  updateDonnaReminder,
} from "@/features/timeline/Timeline.api";
import { TIMELINE_COLORS } from "@/features/timeline/Timeline.colors";
import type {
  CreateDonnaEventInput,
  CreateDonnaReminderInput,
  UpdateDonnaEventInput,
  UpdateDonnaReminderInput,
} from "@/features/timeline/Timeline.types";

import { listCalendarSources, syncCalendarSources } from "./Calendar.api";
import {
  allDayEventsForDay,
  colorForSource,
  dedupeHolidayEvents,
  eventsOverlappingDay,
  layoutTimedEvents,
} from "./Calendar.layout";
import {
  DONNA_EVENT_SOURCE_ID,
  DONNA_REMINDER_SOURCE_ID,
  donnaEventAccountGroup,
  donnaReminderAccountGroup,
  timelineItemToCalendarEvent,
} from "./Calendar.timeline";
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
    if (source.provider_calendar_id === "donna_local") {
      // Virtual Donna Events source — shown via donnaEventAccountGroup instead.
      continue;
    }
    const key = source.connected_account_id;
    if (!key || key === "00000000-0000-0000-0000-000000000000") {
      continue;
    }
    const list = byAccount.get(key) ?? [];
    list.push(source);
    byAccount.set(key, list);
  }

  const accountById = new Map(accounts.map((a) => [a.id, a]));
  for (const accountId of [...byAccount.keys()]) {
    if (!accountById.has(accountId)) {
      byAccount.delete(accountId);
    }
  }

  const groups: CalendarAccountGroup[] = [
    donnaEventAccountGroup(visibleSourceIds.has(DONNA_EVENT_SOURCE_ID)),
    donnaReminderAccountGroup(visibleSourceIds.has(DONNA_REMINDER_SOURCE_ID)),
  ];

  for (const account of accounts) {
    const groupSources = byAccount.get(account.id);
    if (!groupSources?.length) continue;
    const primary =
      groupSources.find((s) => s.is_primary_on_provider) ?? groupSources[0];
    const sourceIds = groupSources.map((s) => s.id);
    const label =
      account.email?.trim() || accountLabel(groupSources, account);
    const email =
      account.email?.trim() ||
      (EMAIL_RE.test(label) ? label : null) ||
      groupSources.find((s) => EMAIL_RE.test(s.provider_calendar_id.trim()))
        ?.provider_calendar_id ||
      null;
    groups.push({
      accountId: account.id,
      label,
      email,
      color: primary ? colorFor(primary.id) : "#c9a87c",
      sourceIds,
      visibleCount: sourceIds.filter((id) => visibleSourceIds.has(id)).length,
    });
  }

  return groups;
}

export type CreateIntent = "event" | "reminder" | null;

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
  const [createIntent, setCreateIntent] = useState<CreateIntent>(null);
  const [createDay, setCreateDay] = useState<Date | null>(null);
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null);

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

  const timelineQuery = useQuery({
    queryKey: [...calendarQueryKeys.all, "timeline", fromIso, toIso],
    queryFn: ({ signal }) =>
      fetchTimeline({ from: fromIso, to: toIso, signal }),
  });

  const syncMutation = useMutation({
    mutationFn: () => syncCalendarSources(),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: calendarQueryKeys.all });
    },
  });

  const invalidate = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: calendarQueryKeys.all });
  }, [queryClient]);

  const createEventMutation = useMutation({
    mutationFn: (body: CreateDonnaEventInput) => createDonnaEvent(body),
    onSuccess: () => void invalidate(),
  });
  const updateEventMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateDonnaEventInput }) =>
      updateDonnaEvent(id, body),
    onSuccess: () => void invalidate(),
  });
  const deleteEventMutation = useMutation({
    mutationFn: (id: string) => deleteDonnaEvent(id),
    onSuccess: () => {
      setSelectedEventId(null);
      setEditingEvent(null);
      void invalidate();
    },
  });
  const createReminderMutation = useMutation({
    mutationFn: (body: CreateDonnaReminderInput) => createDonnaReminder(body),
    onSuccess: () => void invalidate(),
  });
  const updateReminderMutation = useMutation({
    mutationFn: ({
      id,
      body,
    }: {
      id: string;
      body: UpdateDonnaReminderInput;
    }) => updateDonnaReminder(id, body),
    onSuccess: () => void invalidate(),
  });
  const deleteReminderMutation = useMutation({
    mutationFn: (id: string) => deleteDonnaReminder(id),
    onSuccess: () => {
      setSelectedEventId(null);
      setEditingEvent(null);
      void invalidate();
    },
  });

  const enabledSources = useMemo(() => {
    const sources = sourcesQuery.data?.sources ?? [];
    const liveAccountIds = new Set<string>();
    for (const account of sourcesQuery.data?.accounts ?? []) {
      liveAccountIds.add(account.id);
    }
    if (sourcesQuery.data?.account?.id) {
      liveAccountIds.add(sourcesQuery.data.account.id);
    }
    return sources.filter((source) => {
      if (!source.sync_enabled) {
        return false;
      }
      if (source.provider_calendar_id === "donna_local") {
        return true;
      }
      if (liveAccountIds.size === 0) {
        return true;
      }
      return liveAccountIds.has(source.connected_account_id);
    });
  }, [
    sourcesQuery.data?.sources,
    sourcesQuery.data?.accounts,
    sourcesQuery.data?.account,
  ]);

  const sourceColorMap = useMemo(() => {
    const map = new Map<string, string>();
    enabledSources.forEach((source, index) => {
      if (source.provider_calendar_id === "donna_local") {
        map.set(source.id, TIMELINE_COLORS.donnaEvent);
        return;
      }
      map.set(source.id, colorForSource(source.id, source.color, index, source.provider_calendar_id));
    });
    map.set(DONNA_EVENT_SOURCE_ID, TIMELINE_COLORS.donnaEvent);
    map.set(DONNA_REMINDER_SOURCE_ID, TIMELINE_COLORS.donnaReminder);
    return map;
  }, [enabledSources]);

  const visibleSourceIds = useMemo(() => {
    const ids = new Set(
      enabledSources
        .filter((s) => !hiddenSourceIds.has(s.id))
        .map((s) => s.id),
    );
    // Always include Donna filter keys (stable client IDs).
    if (!hiddenSourceIds.has(DONNA_EVENT_SOURCE_ID)) {
      ids.add(DONNA_EVENT_SOURCE_ID);
    }
    if (!hiddenSourceIds.has(DONNA_REMINDER_SOURCE_ID)) {
      ids.add(DONNA_REMINDER_SOURCE_ID);
    }
    // Also keep API virtual donna_local UUID visible if present.
    for (const source of enabledSources) {
      if (
        source.provider_calendar_id === "donna_local" &&
        !hiddenSourceIds.has(DONNA_EVENT_SOURCE_ID)
      ) {
        ids.add(source.id);
      }
    }
    return ids;
  }, [enabledSources, hiddenSourceIds]);

  const events = useMemo(() => {
    const mapped = (timelineQuery.data?.items ?? []).map((item) =>
      timelineItemToCalendarEvent(item),
    );
    const all = normalizeEvents(mapped);
    const visible = all.filter((e) =>
      visibleSourceIds.has(e.calendar_source_id),
    );
    const sourcesById = new Map(
      enabledSources.map((source) => [source.id, source] as const),
    );
    return dedupeHolidayEvents(visible, sourcesById);
  }, [timelineQuery.data?.items, visibleSourceIds, enabledSources]);

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

  const openCreate = useCallback((day: Date) => {
    setCreateDay(day);
    setCreateIntent(null);
    setSelectedEventId(null);
    setEditingEvent(null);
  }, []);

  const toggleAccount = useCallback((sourceIds: string[]) => {
    setHiddenSourceIds((prev) => {
      const next = new Set(prev);
      const allVisible = sourceIds.every((id) => !next.has(id));
      if (allVisible) {
        for (const id of sourceIds) next.add(id);
      } else {
        for (const id of sourceIds) next.delete(id);
      }
      return next;
    });
  }, []);

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

  const startEdit = useCallback((event: CalendarEvent) => {
    setEditingEvent(event);
    setSelectedEventId(null);
    if (event.timeline_type === "REMINDER") {
      setCreateIntent("reminder");
    } else {
      setCreateIntent("event");
    }
  }, []);

  const removeEvent = useCallback(
    async (event: CalendarEvent) => {
      const id = event.mutation_id;
      if (!id) return;
      if (event.timeline_type === "REMINDER") {
        await deleteReminderMutation.mutateAsync(id);
      } else {
        await deleteEventMutation.mutateAsync(id);
      }
    },
    [deleteEventMutation, deleteReminderMutation],
  );

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
      if (sourceId === DONNA_EVENT_SOURCE_ID) return "Donna Events";
      if (sourceId === DONNA_REMINDER_SOURCE_ID) return "Donna Reminders";
      const group = accountGroups.find((g) => g.sourceIds.includes(sourceId));
      if (group?.label?.trim()) return group.label.trim();
      return sourceById(sourceId)?.name?.trim() || "Unknown calendar";
    },
    [accountGroups, sourceById],
  );

  const colorFor = useCallback(
    (sourceId: string, event?: CalendarEvent) => {
      if (event?.accent_color) return event.accent_color;
      return sourceColorMap.get(sourceId) ?? "#c9a87c";
    },
    [sourceColorMap],
  );

  const hasAnySource =
    enabledSources.length > 0 ||
    (timelineQuery.data?.items?.length ?? 0) > 0 ||
    !sourcesQuery.isLoading;

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
    openCreate,
    createIntent,
    setCreateIntent,
    createDay,
    setCreateDay,
    editingEvent,
    setEditingEvent,
    startEdit,
    removeEvent,
    events,
    dayLayout,
    sources: enabledSources,
    hasAnySource,
    accountGroups,
    toggleAccount,
    sync: sourcesQuery.data?.sync,
    isLoading: sourcesQuery.isLoading || timelineQuery.isLoading,
    isFetching: timelineQuery.isFetching || sourcesQuery.isFetching,
    isError: sourcesQuery.isError || timelineQuery.isError,
    errorMessage:
      (sourcesQuery.error as Error | null)?.message ||
      (timelineQuery.error as Error | null)?.message ||
      null,
    refetch: () => {
      void sourcesQuery.refetch();
      void timelineQuery.refetch();
    },
    syncNow: () => syncMutation.mutate(),
    isSyncing: syncMutation.isPending,
    selectedEvent,
    openEvent,
    closeEvent,
    sidebarOpen,
    setSidebarOpen,
    extendAgenda,
    timeZone,
    sourceById,
    calendarLabelFor,
    colorFor,
    createEventMutation,
    updateEventMutation,
    deleteEventMutation,
    createReminderMutation,
    updateReminderMutation,
    deleteReminderMutation,
  };
}

export type CalendarController = ReturnType<typeof useCalendarController>;
