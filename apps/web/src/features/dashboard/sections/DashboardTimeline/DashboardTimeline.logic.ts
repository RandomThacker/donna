import { parseISO } from "date-fns";

import { formatZonedTime } from "@/features/calendar/Calendar.timezone";
import type { TimelineItem } from "@/features/timeline/Timeline.types";
import { sourceLabel } from "@/features/timeline/Timeline.utils";

import type { DashboardTimelineItem } from "../../Dashboard.types";

function itemStart(item: TimelineItem): string {
  return item.occurrence_start || item.start_at;
}

function itemEnd(item: TimelineItem): string {
  return item.occurrence_end || item.end_at;
}

function timelineMeta(item: TimelineItem, timeZone: string): string {
  if (item.type === "REMINDER") {
    return "Reminder";
  }
  const label = sourceLabel(item);
  if (item.all_day) {
    return `${label} · All day`;
  }
  const start = formatZonedTime(parseISO(itemStart(item)), timeZone);
  const end = formatZonedTime(parseISO(itemEnd(item)), timeZone);
  if (start && end && start !== end) {
    return `${start} – ${end}`;
  }
  return label;
}

export function isDashboardTimelineVisible(item: TimelineItem): boolean {
  const status = String(item.status || "").toUpperCase();
  return status !== "CANCELLED";
}

export function timelineItemsToDashboardItems(
  items: TimelineItem[],
  timeZone: string,
): DashboardTimelineItem[] {
  const visible = items
    .filter(isDashboardTimelineVisible)
    .slice()
    .sort((a, b) => {
      if (a.all_day !== b.all_day) {
        return a.all_day ? -1 : 1;
      }
      return itemStart(a).localeCompare(itemStart(b));
    });

  return visible.map((item) => ({
    id: item.occurrence_id || item.id,
    time: item.all_day
      ? "All day"
      : formatZonedTime(parseISO(itemStart(item)), timeZone),
    title: item.title?.trim() || "(No title)",
    meta: timelineMeta(item, timeZone),
  }));
}
