import {
  addDays,
  addMonths,
  addWeeks,
  endOfDay,
  endOfMonth,
  endOfWeek,
  format,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";

import type { LaidOutTimelineItem, TimelineItem, TimelineView } from "./Timeline.types";

export const HOUR_HEIGHT = 64;
export const HOURS = Array.from({ length: 24 }, (_, i) => i);

export function weekDays(cursor: Date): Date[] {
  const start = startOfWeek(cursor, { weekStartsOn: 0 });
  return Array.from({ length: 7 }, (_, i) => addDays(start, i));
}

export function monthCells(cursor: Date): Date[] {
  const start = startOfWeek(startOfMonth(cursor), { weekStartsOn: 0 });
  return Array.from({ length: 42 }, (_, i) => addDays(start, i));
}

export function titleForView(view: TimelineView, cursor: Date): string {
  if (view === "day") return format(cursor, "EEEE, MMM d yyyy");
  if (view === "week") {
    const days = weekDays(cursor);
    const a = days[0]!;
    const b = days[6]!;
    return `${format(a, "MMM d")} – ${format(b, "MMM d, yyyy")}`;
  }
  if (view === "agenda") return "Agenda";
  return format(cursor, "MMMM yyyy");
}

export function navigateCursor(
  view: TimelineView,
  cursor: Date,
  delta: number,
): Date {
  if (view === "day") return addDays(cursor, delta);
  if (view === "week") return addWeeks(cursor, delta);
  if (view === "agenda") return addDays(cursor, delta * 7);
  return addMonths(cursor, delta);
}

export function queryRangeForView(
  view: TimelineView,
  cursor: Date,
  agendaDays = 60,
): { from: Date; to: Date } {
  if (view === "day") {
    return { from: startOfDay(cursor), to: endOfDay(cursor) };
  }
  if (view === "week") {
    const start = startOfWeek(cursor, { weekStartsOn: 0 });
    return { from: startOfDay(start), to: endOfDay(addDays(start, 6)) };
  }
  if (view === "agenda") {
    const from = startOfDay(cursor);
    return { from, to: endOfDay(addDays(from, agendaDays)) };
  }
  const start = startOfWeek(startOfMonth(cursor), { weekStartsOn: 0 });
  const end = endOfWeek(endOfMonth(cursor), { weekStartsOn: 0 });
  return { from: startOfDay(start), to: endOfDay(end) };
}

function minutesFromDayStart(date: Date): number {
  return date.getHours() * 60 + date.getMinutes();
}

export function itemsOverlappingDay(
  items: TimelineItem[],
  day: Date,
): TimelineItem[] {
  const dayStart = startOfDay(day).getTime();
  const dayEnd = endOfDay(day).getTime();
  return items.filter((item) => {
    const start = new Date(item.start_at).getTime();
    const end = new Date(item.end_at).getTime();
    return start <= dayEnd && end >= dayStart;
  });
}

export function allDayItemsForDay(
  items: TimelineItem[],
  day: Date,
): TimelineItem[] {
  return itemsOverlappingDay(items, day).filter((i) => i.all_day);
}

export function layoutTimedItems(
  items: TimelineItem[],
  day: Date,
): LaidOutTimelineItem[] {
  const dayStart = startOfDay(day);
  const dayEnd = endOfDay(day);

  const timed = items
    .filter((e) => !e.all_day)
    .map((item) => {
      const start = new Date(item.start_at);
      const end = new Date(item.end_at);
      const clampedStart = start < dayStart ? dayStart : start;
      const clampedEnd = end > dayEnd ? dayEnd : end;
      const startMin = Math.max(0, minutesFromDayStart(clampedStart));
      const endMin = Math.max(
        startMin + 15,
        Math.min(24 * 60, minutesFromDayStart(clampedEnd) || 24 * 60),
      );
      return { item, startMin, endMin };
    })
    .filter(({ startMin, endMin }) => endMin > startMin)
    .sort((a, b) => a.startMin - b.startMin || b.endMin - a.endMin);

  type Node = {
    item: TimelineItem;
    startMin: number;
    endMin: number;
    column: number;
  };

  const nodes: Node[] = [];
  const active: Node[] = [];

  for (const entry of timed) {
    for (let i = active.length - 1; i >= 0; i -= 1) {
      if (active[i]!.endMin <= entry.startMin) active.splice(i, 1);
    }
    const used = new Set(active.map((n) => n.column));
    let column = 0;
    while (used.has(column)) column += 1;
    const node: Node = { ...entry, column };
    nodes.push(node);
    active.push(node);
  }

  const groups: Node[][] = [];
  let cluster: Node[] = [];
  let clusterEnd = -1;
  for (const node of nodes) {
    if (cluster.length === 0 || node.startMin < clusterEnd) {
      cluster.push(node);
      clusterEnd = Math.max(clusterEnd, node.endMin);
    } else {
      groups.push(cluster);
      cluster = [node];
      clusterEnd = node.endMin;
    }
  }
  if (cluster.length) groups.push(cluster);

  const out: LaidOutTimelineItem[] = [];
  for (const group of groups) {
    const columns = Math.max(1, ...group.map((n) => n.column + 1));
    for (const node of group) {
      out.push({
        item: node.item,
        startMin: node.startMin,
        endMin: node.endMin,
        column: node.column,
        columns,
        top: (node.startMin / 60) * HOUR_HEIGHT,
        height: ((node.endMin - node.startMin) / 60) * HOUR_HEIGHT,
      });
    }
  }
  return out;
}

export function contrastTextFor(background: string): string {
  const raw = background.trim().replace(/^#/, "");
  let r = 0;
  let g = 0;
  let b = 0;
  if (/^[0-9a-fA-F]{3}$/.test(raw)) {
    r = parseInt(raw[0]! + raw[0]!, 16);
    g = parseInt(raw[1]! + raw[1]!, 16);
    b = parseInt(raw[2]! + raw[2]!, 16);
  } else if (/^[0-9a-fA-F]{6}/.test(raw)) {
    r = parseInt(raw.slice(0, 2), 16);
    g = parseInt(raw.slice(2, 4), 16);
    b = parseInt(raw.slice(4, 6), 16);
  } else {
    return "#14110e";
  }
  const luminance = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
  return luminance > 0.62 ? "#14110e" : "#f3f1ec";
}

export function groupAgendaItems(
  items: TimelineItem[],
  fromDate: Date,
): Array<{ key: string; date: Date; label: string; items: TimelineItem[] }> {
  const from = startOfDay(fromDate).getTime();
  const sorted = [...items]
    .filter((i) => new Date(i.end_at).getTime() >= from)
    .sort(
      (a, b) =>
        new Date(a.start_at).getTime() - new Date(b.start_at).getTime(),
    );

  const groups: Array<{
    key: string;
    date: Date;
    label: string;
    items: TimelineItem[];
  }> = [];
  const map = new Map<string, (typeof groups)[number]>();

  for (const item of sorted) {
    const date = startOfDay(new Date(item.start_at));
    const key = format(date, "yyyy-MM-dd");
    let group = map.get(key);
    if (!group) {
      group = {
        key,
        date,
        label: format(date, "EEEE, MMMM d"),
        items: [],
      };
      map.set(key, group);
      groups.push(group);
    }
    group.items.push(item);
  }
  return groups;
}
