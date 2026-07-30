/** Centralized Timeline visual tokens — Phase 3.1 */

export const TIMELINE_COLORS = {
  google: "#3B82F6",
  ics: "#22C55E",
  microsoft: "#0EA5E9",
  donnaEvent: "#8B5CF6",
  donnaReminder: "#F97316",
  cancelled: "#6B7280",
  completed: "#10B981",
} as const;

export type TimelineSourceKey =
  | "GOOGLE"
  | "MICROSOFT_ICS"
  | "DONNA_EVENT"
  | "DONNA_REMINDER";

export type TimelineFilterKey =
  | "google"
  | "ics"
  | "donna_events"
  | "donna_reminders"
  | "completed"
  | "cancelled";

export const DEFAULT_TIMELINE_FILTERS: Record<TimelineFilterKey, boolean> = {
  google: true,
  ics: true,
  donna_events: true,
  donna_reminders: true,
  completed: false,
  cancelled: false,
};

export const TIMELINE_FILTER_LABELS: Record<TimelineFilterKey, string> = {
  google: "Google",
  ics: "Microsoft / ICS",
  donna_events: "Donna Events",
  donna_reminders: "Donna Reminders",
  completed: "Completed",
  cancelled: "Cancelled",
};
