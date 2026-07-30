export type CalendarView = "day" | "week" | "month" | "agenda";

export type CalendarOrganizer = {
  email?: string;
  displayName?: string;
  self?: boolean;
};

export type CalendarAttendee = {
  email?: string;
  displayName?: string;
  responseStatus?: string;
  self?: boolean;
  organizer?: boolean;
};

export type CalendarEvent = {
  id: string;
  public_id: string;
  calendar_source_id: string;
  title: string;
  description?: string;
  location?: string;
  start_time: string;
  end_time: string;
  timezone?: string;
  all_day: boolean;
  status: string;
  visibility?: string;
  organizer?: unknown;
  attendees: unknown;
  recurring_event_id?: string;
  provider_recurring_event_id?: string;
  provider_event_id?: string;
  provider_updated_at?: string;
  origin: string;
  created_at: string;
  updated_at: string;
  /** Timeline-backed fields (Donna CRUD). */
  read_only?: boolean;
  timeline_source?: string;
  timeline_type?: "EVENT" | "REMINDER";
  occurrence_id?: string;
  mutation_id?: string;
  recurrence_rule?: string;
  accent_color?: string;
};

export type CalendarEventsResponse = {
  events: CalendarEvent[];
  from: string;
  to: string;
};

export type CalendarSource = {
  id: string;
  public_id: string;
  connected_account_id: string;
  provider_calendar_id: string;
  name: string;
  color?: string;
  is_primary_on_provider: boolean;
  is_writable: boolean;
  access_role?: string;
  sync_enabled: boolean;
  sync_token?: string;
  last_synced_at?: string;
  timezone?: string;
  provider_metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type CalendarAccountSync = {
  sync_status: string;
  last_successful_sync?: string;
  last_failed_sync?: string;
  sync_duration_ms?: number;
  records_created: number;
  records_updated: number;
  records_deleted: number;
};

export type CalendarConnectedAccount = {
  id: string;
  provider: string;
  display_name?: string;
  email?: string | null;
  avatar_url?: string | null;
};

export type CalendarSourcesResponse = {
  sources: CalendarSource[];
  account?: CalendarConnectedAccount;
  accounts?: CalendarConnectedAccount[];
  sync?: CalendarAccountSync;
};

/** One connected Google account shown in the calendars list. */
export type CalendarAccountGroup = {
  accountId: string;
  label: string;
  /** Provider account email when known — used to match the Donna login inbox. */
  email?: string | null;
  color: string;
  sourceIds: string[];
  visibleCount: number;
};
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
  sources: CalendarSource[];
  incremental: boolean;
  skipped: boolean;
  sync_status: string;
};

/** Positioned timed event for day/week grids. */
export type LaidOutEvent = {
  event: CalendarEvent;
  top: number;
  height: number;
  left: number;
  width: number;
  column: number;
  columnCount: number;
};

export type AgendaGroup = {
  key: string;
  date: Date;
  label: string;
  events: CalendarEvent[];
};
