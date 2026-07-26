import { apiRequest } from "@/lib/api/client";

export type CalendarSyncFailure = {
  calendar_source_id?: string;
  provider_calendar_id?: string;
  name?: string;
  stage: string;
  error: string;
};

export type CalendarSyncResult = {
  run_id?: string;
  trigger: string;
  status: string;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  calendars_processed: number;
  sources_created: number;
  sources_updated: number;
  sources_deleted: number;
  events_created: number;
  events_updated: number;
  events_deleted: number;
  failures: CalendarSyncFailure[];
  sources: unknown[];
  incremental: boolean;
  skipped: boolean;
  sync_status: string;
};

/** Manual sync — orchestrates sources + events into Donna DB. */
export function syncCalendarSources(signal?: AbortSignal): Promise<CalendarSyncResult> {
  return apiRequest<CalendarSyncResult>("/api/v1/calendar/sync", {
    method: "POST",
    signal,
  });
}

/**
 * Startup / on-demand freshness: full pipeline when last success is older than 2 minutes.
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
