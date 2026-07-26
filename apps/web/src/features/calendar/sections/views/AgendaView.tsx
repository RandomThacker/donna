"use client";

import { useVirtualizer } from "@tanstack/react-virtual";
import { useEffect, useMemo, useRef } from "react";

import { Icon } from "@/components/common";

import { isRecurring } from "../../Calendar.layout";
import type { AgendaGroup, CalendarEvent } from "../../Calendar.types";
import {
  format,
  formatEventTime,
  isToday,
  parseISO,
  startOfDay,
} from "../../Calendar.utils";
import { timelineStyles as styles } from "./view.styles";

type AgendaViewProps = {
  events: CalendarEvent[];
  colorFor: (sourceId: string) => string;
  onEventClick: (event: CalendarEvent) => void;
  onNearEnd: () => void;
};

function groupEvents(events: CalendarEvent[]): AgendaGroup[] {
  const map = new Map<string, AgendaGroup>();
  const sorted = [...events].sort(
    (a, b) =>
      new Date(a.start_time).getTime() - new Date(b.start_time).getTime(),
  );

  for (const event of sorted) {
    const date = startOfDay(parseISO(event.start_time));
    const key = format(date, "yyyy-MM-dd");
    let group = map.get(key);
    if (!group) {
      const label = isToday(date)
        ? `Today · ${format(date, "MMMM d")}`
        : format(date, "EEEE, MMMM d");
      group = { key, date, label, events: [] };
      map.set(key, group);
    }
    group.events.push(event);
  }

  return Array.from(map.values());
}

export function AgendaView({
  events,
  colorFor,
  onEventClick,
  onNearEnd,
}: AgendaViewProps) {
  const parentRef = useRef<HTMLDivElement>(null);
  const groups = useMemo(() => groupEvents(events), [events]);

  const virtualizer = useVirtualizer({
    count: groups.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index) => {
      const group = groups[index];
      return 52 + (group?.events.length ?? 1) * 76;
    },
    overscan: 4,
  });

  const items = virtualizer.getVirtualItems();

  useEffect(() => {
    const last = items[items.length - 1];
    if (!last) {
      return;
    }
    if (last.index >= groups.length - 3) {
      onNearEnd();
    }
  }, [items, groups.length, onNearEnd]);

  if (groups.length === 0) {
    return <div className={styles.empty}>No upcoming events.</div>;
  }

  return (
    <div ref={parentRef} className={styles.agendaRoot}>
      <div
        style={{
          height: virtualizer.getTotalSize(),
          width: "100%",
          position: "relative",
        }}
      >
        {items.map((row) => {
          const group = groups[row.index];
          if (!group) {
            return null;
          }
          return (
            <div
              key={group.key}
              className={styles.agendaGroup}
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                width: "100%",
                transform: `translateY(${row.start}px)`,
              }}
              ref={virtualizer.measureElement}
              data-index={row.index}
            >
              <div className={styles.agendaSticky}>
                <p className={styles.agendaDate}>{group.label}</p>
              </div>
              <div className={styles.agendaList}>
                {group.events.map((event) => (
                  <button
                    key={event.id}
                    type="button"
                    className={styles.agendaCard}
                    onClick={() => onEventClick(event)}
                  >
                    <span
                      className={styles.agendaColor}
                      style={{
                        backgroundColor: colorFor(event.calendar_source_id),
                      }}
                      aria-hidden
                    />
                    <div className={styles.agendaBody}>
                      <p className={styles.agendaTitle}>
                        {event.title || "(No title)"}
                      </p>
                      <div className={styles.agendaMeta}>
                        <span>{formatEventTime(event)}</span>
                        {event.location ? (
                          <span className="inline-flex items-center gap-1">
                            <Icon name="mapPin" className="h-3 w-3" />
                            {event.location}
                          </span>
                        ) : null}
                        {isRecurring(event) ? (
                          <Icon name="repeat" className="h-3 w-3" />
                        ) : null}
                        {event.all_day ? <span>All day</span> : null}
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
