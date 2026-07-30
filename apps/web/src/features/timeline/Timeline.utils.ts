import { TIMELINE_COLORS, type TimelineFilterKey } from "./Timeline.colors";
import type { TimelineItem } from "./Timeline.types";

export function providerFromItem(item: TimelineItem): string {
  const meta = item.metadata ?? {};
  const provider = meta.provider;
  return typeof provider === "string" ? provider.toLowerCase() : "";
}

export function colorForTimelineItem(item: TimelineItem): string {
  if (item.status === "CANCELLED") return TIMELINE_COLORS.cancelled;
  if (item.status === "COMPLETED") return TIMELINE_COLORS.completed;
  if (item.color && /^#?[0-9a-fA-F]{3,8}$/.test(item.color)) {
    return item.color.startsWith("#") ? item.color : `#${item.color}`;
  }
  if (item.source === "DONNA" && item.type === "REMINDER") {
    return TIMELINE_COLORS.donnaReminder;
  }
  if (item.source === "DONNA") {
    return TIMELINE_COLORS.donnaEvent;
  }
  if (item.source === "GOOGLE" || providerFromItem(item) === "google") {
    return TIMELINE_COLORS.google;
  }
  if (providerFromItem(item) === "ics") {
    return TIMELINE_COLORS.ics;
  }
  if (
    item.source === "MICROSOFT_ICS" ||
    providerFromItem(item) === "microsoft"
  ) {
    return TIMELINE_COLORS.microsoft;
  }
  return TIMELINE_COLORS.donnaEvent;
}

export function entityIdForMutation(item: TimelineItem): string | null {
  if (item.read_only || item.source !== "DONNA") return null;
  if (item.parent_id) return item.parent_id;
  // Occurrence ids look like `{uuid}_{stamp}` — strip stamp when present.
  const id = item.id;
  if (/^[0-9a-f-]{36}_/i.test(id)) {
    return id.slice(0, 36);
  }
  if (/^[0-9a-f-]{36}$/i.test(id)) return id;
  return null;
}

export function matchesTimelineFilters(
  item: TimelineItem,
  filters: Record<TimelineFilterKey, boolean>,
  search: string,
): boolean {
  const status = item.status;
  if (status === "CANCELLED" && !filters.cancelled) return false;
  if (status === "COMPLETED" && !filters.completed) return false;
  if (status !== "CANCELLED" && status !== "COMPLETED") {
    // active path
  }

  if (item.source === "GOOGLE") {
    if (!filters.google) return false;
  } else if (item.source === "MICROSOFT_ICS") {
    if (!filters.ics) return false;
  } else if (item.source === "DONNA" && item.type === "REMINDER") {
    if (!filters.donna_reminders) return false;
  } else if (item.source === "DONNA") {
    if (!filters.donna_events) return false;
  }

  const q = search.trim().toLowerCase();
  if (!q) return true;
  const hay = `${item.title}\n${item.description ?? ""}`.toLowerCase();
  return hay.includes(q);
}

export function isDonnaEditable(item: TimelineItem): boolean {
  return !item.read_only && item.source === "DONNA";
}

export function sourceLabel(item: TimelineItem): string {
  if (item.source === "DONNA" && item.type === "REMINDER") return "Donna Reminder";
  if (item.source === "DONNA") return "Donna Event";
  if (item.source === "GOOGLE") return "Google";
  const p = providerFromItem(item);
  if (p === "ics") return "ICS";
  if (p === "microsoft") return "Microsoft";
  return "Microsoft / ICS";
}

export function notificationPolicyLabel(item: TimelineItem): string {
  if (item.source === "DONNA" && item.type === "EVENT") {
    return "Remind 15 minutes before";
  }
  if (item.source === "DONNA" && item.type === "REMINDER") {
    return "At trigger time";
  }
  return "Provider-managed";
}
