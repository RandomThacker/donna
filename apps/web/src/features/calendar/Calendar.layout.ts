import type { CalendarEvent, LaidOutEvent } from "./Calendar.types";

export const HOUR_HEIGHT = 64;
export const DAY_START_HOUR = 0;
export const DAY_END_HOUR = 24;
export const HOURS = Array.from({ length: DAY_END_HOUR - DAY_START_HOUR }, (_, i) => i);

const FALLBACK_PALETTE = [
  "#c9a87c",
  "#6b9ac4",
  "#7eb89a",
  "#c47a8a",
  "#9b8fd9",
  "#d4a574",
  "#5ba8a0",
  "#b08d57",
] as const;

function parseHexColor(color: string): { r: number; g: number; b: number } | null {
  const raw = color.trim().replace(/^#/, "");
  if (/^[0-9a-fA-F]{3}$/.test(raw)) {
    return {
      r: parseInt(raw[0]! + raw[0]!, 16),
      g: parseInt(raw[1]! + raw[1]!, 16),
      b: parseInt(raw[2]! + raw[2]!, 16),
    };
  }
  if (/^[0-9a-fA-F]{6}$/.test(raw) || /^[0-9a-fA-F]{8}$/.test(raw)) {
    return {
      r: parseInt(raw.slice(0, 2), 16),
      g: parseInt(raw.slice(2, 4), 16),
      b: parseInt(raw.slice(4, 6), 16),
    };
  }
  return null;
}

/** Dark ink on light calendar colors, light ink on dark ones. */
export function contrastTextFor(background: string): string {
  const rgb = parseHexColor(background);
  if (!rgb) {
    return "#14110e";
  }
  const luminance = (0.2126 * rgb.r + 0.7152 * rgb.g + 0.0722 * rgb.b) / 255;
  return luminance > 0.62 ? "#14110e" : "#f3f1ec";
}

export function colorForSource(
  sourceId: string,
  color: string | undefined,
  index: number,
): string {
  if (color && /^#?[0-9a-fA-F]{3,8}$/.test(color)) {
    return color.startsWith("#") ? color : `#${color}`;
  }
  return FALLBACK_PALETTE[index % FALLBACK_PALETTE.length] ?? FALLBACK_PALETTE[0];
}

export function minutesFromDayStart(date: Date): number {
  return date.getHours() * 60 + date.getMinutes();
}

export function layoutTimedEvents(
  events: CalendarEvent[],
  day: Date,
): LaidOutEvent[] {
  const dayStart = new Date(day);
  dayStart.setHours(0, 0, 0, 0);
  const dayEnd = new Date(day);
  dayEnd.setHours(23, 59, 59, 999);

  const timed = events
    .filter((e) => !e.all_day)
    .map((event) => {
      const start = new Date(event.start_time);
      const end = new Date(event.end_time);
      const clampedStart = start < dayStart ? dayStart : start;
      const clampedEnd = end > dayEnd ? dayEnd : end;
      const startMin = Math.max(0, minutesFromDayStart(clampedStart));
      const endMin = Math.max(
        startMin + 15,
        Math.min(24 * 60, minutesFromDayStart(clampedEnd) || 24 * 60),
      );
      return { event, startMin, endMin };
    })
    .filter(({ startMin, endMin }) => endMin > startMin)
    .sort((a, b) => a.startMin - b.startMin || b.endMin - a.endMin);

  type Node = {
    event: CalendarEvent;
    startMin: number;
    endMin: number;
    column: number;
  };

  const nodes: Node[] = [];
  const active: Node[] = [];

  for (const item of timed) {
    for (let i = active.length - 1; i >= 0; i -= 1) {
      if (active[i]!.endMin <= item.startMin) {
        active.splice(i, 1);
      }
    }
    const used = new Set(active.map((n) => n.column));
    let column = 0;
    while (used.has(column)) {
      column += 1;
    }
    const node: Node = { ...item, column };
    nodes.push(node);
    active.push(node);
  }

  // Cluster by overlap to compute column counts.
  const clusters: Node[][] = [];
  let cluster: Node[] = [];
  let clusterEnd = -1;
  for (const node of nodes) {
    if (cluster.length === 0 || node.startMin < clusterEnd) {
      cluster.push(node);
      clusterEnd = Math.max(clusterEnd, node.endMin);
    } else {
      clusters.push(cluster);
      cluster = [node];
      clusterEnd = node.endMin;
    }
  }
  if (cluster.length) {
    clusters.push(cluster);
  }

  const laid: LaidOutEvent[] = [];
  for (const group of clusters) {
    const columnCount = Math.max(1, ...group.map((n) => n.column + 1));
    for (const node of group) {
      const top = (node.startMin / 60) * HOUR_HEIGHT;
      const height = Math.max(
        20,
        ((node.endMin - node.startMin) / 60) * HOUR_HEIGHT - 2,
      );
      const width = 100 / columnCount;
      laid.push({
        event: node.event,
        top,
        height,
        left: node.column * width,
        width,
        column: node.column,
        columnCount,
      });
    }
  }
  return laid;
}

export function eventsOverlappingDay(
  events: CalendarEvent[],
  day: Date,
): CalendarEvent[] {
  const dayStart = new Date(day);
  dayStart.setHours(0, 0, 0, 0);
  const dayEnd = new Date(day);
  dayEnd.setHours(23, 59, 59, 999);
  const dayKey = localDateKey(dayStart);

  return events.filter((event) => {
    if (event.all_day) {
      // Google/Microsoft all-day ends are exclusive (event on the 25th → end = 26th).
      // Compare civil dates from the stored UTC midnight values.
      const startKey = utcDateKey(new Date(event.start_time));
      const endKey = utcDateKey(new Date(event.end_time));
      return startKey <= dayKey && endKey > dayKey;
    }
    const start = new Date(event.start_time);
    const end = new Date(event.end_time);
    return start <= dayEnd && end > dayStart;
  });
}

function localDateKey(day: Date): string {
  const y = day.getFullYear();
  const m = String(day.getMonth() + 1).padStart(2, "0");
  const d = String(day.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function utcDateKey(day: Date): string {
  const y = day.getUTCFullYear();
  const m = String(day.getUTCMonth() + 1).padStart(2, "0");
  const d = String(day.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function allDayEventsForDay(
  events: CalendarEvent[],
  day: Date,
): CalendarEvent[] {
  return eventsOverlappingDay(events, day).filter((e) => e.all_day);
}

/** Google subscribed holiday feeds (often duplicated across accounts / locales). */
export function isHolidaySource(source: {
  provider_calendar_id?: string;
  name?: string;
}): boolean {
  const providerId = source.provider_calendar_id?.trim().toLowerCase() ?? "";
  if (providerId.includes("holiday@group.v.calendar.google.com")) {
    return true;
  }
  const name = source.name?.trim().toLowerCase() ?? "";
  return name.startsWith("holidays in ");
}

/**
 * Collapse identical holiday chips when the same feed is connected more than once
 * (e.g. en + en-in India holidays, or the same holidays under two Google accounts).
 */
export function dedupeHolidayEvents<
  T extends {
    id: string;
    title: string;
    start_time: string;
    end_time: string;
    all_day: boolean;
    calendar_source_id: string;
  },
>(
  events: T[],
  sourcesById: Map<string, { provider_calendar_id?: string; name?: string }>,
): T[] {
  const seenHolidayKeys = new Set<string>();
  const out: T[] = [];

  for (const event of events) {
    const source = sourcesById.get(event.calendar_source_id);
    if (!source || !isHolidaySource(source)) {
      out.push(event);
      continue;
    }

    const startKey = event.all_day
      ? utcDateKey(new Date(event.start_time))
      : event.start_time;
    const endKey = event.all_day
      ? utcDateKey(new Date(event.end_time))
      : event.end_time;
    const key = `${event.title.trim().toLowerCase()}|${startKey}|${endKey}|${event.all_day ? "1" : "0"}`;
    if (seenHolidayKeys.has(key)) {
      continue;
    }
    seenHolidayKeys.add(key);
    out.push(event);
  }

  return out;
}

export function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

export function isRecurring(event: CalendarEvent): boolean {
  return Boolean(event.recurring_event_id || event.provider_recurring_event_id);
}
