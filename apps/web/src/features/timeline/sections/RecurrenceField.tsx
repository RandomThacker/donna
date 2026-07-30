"use client";

import { useMemo } from "react";

const WEEKDAYS = [
  { code: "MO", label: "Mon" },
  { code: "TU", label: "Tue" },
  { code: "WE", label: "Wed" },
  { code: "TH", label: "Thu" },
  { code: "FR", label: "Fri" },
  { code: "SA", label: "Sat" },
  { code: "SU", label: "Sun" },
] as const;

export type WeekdayCode = (typeof WEEKDAYS)[number]["code"];

const PRESET_VALUES = new Set(["", "FREQ=DAILY", "FREQ=WEEKLY", "FREQ=MONTHLY"]);

const CUSTOM_SELECT = "__custom__";

export const formFieldClass =
  "w-full rounded-xl border border-donna-border bg-donna-surface-2 px-3 py-2 text-sm text-donna-text outline-none transition-colors placeholder:text-donna-faint focus:border-donna-accent/50 disabled:opacity-50";

type ParsedRecurrence = {
  selectValue: string;
  customDays: WeekdayCode[];
};

export function parseRecurrenceRule(rule: string): ParsedRecurrence {
  const trimmed = rule.trim();
  if (!trimmed) {
    return { selectValue: "", customDays: [] };
  }
  if (PRESET_VALUES.has(trimmed)) {
    return { selectValue: trimmed, customDays: [] };
  }

  const upper = trimmed.toUpperCase().replace(/^RRULE:/, "");
  const parts: Record<string, string> = {};
  for (const part of upper.split(";")) {
    const [k, v = ""] = part.split("=");
    if (k) parts[k] = v;
  }

  if (parts.FREQ === "WEEKLY" && parts.BYDAY) {
    const days = parts.BYDAY.split(",")
      .map((d) => d.trim())
      .filter((d): d is WeekdayCode =>
        WEEKDAYS.some((day) => day.code === d),
      );
    if (days.length > 0) {
      return { selectValue: CUSTOM_SELECT, customDays: days };
    }
  }

  // Unknown / advanced rule — treat as custom with empty day chips so user can rebuild.
  return { selectValue: CUSTOM_SELECT, customDays: [] };
}

export function buildCustomWeeklyRule(days: WeekdayCode[]): string | null {
  if (days.length === 0) return null;
  const order = WEEKDAYS.map((d) => d.code);
  const sorted = [...days].sort(
    (a, b) => order.indexOf(a) - order.indexOf(b),
  );
  return `FREQ=WEEKLY;BYDAY=${sorted.join(",")}`;
}

type Props = {
  value: string;
  onChange: (rule: string) => void;
  error?: string | null;
};

export function RecurrenceField({ value, onChange, error }: Props) {
  const parsed = useMemo(() => parseRecurrenceRule(value), [value]);
  const isCustom = parsed.selectValue === CUSTOM_SELECT;

  function setPreset(next: string) {
    if (next === CUSTOM_SELECT) {
      const seed =
        parsed.customDays.length > 0
          ? parsed.customDays
          : ([weekdayFromToday()] as WeekdayCode[]);
      onChange(buildCustomWeeklyRule(seed) ?? "");
      return;
    }
    onChange(next);
  }

  function toggleDay(code: WeekdayCode) {
    const set = new Set(parsed.customDays);
    if (set.has(code)) {
      set.delete(code);
    } else {
      set.add(code);
    }
    const next = [...set] as WeekdayCode[];
    onChange(buildCustomWeeklyRule(next) ?? "");
  }

  return (
    <div className="space-y-2">
      <label className="block">
        <span className="mb-1 block text-xs font-medium text-donna-muted">
          Recurrence
        </span>
        <select
          className={formFieldClass}
          value={parsed.selectValue}
          onChange={(e) => setPreset(e.target.value)}
        >
          <option value="">Does not repeat</option>
          <option value="FREQ=DAILY">Daily</option>
          <option value="FREQ=WEEKLY">Weekly</option>
          <option value="FREQ=MONTHLY">Monthly</option>
          <option value={CUSTOM_SELECT}>Custom days…</option>
        </select>
      </label>

      {isCustom ? (
        <div>
          <p className="mb-2 text-xs text-donna-muted">
            Repeat on these days
          </p>
          <div className="flex flex-wrap gap-1.5">
            {WEEKDAYS.map((day) => {
              const active = parsed.customDays.includes(day.code);
              return (
                <button
                  key={day.code}
                  type="button"
                  aria-pressed={active}
                  onClick={() => toggleDay(day.code)}
                  className={[
                    "min-w-[2.75rem] rounded-full px-2.5 py-1.5 text-xs font-medium transition-colors",
                    active
                      ? "bg-donna-accent text-donna-on-accent"
                      : "border border-donna-border bg-donna-surface-2 text-donna-muted hover:border-donna-accent/40 hover:text-donna-text",
                  ].join(" ")}
                >
                  {day.label}
                </button>
              );
            })}
          </div>
          {error ? (
            <p className="mt-1.5 text-xs text-rose-400">{error}</p>
          ) : parsed.customDays.length === 0 ? (
            <p className="mt-1.5 text-xs text-donna-faint">
              Pick at least one day
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function weekdayFromToday(): WeekdayCode {
  // JS: 0=Sun … 6=Sat → RRULE SU…SA
  const map: WeekdayCode[] = ["SU", "MO", "TU", "WE", "TH", "FR", "SA"];
  return map[new Date().getDay()] ?? "MO";
}
