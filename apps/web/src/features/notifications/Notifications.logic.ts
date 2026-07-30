"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/features/auth";

import {
  dismissNotification,
  listNotifications,
  markNotificationRead,
} from "./Notifications.api";
import type {
  DonnaNotification,
  NotificationFilter,
} from "./Notifications.types";
import {
  FILTER_STORAGE_KEY,
  PAGE_SIZE,
  filterNotifications,
  groupNotifications,
  loadStoredFilter,
  paginateItems,
  saveStoredFilter,
  unreadCount,
} from "./Notifications.utils";

export const notificationQueryKeys = {
  all: ["notifications"] as const,
  list: () => [...notificationQueryKeys.all, "list"] as const,
};

export function useNotificationsCenter() {
  const { status: authStatus } = useAuth();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [filter, setFilterState] = useState<NotificationFilter>("all");
  const [search, setSearch] = useState("");
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setFilterState(loadStoredFilter());
    setHydrated(true);
  }, []);

  const listQuery = useQuery({
    queryKey: notificationQueryKeys.list(),
    queryFn: ({ signal }) => listNotifications({ signal }),
    refetchInterval: 30_000,
    refetchOnWindowFocus: true,
    enabled: hydrated && authStatus === "authenticated",
  });

  const items = listQuery.data ?? [];

  const setFilter = useCallback((next: NotificationFilter) => {
    setFilterState(next);
    saveStoredFilter(next);
    setVisibleCount(PAGE_SIZE);
  }, []);

  const filtered = useMemo(
    () => filterNotifications(items, filter, search),
    [items, filter, search],
  );

  const visible = useMemo(
    () => paginateItems(filtered, visibleCount),
    [filtered, visibleCount],
  );

  const groups = useMemo(() => groupNotifications(visible), [visible]);

  const selected = useMemo(
    () => items.find((n) => n.id === selectedId) ?? null,
    [items, selectedId],
  );

  const badgeCount = useMemo(() => unreadCount(items), [items]);

  const invalidate = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: notificationQueryKeys.all,
    });
  }, [queryClient]);

  const readMutation = useMutation({
    mutationFn: (id: string) => markNotificationRead(id),
    onSuccess: async (updated) => {
      queryClient.setQueryData<DonnaNotification[]>(
        notificationQueryKeys.list(),
        (current) =>
          (current ?? []).map((n) => (n.id === updated.id ? updated : n)),
      );
      await invalidate();
    },
  });

  const dismissMutation = useMutation({
    mutationFn: (id: string) => dismissNotification(id),
    onSuccess: async (updated) => {
      queryClient.setQueryData<DonnaNotification[]>(
        notificationQueryKeys.list(),
        (current) =>
          (current ?? []).map((n) => (n.id === updated.id ? updated : n)),
      );
      await invalidate();
    },
  });

  const openDrawer = useCallback(() => {
    setOpen(true);
    setSelectedId(null);
  }, []);

  const closeDrawer = useCallback(() => {
    setOpen(false);
    setSelectedId(null);
    setSearch("");
  }, []);

  const openDetails = useCallback((id: string) => {
    setSelectedId(id);
  }, []);

  const backToList = useCallback(() => {
    setSelectedId(null);
  }, []);

  const loadMore = useCallback(() => {
    setVisibleCount((n) => n + PAGE_SIZE);
  }, []);

  const markRead = useCallback(
    async (id: string) => {
      await readMutation.mutateAsync(id);
    },
    [readMutation],
  );

  const dismiss = useCallback(
    async (id: string) => {
      await dismissMutation.mutateAsync(id);
    },
    [dismissMutation],
  );

  const hasMore = visible.length < filtered.length;

  return {
    open,
    openDrawer,
    closeDrawer,
    filter,
    setFilter,
    search,
    setSearch,
    items,
    filtered,
    groups,
    selected,
    openDetails,
    backToList,
    badgeCount,
    hasMore,
    loadMore,
    markRead,
    dismiss,
    isLoading: listQuery.isLoading,
    isError: listQuery.isError,
    isFetching: listQuery.isFetching,
    isSaving: readMutation.isPending || dismissMutation.isPending,
    storageKey: FILTER_STORAGE_KEY,
  };
}

export type NotificationsCenterController = ReturnType<
  typeof useNotificationsCenter
>;
