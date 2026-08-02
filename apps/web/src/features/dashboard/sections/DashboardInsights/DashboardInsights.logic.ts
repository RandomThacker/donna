import {
  endOfZonedDay,
  formatZonedTime,
  startOfZonedDay,
  zonedDateKey,
} from "@/features/calendar/Calendar.timezone";

export type InsightEvent = {
  start_time: string;
  end_time: string;
  all_day: boolean;
  title?: string;
};

export type InsightTask = {
  completed: boolean;
};

export type BuiltInsight = {
  id: string;
  text: string;
};

export type BuildInsightsInput = {
  now: Date;
  timeZone: string;
  todayEvents: InsightEvent[];
  tomorrowEvents: InsightEvent[];
  todayTasks: InsightTask[];
};

type RankedInsight = BuiltInsight & { rank: number };

const MIN_FOCUS_MINUTES = 45;
const SOON_MEETING_MINUTES = 90;
const BUSY_TOMORROW_COUNT = 4;
const MAX_INSIGHTS = 3;

function timedBlocks(
  events: InsightEvent[],
): Array<{ start: number; end: number }> {
  return events
    .filter((event) => !event.all_day)
    .map((event) => ({
      start: new Date(event.start_time).getTime(),
      end: new Date(event.end_time).getTime(),
    }))
    .filter((block) => block.end > block.start)
    .sort((a, b) => a.start - b.start);
}

function mergeBlocks(
  blocks: Array<{ start: number; end: number }>,
): Array<{ start: number; end: number }> {
  if (blocks.length === 0) return [];
  const merged: Array<{ start: number; end: number }> = [
    { ...blocks[0]! },
  ];
  for (let i = 1; i < blocks.length; i++) {
    const current = blocks[i]!;
    const last = merged[merged.length - 1]!;
    if (current.start <= last.end) {
      last.end = Math.max(last.end, current.end);
    } else {
      merged.push({ ...current });
    }
  }
  return merged;
}

function formatDuration(minutes: number): string {
  if (minutes >= 120) {
    const hours = Math.round(minutes / 60);
    return hours === 2 ? "Two-hour" : `${hours}-hour`;
  }
  if (minutes >= 90) {
    return "90-minute";
  }
  if (minutes >= 60) {
    return "One-hour";
  }
  return `${minutes}-minute`;
}

function focusWindowInsight(
  now: Date,
  timeZone: string,
  todayEvents: InsightEvent[],
): RankedInsight | null {
  const dayStart = startOfZonedDay(now, timeZone).getTime();
  const dayEnd = endOfZonedDay(now, timeZone).getTime();
  const cursor = Math.max(now.getTime(), dayStart);
  if (cursor >= dayEnd) return null;

  const blocks = mergeBlocks(
    timedBlocks(todayEvents).map((block) => ({
      start: Math.max(block.start, dayStart),
      end: Math.min(block.end, dayEnd),
    })),
  ).filter((block) => block.end > cursor);

  const gaps: Array<{ start: number; end: number }> = [];
  let edge = cursor;
  for (const block of blocks) {
    if (block.start > edge) {
      gaps.push({ start: edge, end: block.start });
    }
    edge = Math.max(edge, block.end);
  }
  if (edge < dayEnd) {
    gaps.push({ start: edge, end: dayEnd });
  }

  let best: { start: number; end: number } | null = null;
  for (const gap of gaps) {
    const minutes = Math.floor((gap.end - gap.start) / 60_000);
    if (minutes < MIN_FOCUS_MINUTES) continue;
    if (!best || gap.end - gap.start > best.end - best.start) {
      best = gap;
    }
  }
  if (!best) return null;

  const minutes = Math.floor((best.end - best.start) / 60_000);
  const startDate = new Date(best.start);
  const hour = Number(
    new Intl.DateTimeFormat("en-US", {
      timeZone,
      hour: "numeric",
      hourCycle: "h23",
    })
      .formatToParts(startDate)
      .find((part) => part.type === "hour")?.value ?? "0",
  );
  const duration = formatDuration(minutes);
  let text: string;
  if (hour >= 12 && hour < 14) {
    text = `${duration} focus window after lunch.`;
  } else if (best.start <= now.getTime() + 5 * 60_000) {
    text = `${duration} open block right now.`;
  } else {
    const when = formatZonedTime(startDate, timeZone);
    text = `${duration} focus window starting at ${when}.`;
  }
  return { id: "focus-window", text, rank: 20 };
}

