/** Donna calendar display timezone. Phase 1 default is India Standard Time. */
export const DEFAULT_CALENDAR_TIMEZONE = "Asia/Kolkata";

/**
 * Resolve which zone the calendar UI should use.
 * Legacy accounts stored UTC as the signup default — treat that as “use Donna default”.
 */
export function resolveCalendarTimeZone(userTz?: string | null): string {
  const tz = userTz?.trim();
  if (!tz || tz === "UTC" || tz === "Etc/UTC") {
    return DEFAULT_CALENDAR_TIMEZONE;
  }
  return tz;
}

type ZonedParts = {
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
};

function zonedParts(date: Date, timeZone: string): ZonedParts {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(date);

  const read = (type: Intl.DateTimeFormatPartTypes): number => {
    const value = parts.find((p) => p.type === type)?.value;
    return value ? Number(value) : 0;
  };

  return {
    year: read("year"),
    month: read("month"),
    day: read("day"),
    hour: read("hour"),
    minute: read("minute"),
  };
}

/** Civil YYYY-MM-DD in the given IANA timezone. */
export function zonedDateKey(date: Date, timeZone: string): string {
  const p = zonedParts(date, timeZone);
  return `${p.year}-${String(p.month).padStart(2, "0")}-${String(p.day).padStart(2, "0")}`;
}

/** UTC civil date key — used for all-day events stored as midnight UTC. */
export function utcDateKey(date: Date): string {
  const y = date.getUTCFullYear();
  const m = String(date.getUTCMonth() + 1).padStart(2, "0");
  const d = String(date.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

/** Instant for 00:00:00.000 in `timeZone` on the civil day of `date` in that zone. */
export function startOfZonedDay(date: Date, timeZone: string): Date {
  const key = zonedDateKey(date, timeZone);
  return zonedCivilToUtc(key, "00:00:00.000", timeZone);
}

/** Instant for 23:59:59.999 in `timeZone` on the civil day of `date` in that zone. */
export function endOfZonedDay(date: Date, timeZone: string): Date {
  const key = zonedDateKey(date, timeZone);
  return zonedCivilToUtc(key, "23:59:59.999", timeZone);
}

function zonedCivilToUtc(
  dateKey: string,
  time: string,
  timeZone: string,
): Date {
  // Iterate to find the UTC instant whose wall-clock in `timeZone` matches.
  const [y, m, d] = dateKey.split("-").map(Number);
  const [hh, mm, ssMs] = time.split(":");
  const ss = Number(ssMs?.split(".")[0] ?? 0);
  const ms = Number(ssMs?.split(".")[1] ?? 0);
  const guess = Date.UTC(y!, (m ?? 1) - 1, d ?? 1, Number(hh), Number(mm), ss, ms);

  // Refine: compare zoned parts of guess to desired wall time.
  let utc = guess;
  for (let i = 0; i < 3; i++) {
    const p = zonedParts(new Date(utc), timeZone);
    const asUtc = Date.UTC(p.year, p.month - 1, p.day, p.hour, p.minute, ss, ms);
    const desired = Date.UTC(y!, (m ?? 1) - 1, d ?? 1, Number(hh), Number(mm), ss, ms);
    utc += desired - asUtc;
  }
  return new Date(utc);
}

export function formatZonedTime(date: Date, timeZone: string): string {
  return new Intl.DateTimeFormat("en-US", {
    timeZone,
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  }).format(date);
}

export function isZonedToday(date: Date, timeZone: string, now = new Date()): boolean {
  return zonedDateKey(date, timeZone) === zonedDateKey(now, timeZone);
}

/** Agenda / list day key for an event in the calendar timezone. */
export function eventAgendaDateKey(
  event: { start_time: string; all_day: boolean },
  timeZone: string,
): string | null {
  const start = new Date(event.start_time);
  if (Number.isNaN(start.getTime())) {
    return null;
  }
  if (event.all_day) {
    return utcDateKey(start);
  }
  return zonedDateKey(start, timeZone);
}
