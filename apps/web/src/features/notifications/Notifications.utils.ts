import {
  format,
  formatDistanceToNow,
  isToday,
  isYesterday,
  parseISO,
  startOfWeek,
} from "date-fns";

import type {
  DonnaNotification,
  NotificationFilter,
  NotificationGroup,
  NotificationGroupKey,
  NotificationPayload,
  NotificationStatus,
  StatusTimelineStep,
} from "./Notifications.types";

export const FILTER_STORAGE_KEY = "donna.notifications.filter";
export const PAGE_SIZE = 50;

export const FILTER_OPTIONS: { id: NotificationFilter; label: string }[] = [
  { id: "all", label: "All" },
  { id: "unread", label: "Unread" },
  { id: "pending", label: "Pending" },
  { id: "sent", label: "Sent" },
  { id: "failed", label: "Failed" },
  { id: "dismissed", label: "Dismissed" },
];

const STATUS_COLORS: Record<string, string> = {
  PENDING: "#eab308",
  SENT: "#22c55e",
  READ: "#3b82f6",
  DISMISSED: "#9ca3af",
  FAILED: "#ef4444",
};

export function statusColor(status: string): string {
  return STATUS_COLORS[status.toUpperCase()] ?? "#9ca3af";
}

export function statusLabel(status: string): string {
  const upper = status.toUpperCase();
  return upper.charAt(0) + upper.slice(1).toLowerCase();
}

export function parsePayload(
  raw: DonnaNotification["payload"],
): NotificationPayload {
  if (!raw || typeof raw !== "object") return {};
  return raw as NotificationPayload;
}

export function notificationSource(n: DonnaNotification): string {
  const payload = parsePayload(n.payload);
  if (payload.source?.trim()) return payload.source.trim();
  if (n.notification_type === "REMINDER") return "Donna Reminder";
  if (n.notification_type === "EVENT") return "Donna Event";
  return "Donna";
}

export function notificationSortAt(n: DonnaNotification): Date {
  const raw = n.scheduled_for || n.created_at;
  try {
    return parseISO(raw);
  } catch {
    return new Date(0);
  }
}

export function formatRelativeTime(iso: string | null | undefined): string {
  if (!iso) return "";
  try {
    return formatDistanceToNow(parseISO(iso), { addSuffix: true });
  } catch {
    return "";
  }
}

export function formatClock(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return format(parseISO(iso), "h:mm a");
  } catch {
    return "—";
  }
}

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    const d = parseISO(iso);
    if (isToday(d)) return `Today ${format(d, "h:mm a")}`;
    if (isYesterday(d)) return `Yesterday ${format(d, "h:mm a")}`;
    return format(d, "MMM d · h:mm a");
  } catch {
    return "—";
  }
}

export function matchesSearch(
  n: DonnaNotification,
  query: string,
): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  return (
    n.title.toLowerCase().includes(q) || n.body.toLowerCase().includes(q)
  );
}

export function filterMatchesStatus(
  filter: NotificationFilter,
  status: string,
): boolean {
  const upper = status.toUpperCase();
  switch (filter) {
    case "all":
      return true;
    case "unread":
      return upper === "SENT";
    case "pending":
      return upper === "PENDING";
    case "sent":
      return upper === "SENT";
    case "failed":
      return upper === "FAILED";
    case "dismissed":
      return upper === "DISMISSED";
    default:
      return true;
  }
}

export function unreadCount(items: DonnaNotification[]): number {
  return items.filter((n) => n.status.toUpperCase() === "SENT").length;
}

export function filterNotifications(
  items: DonnaNotification[],
  filter: NotificationFilter,
  search: string,
): DonnaNotification[] {
  return items
    .filter((n) => filterMatchesStatus(filter, n.status))
    .filter((n) => matchesSearch(n, search))
    .sort(
      (a, b) => notificationSortAt(b).getTime() - notificationSortAt(a).getTime(),
    );
}

function groupKeyFor(date: Date, now: Date): NotificationGroupKey {
  if (isToday(date)) return "today";
  if (isYesterday(date)) return "yesterday";
  const weekStart = startOfWeek(now, { weekStartsOn: 1 });
  if (date >= weekStart) return "earlier_this_week";
  return "older";
}

const GROUP_LABELS: Record<NotificationGroupKey, string> = {
  today: "Today",
  yesterday: "Yesterday",
  earlier_this_week: "Earlier This Week",
  older: "Older",
};

const GROUP_ORDER: NotificationGroupKey[] = [
  "today",
  "yesterday",
  "earlier_this_week",
  "older",
];