function taskProgressInsight(tasks: InsightTask[]): RankedInsight | null {
  if (tasks.length === 0) return null;
  const done = tasks.filter((task) => task.completed).length;
  const total = tasks.length;
  if (done === total) {
    return {
      id: "tasks-done",
      text: "All of today's tasks are done.",
      rank: 30,
    };
  }
  if (done === 0) {
    return {
      id: "tasks-open",
      text:
        total === 1
          ? "1 task still open today."
          : `${total} tasks still open today.`,
      rank: 35,
    };
  }
  const pct = Math.round((done / total) * 100);
  return {
    id: "tasks-progress",
    text: `${pct}% of today's tasks done (${done} of ${total}).`,
    rank: 30,
  };
}

function tomorrowInsight(tomorrowEvents: InsightEvent[]): RankedInsight | null {
  const count = tomorrowEvents.length;
  if (count === 0) {
    return {
      id: "tomorrow-clear",
      text: "Tomorrow's calendar is clear.",
      rank: 50,
    };
  }
  if (count >= BUSY_TOMORROW_COUNT) {
    return {
      id: "tomorrow-busy",
      text: "Tomorrow looks busy — keep tonight light.",
      rank: 25,
    };
  }
  return {
    id: "tomorrow-count",
    text:
      count === 1
        ? "1 thing on tomorrow's calendar."
        : `${count} things on tomorrow's calendar.`,
    rank: 45,
  };
}

function nextMeetingInsight(
  now: Date,
  todayEvents: InsightEvent[],
): RankedInsight | null {
  const nowMs = now.getTime();
  const upcoming = todayEvents
    .filter((event) => !event.all_day)
    .map((event) => ({
      start: new Date(event.start_time).getTime(),
      title: event.title?.trim() || "Meeting",
    }))
    .filter((event) => event.start > nowMs)
    .sort((a, b) => a.start - b.start)[0];
  if (!upcoming) return null;

  const mins = Math.round((upcoming.start - nowMs) / 60_000);
  if (mins > SOON_MEETING_MINUTES) return null;
  if (mins <= 0) return null;
  if (mins < 60) {
    return {
      id: "meeting-soon",
      text: `${upcoming.title} starts in ${mins} min.`,
      rank: 10,
    };
  }
  return {
    id: "meeting-soon",
    text: `${upcoming.title} starts in about an hour.`,
    rank: 10,
  };
}

function clearDayInsight(
  now: Date,
  todayEvents: InsightEvent[],
): RankedInsight | null {
  const remaining = todayEvents.filter((event) => {
    if (event.all_day) return true;
    return new Date(event.end_time).getTime() > now.getTime();
  });
  if (remaining.length > 0) return null;
  if (todayEvents.length === 0) {
    return {
      id: "clear-day",
      text: "Nothing on the books today — enjoy the space.",
      rank: 40,
    };
  }
  return {
    id: "clear-rest",
    text: "No meetings left today.",
    rank: 40,
  };
}

/** Rule-based dashboard insights — no AI. Deterministic from calendar + tasks. */
export function buildDashboardInsights(
  input: BuildInsightsInput,
): BuiltInsight[] {
  const candidates: RankedInsight[] = [];

  const soon = nextMeetingInsight(input.now, input.todayEvents);
  if (soon) candidates.push(soon);

  const focus = focusWindowInsight(
    input.now,
    input.timeZone,
    input.todayEvents,
  );
  if (focus) candidates.push(focus);

  const tasks = taskProgressInsight(input.todayTasks);
  if (tasks) candidates.push(tasks);

  const tomorrow = tomorrowInsight(input.tomorrowEvents);
  if (tomorrow) candidates.push(tomorrow);

  const clear = clearDayInsight(input.now, input.todayEvents);
  if (clear) candidates.push(clear);

  candidates.sort((a, b) => a.rank - b.rank);

  const seen = new Set<string>();
  const out: BuiltInsight[] = [];
  for (const item of candidates) {
    if (seen.has(item.id)) continue;
    seen.add(item.id);
    out.push({ id: item.id, text: item.text });
    if (out.length >= MAX_INSIGHTS) break;
  }

  if (out.length === 0) {
    return [
      {
        id: "quiet",
        text: "Nothing stands out yet — check back as your day fills in.",
      },
    ];
  }
  return out;
}

/** Civil tomorrow date for calendar-day hooks (midday avoids DST edge cases). */
export function tomorrowAnchor(now: Date, timeZone: string): Date {
  const todayStart = startOfZonedDay(now, timeZone);
  return new Date(todayStart.getTime() + 36 * 60 * 60 * 1000);
}

export function todayDateKey(now: Date, timeZone: string): string {
  return zonedDateKey(now, timeZone);
}
