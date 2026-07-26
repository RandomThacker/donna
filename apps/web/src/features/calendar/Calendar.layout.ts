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
  const start = new Date(day);
  start.setHours(0, 0, 0, 0);
  const end = new Date(day);
  end.setHours(23, 59, 59, 999);
  return events.filter((event) => {
    const s = new Date(event.start_time);
    const e = new Date(event.end_time);
    return s <= end && e >= start;
  });
}

export function allDayEventsForDay(
  events: CalendarEvent[],
  day: Date,
): CalendarEvent[] {
  return eventsOverlappingDay(events, day).filter((e) => e.all_day);
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
