"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";

import type { TimelineFilterKey } from "./Timeline.colors";
import {
  createDonnaEvent,
  createDonnaReminder,
  deleteDonnaEvent,
  deleteDonnaReminder,
  fetchTimeline,
  updateDonnaEvent,
  updateDonnaReminder,
} from "./Timeline.api";
import {
  allDayItemsForDay,
  itemsOverlappingDay,
  layoutTimedItems,
  navigateCursor,
  queryRangeForView,
  titleForView,
} from "./Timeline.layout";
import {
  loadTimelineFilters,
  loadTimelineView,
  saveTimelineFilters,
  saveTimelineView,
} from "./Timeline.persistence";
import type {
  CreateDonnaEventInput,
  CreateDonnaReminderInput,
  TimelineItem,
  TimelineView,
  UpdateDonnaEventInput,
  UpdateDonnaReminderInput,
} from "./Timeline.types";
import { entityIdForMutation, matchesTimelineFilters } from "./Timeline.utils";

export const timelineQueryKeys = {
  all: ["timeline"] as const,
  range: (from: string, to: string) =>
    ["timeline", "items", from, to] as const,
};

export type CreateIntent = "event" | "reminder" | null;

export function useTimelineController() {
  const queryClient = useQueryClient();
  const [view, setViewState] = useState<TimelineView>("week");
  const [cursor, setCursor] = useState(() => new Date());
  const [filters, setFiltersState] = useState(loadTimelineFilters);
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<TimelineItem | null>(null);
  const [createIntent, setCreateIntent] = useState<CreateIntent>(null);
  const [createDay, setCreateDay] = useState<Date | null>(null);
  const [editing, setEditing] = useState<TimelineItem | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setViewState(loadTimelineView("week"));
    setFiltersState(loadTimelineFilters());
    setHydrated(true);
  }, []);

  const setView = useCallback((next: TimelineView) => {
    setViewState(next);
    saveTimelineView(next);
  }, []);

  const setFilters = useCallback(
    (next: Record<TimelineFilterKey, boolean>) => {
      setFiltersState(next);
      saveTimelineFilters(next);
    },
    [],
  );

  const toggleFilter = useCallback((key: TimelineFilterKey) => {
    setFiltersState((prev) => {
      const next = { ...prev, [key]: !prev[key] };
      saveTimelineFilters(next);
      return next;
    });
  }, []);

  const range = useMemo(
    () => queryRangeForView(view, cursor),
    [view, cursor],
  );
  const fromIso = range.from.toISOString();
  const toIso = range.to.toISOString();

  const timelineQuery = useQuery({
    queryKey: timelineQueryKeys.range(fromIso, toIso),
    queryFn: ({ signal }) =>
      fetchTimeline({ from: fromIso, to: toIso, signal }),
    enabled: hydrated,
  });

  const items = useMemo(() => {
    const all = timelineQuery.data?.items ?? [];
    return all.filter((item) => matchesTimelineFilters(item, filters, search));
  }, [timelineQuery.data?.items, filters, search]);

  const invalidate = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: timelineQueryKeys.all });
  }, [queryClient]);

  const createEventMutation = useMutation({
    mutationFn: (body: CreateDonnaEventInput) => createDonnaEvent(body),
    onSuccess: () => void invalidate(),
  });

  const updateEventMutation = useMutation({
    mutationFn: ({
      id,
      body,
    }: {
      id: string;
      body: UpdateDonnaEventInput;
    }) => updateDonnaEvent(id, body),
    onSuccess: () => void invalidate(),
  });

  const deleteEventMutation = useMutation({
    mutationFn: (id: string) => deleteDonnaEvent(id),
    onSuccess: () => {
      setSelected(null);
      setEditing(null);
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
      setSelected(null);
      setEditing(null);
      void invalidate();
    },
  });

  const dayLayout = useCallback(
    (day: Date) => ({
      allDay: allDayItemsForDay(items, day),
      timed: layoutTimedItems(itemsOverlappingDay(items, day), day),
    }),
    [items],
  );

  const openCreate = useCallback((day: Date, intent: CreateIntent = null) => {
    setCreateDay(day);
    setCreateIntent(intent);
    setSelected(null);
  }, []);

  const openItem = useCallback((item: TimelineItem) => {
    setSelected(item);
    setCreateIntent(null);
    setEditing(null);
  }, []);

  const openOccurrence = useCallback(
    (occurrenceId: string) => {
      const all = timelineQuery.data?.items ?? [];
      const hit =
        all.find((i) => i.occurrence_id === occurrenceId) ??
        all.find((i) => i.id === occurrenceId);
      if (hit) openItem(hit);
    },
    [timelineQuery.data?.items, openItem],
  );

  const startEdit = useCallback((item: TimelineItem) => {
    setEditing(item);
    setSelected(null);
  }, []);

  const removeItem = useCallback(
    async (item: TimelineItem) => {
      const id = entityIdForMutation(item);
      if (!id) return;
      if (item.type === "REMINDER") {
        await deleteReminderMutation.mutateAsync(id);
      } else {
        await deleteEventMutation.mutateAsync(id);
      }
    },
    [deleteEventMutation, deleteReminderMutation],
  );

  return {
    view,
    setView,
    cursor,
    setCursor,
    filters,
    setFilters,
    toggleFilter,
    search,
    setSearch,
    items,
    selected,
    setSelected,
    openItem,
    openOccurrence,
    createIntent,
    setCreateIntent,
    createDay,
    setCreateDay,
    openCreate,
    editing,
    setEditing,
    startEdit,
    sidebarOpen,
    setSidebarOpen,
    title: titleForView(view, cursor),
    goToday: () => setCursor(new Date()),
    goPrev: () => setCursor((c) => navigateCursor(view, c, -1)),
    goNext: () => setCursor((c) => navigateCursor(view, c, 1)),
    dayLayout,
    isLoading: !hydrated || timelineQuery.isLoading,
    isFetching: timelineQuery.isFetching,
    isError: timelineQuery.isError,
    errorMessage: (timelineQuery.error as Error | null)?.message ?? null,
    refetch: () => void timelineQuery.refetch(),
    createEventMutation,
    updateEventMutation,
    deleteEventMutation,
    createReminderMutation,
    updateReminderMutation,
    deleteReminderMutation,
    removeItem,
    invalidate,
  };
}

export type TimelineController = ReturnType<typeof useTimelineController>;
