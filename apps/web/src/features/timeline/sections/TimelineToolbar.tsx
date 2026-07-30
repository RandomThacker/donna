"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { TimelineView } from "../Timeline.types";

const views: TimelineView[] = ["day", "week", "month", "agenda"];

type Props = {
  title: string;
  view: TimelineView;
  onView: (view: TimelineView) => void;
  onToday: () => void;
  onPrev: () => void;
  onNext: () => void;
  onMenu: () => void;
  onCreate: () => void;
  search: string;
  onSearch: (value: string) => void;
  fetching?: boolean;
};

export function TimelineToolbar({
  title,
  view,
  onView,
  onToday,
  onPrev,
  onNext,
  onMenu,
  onCreate,
  search,
  onSearch,
  fetching,
}: Props) {
  return (
    <header className="flex shrink-0 flex-col gap-3 border-b border-donna-hairline px-3 py-3 sm:px-4">
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="grid h-9 w-9 place-items-center rounded-full text-donna-muted hover:bg-donna-surface lg:hidden"
          aria-label="Open filters"
          onClick={onMenu}
        >
          <Icon name="calendar" className="h-4 w-4" />
        </button>
        <div className="flex items-center gap-1">
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-full hover:bg-donna-surface"
            aria-label="Previous"
            onClick={onPrev}
          >
            <Icon name="chevronLeft" className="h-4 w-4" />
          </button>
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-full hover:bg-donna-surface"
            aria-label="Next"
            onClick={onNext}
          >
            <Icon name="chevronRight" className="h-4 w-4" />
          </button>
          <button
            type="button"
            className="rounded-full border border-donna-border px-3 py-1.5 text-sm hover:bg-donna-surface"
            onClick={onToday}
          >
            Today
          </button>
        </div>
        <h1 className="min-w-0 flex-1 truncate font-display text-xl tracking-tight sm:text-2xl">
          {title}
        </h1>
        <button
          type="button"
          className="inline-flex items-center gap-1.5 rounded-full bg-donna-accent px-3 py-2 text-sm font-medium text-donna-on-accent"
          onClick={onCreate}
        >
          <Icon name="plus" className="h-4 w-4" />
          <span className="hidden sm:inline">Create</span>
        </button>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <div
          className="inline-flex rounded-full border border-donna-hairline bg-donna-surface p-0.5"
          role="tablist"
          aria-label="Timeline views"
        >
          {views.map((v) => (
            <button
              key={v}
              type="button"
              role="tab"
              aria-selected={view === v}
              className={cn(
                "rounded-full px-3 py-1.5 text-xs capitalize sm:text-sm",
                view === v
                  ? "bg-donna-accent-soft text-donna-accent"
                  : "text-donna-muted hover:text-donna-text",
              )}
              onClick={() => onView(v)}
            >
              {v}
            </button>
          ))}
        </div>
        <label className="relative ml-auto min-w-[10rem] flex-1 sm:max-w-xs">
          <span className="sr-only">Search timeline</span>
          <Icon
            name="search"
            className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-donna-muted"
          />
          <input
            value={search}
            onChange={(e) => onSearch(e.target.value)}
            placeholder="Search title or description"
            className="w-full rounded-full border border-donna-hairline bg-donna-surface py-2 pl-9 pr-3 text-sm outline-none focus:border-donna-accent/40"
          />
        </label>
        {fetching ? (
          <span className="text-xs text-donna-faint" aria-live="polite">
            Updating…
          </span>
        ) : null}
      </div>
    </header>
  );
}
