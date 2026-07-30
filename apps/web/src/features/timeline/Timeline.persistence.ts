import {
  DEFAULT_TIMELINE_FILTERS,
  type TimelineFilterKey,
} from "./Timeline.colors";
import type { TimelineView } from "./Timeline.types";

const VIEW_KEY = "donna-timeline-view";
const FILTERS_KEY = "donna-timeline-filters";

export function loadTimelineView(fallback: TimelineView = "week"): TimelineView {
  if (typeof window === "undefined") return fallback;
  try {
    const raw = window.localStorage.getItem(VIEW_KEY);
    if (raw === "day" || raw === "week" || raw === "month" || raw === "agenda") {
      return raw;
    }
  } catch {
    // ignore
  }
  return fallback;
}

export function saveTimelineView(view: TimelineView): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(VIEW_KEY, view);
  } catch {
    // ignore
  }
}

export function loadTimelineFilters(): Record<TimelineFilterKey, boolean> {
  if (typeof window === "undefined") return { ...DEFAULT_TIMELINE_FILTERS };
  try {
    const raw = window.localStorage.getItem(FILTERS_KEY);
    if (!raw) return { ...DEFAULT_TIMELINE_FILTERS };
    const parsed = JSON.parse(raw) as Partial<Record<TimelineFilterKey, boolean>>;
    return { ...DEFAULT_TIMELINE_FILTERS, ...parsed };
  } catch {
    return { ...DEFAULT_TIMELINE_FILTERS };
  }
}

export function saveTimelineFilters(
  filters: Record<TimelineFilterKey, boolean>,
): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(FILTERS_KEY, JSON.stringify(filters));
  } catch {
    // ignore
  }
}
