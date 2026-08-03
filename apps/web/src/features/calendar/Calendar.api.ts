import { apiRequest } from "@/lib/api/client";

import type {
  CalendarEventsResponse,
  CalendarSourcesResponse,
  CalendarSyncResult,
} from "./Calendar.types";

export type ListEventsParams = {
  from: string;
  to: string;
  signal?: AbortSignal;
};

/** Manual sync — orchestrates sources + events into Donna DB. */
export function syncCalendarSources(signal?: AbortSignal): Promise<CalendarSyncResult> {
  return apiRequest<CalendarSyncResult>("/api/v1/calendar/sync", {
    method: "POST",
    signal,
  });
}

/**
 * Startup / on-demand freshness: full pipeline when last success is older than 15 minutes.
 */
export function ensureCalendarSourcesFresh(
  signal?: AbortSignal,
): Promise<CalendarSyncResult> {
  return apiRequest<CalendarSyncResult>("/api/v1/calendar/sync/ensure", {
    method: "POST",
    signal,
  });
}

/** List unified calendar events from Donna DB. Never hits Google. */
export function listCalendarEvents(
  params: ListEventsParams,
): Promise<CalendarEventsResponse> {
  const query = new URLSearchParams({
    from: params.from,
    to: params.to,
  });
  return apiRequest<CalendarEventsResponse>(
    `/api/v1/calendar/events?${query.toString()}`,
    { signal: params.signal },
  );
}

/** List calendar sources + account sync observability from Donna DB. */
export function listCalendarSources(
  signal?: AbortSignal,
): Promise<CalendarSourcesResponse> {
  return apiRequest<CalendarSourcesResponse>("/api/v1/calendar/sources", {
    signal,
  });
}
