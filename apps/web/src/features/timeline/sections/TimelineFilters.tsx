"use client";

import {
  TIMELINE_COLORS,
  TIMELINE_FILTER_LABELS,
  type TimelineFilterKey,
} from "../Timeline.colors";

type Props = {
  cursor: Date;
  onSelectDay: (day: Date) => void;
  filters: Record<TimelineFilterKey, boolean>;
  onToggle: (key: TimelineFilterKey) => void;
};

const FILTER_SWATCH: Partial<Record<TimelineFilterKey, string>> = {
  google: TIMELINE_COLORS.google,
  ics: TIMELINE_COLORS.ics,
  donna_events: TIMELINE_COLORS.donnaEvent,
  donna_reminders: TIMELINE_COLORS.donnaReminder,
};

const KEYS: TimelineFilterKey[] = [
  "google",
  "ics",
  "donna_events",
  "donna_reminders",
  "completed",
  "cancelled",
];

export function TimelineFilters({
  cursor,
  onSelectDay,
  filters,
  onToggle,
}: Props) {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <p className="mb-2 text-[0.65rem] font-semibold uppercase tracking-[0.18em] text-donna-faint">
          Jump to
        </p>
        <button
          type="button"
          className="w-full rounded-xl border border-donna-hairline bg-donna-bg px-3 py-2 text-left text-sm hover:border-donna-accent/30"
          onClick={() => onSelectDay(new Date())}
        >
          Today ·{" "}
          {cursor.toLocaleDateString(undefined, {
            weekday: "short",
            month: "short",
            day: "numeric",
          })}
        </button>
      </div>
      <div>
        <p className="mb-2 text-[0.65rem] font-semibold uppercase tracking-[0.18em] text-donna-faint">
          Sources
        </p>
        <ul className="space-y-1.5">
          {KEYS.map((key) => (
            <li key={key}>
              <label className="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-donna-bg/60">
                <input
                  type="checkbox"
                  checked={filters[key]}
                  onChange={() => onToggle(key)}
                  className="accent-[var(--color-donna-accent,#c9a87c)]"
                />
                {FILTER_SWATCH[key] ? (
                  <span
                    className="h-2.5 w-2.5 rounded-full"
                    style={{ background: FILTER_SWATCH[key] }}
                    aria-hidden
                  />
                ) : null}
                <span>{TIMELINE_FILTER_LABELS[key]}</span>
              </label>
            </li>
          ))}
        </ul>
      </div>
      <p className="text-xs leading-relaxed text-donna-faint">
        Timeline reads one API — Google, ICS, Donna events and reminders —
        never the providers directly.
      </p>
    </div>
  );
}
