import { TIMELINE_COLORS } from "@/features/timeline/Timeline.colors";
import type { TimelineItem } from "@/features/timeline/Timeline.types";

import type { CalendarEvent } from "./Calendar.types";

/** Stable client filter keys — must not depend on API virtual source UUIDs. */
export const DONNA_EVENT_SOURCE_ID = "donna-events";
export const DONNA_REMINDER_SOURCE_ID = "donna-reminders";

export function timelineItemToCalendarEvent(
  item: TimelineItem,
): CalendarEvent {
  const meta = item.metadata ?? {};
  const metaSource =
    typeof meta.calendar_source_id === "string"
      ? meta.calendar_source_id
      : null;
  const publicId =
    typeof meta.public_id === "string" ? meta.public_id : item.id;

  let calendarSourceId = DONNA_EVENT_SOURCE_ID;
  if (item.source === "DONNA" && item.type === "REMINDER") {
    calendarSourceId = DONNA_REMINDER_SOURCE_ID;
  } else if (item.source === "DONNA") {
    calendarSourceId = DONNA_EVENT_SOURCE_ID;
  } else if (metaSource) {
    calendarSourceId = metaSource;
  }

  const cancelled = item.status === "CANCELLED";
  const parentId = item.parent_id ?? null;
  let mutationId: string | undefined;
  if (!item.read_only && item.source === "DONNA") {
    if (parentId) {
      mutationId = parentId;
    } else if (/^[0-9a-f-]{36}_/i.test(item.id)) {
      mutationId = item.id.slice(0, 36);
    } else if (/^[0-9a-f-]{36}$/i.test(item.id)) {
      mutationId = item.id;
    }
  }

  return {
    id: item.occurrence_id || item.id,
    public_id: publicId,
    calendar_source_id: calendarSourceId,
    title: item.title,
    description: item.description ?? undefined,
    location:
      typeof meta.location === "string" ? meta.location : undefined,
    start_time: item.start_at,
    end_time: item.end_at,
    timezone: item.timezone,
    all_day: item.all_day,
    status: cancelled ? "cancelled" : "confirmed",
    attendees: [],
    recurring_event_id: parentId ?? undefined,
    provider_recurring_event_id: item.is_recurring ? "rrule" : undefined,
    origin: item.source === "DONNA" ? "donna" : "provider_sync",
    created_at: item.start_at,
    updated_at: item.start_at,
    read_only: item.read_only,
    timeline_source: item.source,
    timeline_type: item.type === "REMINDER" ? "REMINDER" : "EVENT",
    occurrence_id: item.occurrence_id,
    mutation_id: mutationId,
    recurrence_rule: item.recurrence_rule ?? undefined,
    accent_color:
      item.source === "DONNA" && item.type === "REMINDER"
        ? TIMELINE_COLORS.donnaReminder
        : item.source === "DONNA"
          ? TIMELINE_COLORS.donnaEvent
          : undefined,
  };
}

export function donnaEventAccountGroup(visible: boolean) {
  return {
    accountId: "donna-events",
    label: "Donna Events",
    email: null as string | null,
    color: TIMELINE_COLORS.donnaEvent,
    sourceIds: [DONNA_EVENT_SOURCE_ID],
    visibleCount: visible ? 1 : 0,
  };
}

export function donnaReminderAccountGroup(visible: boolean) {
  return {
    accountId: "donna-reminders",
    label: "Donna Reminders",
    email: null as string | null,
    color: TIMELINE_COLORS.donnaReminder,
    sourceIds: [DONNA_REMINDER_SOURCE_ID],
    visibleCount: visible ? 1 : 0,
  };
}