export function groupNotifications(
  items: DonnaNotification[],
  now = new Date(),
): NotificationGroup[] {
  const buckets = new Map<NotificationGroupKey, DonnaNotification[]>();
  for (const key of GROUP_ORDER) {
    buckets.set(key, []);
  }
  for (const item of items) {
    const key = groupKeyFor(notificationSortAt(item), now);
    buckets.get(key)!.push(item);
  }
  return GROUP_ORDER.map((key) => ({
    key,
    label: GROUP_LABELS[key],
    items: buckets.get(key) ?? [],
  })).filter((g) => g.items.length > 0);
}

export function paginateItems<T>(items: T[], visibleCount: number): T[] {
  return items.slice(0, Math.max(0, visibleCount));
}

export function failureReason(n: DonnaNotification): string | null {
  const payload = parsePayload(n.payload);
  if (payload.failureReason?.trim()) return payload.failureReason.trim();
  if (payload.error?.trim()) return payload.error.trim();
  const channels = n.channel_delivery_status;
  if (channels && typeof channels === "object") {
    for (const [channel, status] of Object.entries(channels)) {
      if (String(status).toUpperCase() === "FAILED") {
        return `${channel} failed`;
      }
    }
  }
  return n.status.toUpperCase() === "FAILED" ? "Delivery failed" : null;
}

export function buildStatusTimeline(
  n: DonnaNotification,
): StatusTimelineStep[] {
  const status = n.status.toUpperCase() as NotificationStatus;
  const steps: StatusTimelineStep[] = [
    {
      id: "created",
      label: "Notification Created",
      at: n.created_at,
      done: true,
    },
    {
      id: "queued",
      label: "Queued",
      at: n.created_at,
      done: true,
    },
    {
      id: "sent",
      label: status === "FAILED" ? "Delivery Failed" : "Delivered",
      at: n.sent_at ?? null,
      done: Boolean(n.sent_at) || status === "FAILED" || status === "SENT" || status === "READ" || status === "DISMISSED",
    },
    {
      id: "read",
      label: "Marked Read",
      at: n.read_at ?? null,
      done: Boolean(n.read_at) || status === "READ" || status === "DISMISSED",
    },
    {
      id: "dismissed",
      label: "Dismissed",
      at: n.dismissed_at ?? null,
      done: Boolean(n.dismissed_at) || status === "DISMISSED",
    },
  ];

  if (status === "PENDING") {
    return steps.map((s) =>
      s.id === "sent" || s.id === "read" || s.id === "dismissed"
        ? { ...s, done: false, at: null }
        : s,
    );
  }
  if (status === "FAILED") {
    return steps.filter(
      (s) => s.id === "created" || s.id === "queued" || s.id === "sent",
    );
  }
  if (status === "SENT") {
    return steps.map((s) =>
      s.id === "read" || s.id === "dismissed"
        ? { ...s, done: false }
        : s,
    );
  }
  if (status === "READ") {
    return steps.map((s) =>
      s.id === "dismissed" ? { ...s, done: false, at: null } : s,
    );
  }
  return steps;
}

export function loadStoredFilter(): NotificationFilter {
  if (typeof window === "undefined") return "all";
  try {
    const raw = window.localStorage.getItem(FILTER_STORAGE_KEY);
    if (
      raw === "all" ||
      raw === "unread" ||
      raw === "pending" ||
      raw === "sent" ||
      raw === "failed" ||
      raw === "dismissed"
    ) {
      return raw;
    }
  } catch {
    // ignore
  }
  return "all";
}

export function saveStoredFilter(filter: NotificationFilter): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(FILTER_STORAGE_KEY, filter);
  } catch {
    // ignore
  }
}

export function calendarHrefForOccurrence(occurrenceId: string): string {
  return `/dashboard/calendar?event=${encodeURIComponent(occurrenceId)}`;
}

export function isDevBuild(): boolean {
  return process.env.NODE_ENV === "development";
}

export function cardIconName(
  n: DonnaNotification,
): "bell" | "clock" | "calendar" {
  if (n.notification_type === "REMINDER") return "clock";
  const source = notificationSource(n).toLowerCase();
  if (source.includes("google") || source.includes("microsoft")) {
    return "calendar";
  }
  return "bell";
}

export function availableActions(
  status: string,
): Array<"read" | "dismiss"> {
  const upper = status.toUpperCase();
  if (upper === "SENT") return ["read", "dismiss"];
  if (upper === "READ" || upper === "FAILED") return ["dismiss"];
  return [];
}
