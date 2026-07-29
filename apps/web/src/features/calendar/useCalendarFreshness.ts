"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";

import { syncCalendarSources } from "./Calendar.api";
import { calendarQueryKeys } from "./Calendar.utils";

/**
 * Pulls provider calendars into Donna DB, then refreshes local event queries.
 * Shared across dashboard widgets so multiple mounts only sync once per window.
 */
export function useCalendarFreshness() {
  const queryClient = useQueryClient();

  return useQuery({
    queryKey: calendarQueryKeys.freshness,
    queryFn: async ({ signal }) => {
      const result = await syncCalendarSources(signal);
      await queryClient.invalidateQueries({
        queryKey: calendarQueryKeys.all,
        predicate: (query) => query.queryKey[1] !== "freshness",
      });
      return result;
    },
    staleTime: 60_000,
    refetchOnMount: true,
    refetchOnWindowFocus: true,
    retry: 1,
  });
}
