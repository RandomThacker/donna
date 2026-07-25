import { apiRequest } from "@/lib/api/client";

export type CalendarSyncResult = {
  sources: unknown[];
  created_count: number;
  updated_count: number;
  removed_count: number;
  synced_at: string;
  duration_ms: number;
  incremental: boolean;
  skipped: boolean;
  sync_status: string;
};

/** Manual sync — always hits Google and refreshes Donna DB. */
export function syncCalendarSources(signal?: AbortSignal): Promise<CalendarSyncResult> {
  return apiRequest<CalendarSyncResult>("/api/v1/calendar/sync", {
    method: "POST",
    signal,
  });
}

/**
 * Startup / on-demand freshness: incremental sync only when last success is older than 2 minutes.
 * AI workflows that need calendar data should call this first.
 */
export function ensureCalendarSourcesFresh(
  signal?: AbortSignal,
): Promise<CalendarSyncResult> {
  return apiRequest<CalendarSyncResult>("/api/v1/calendar/sync/ensure", {
    method: "POST",
    signal,
  });
}
