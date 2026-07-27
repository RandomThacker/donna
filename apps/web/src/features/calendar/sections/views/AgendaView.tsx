"use client";

import { useEffect, useMemo, useRef } from "react";

import { Icon } from "@/components/common";

import { groupAgendaEvents } from "../../Calendar.agenda";
import { isRecurring } from "../../Calendar.layout";
import type { CalendarEvent } from "../../Calendar.types";
import { resolveCalendarTimeZone } from "../../Calendar.timezone";
import { formatEventTime } from "../../Calendar.utils";
import { timelineStyles as styles } from "./view.styles";

type AgendaViewProps = {
  events: CalendarEvent[];
  colorFor: (sourceId: string) => string;
  onEventClick: (event: CalendarEvent) => void;
  onNearEnd: () => void;
  /** Agenda starts on this civil day (inclusive), in `timeZone`. */
  fromDate: Date;
  timeZone?: string;
};

function AgendaEventRow({
  event,
  color,
  timeZone,
  onEventClick,
}: {
  event: CalendarEvent;
  color: string;
  timeZone: string;
  onEventClick: (event: CalendarEvent) => void;
}) {
  return (
    <button
      type="button"
      className={styles.agendaCard}
      onClick={() => onEventClick(event)}
    >
      <span
        className={styles.agendaColor}
        style={{ backgroundColor: color }}
        aria-hidden
      />
      <div className={styles.agendaBody}>
        <p className={styles.agendaTitle}>{event.title || "(No title)"}</p>
        <div className={styles.agendaMeta}>
          <span>{formatEventTime(event, timeZone)}</span>
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
  );
}

export function AgendaView({
  events,
  colorFor,
  onEventClick,
  onNearEnd,
  fromDate,
  timeZone: timeZoneProp,
}: AgendaViewProps) {
  const timeZone = resolveCalendarTimeZone(timeZoneProp);
  const parentRef = useRef<HTMLDivElement>(null);
  const groups = useMemo(
    () => groupAgendaEvents(events, fromDate, timeZone),
    [events, fromDate, timeZone],
  );

  useEffect(() => {
    const root = parentRef.current;
    if (!root) {
      return;
    }
    const onScroll = () => {
      const remaining = root.scrollHeight - root.scrollTop - root.clientHeight;
      if (remaining < 480) {
        onNearEnd();
      }
    };
    root.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => root.removeEventListener("scroll", onScroll);
  }, [onNearEnd, groups.length]);

  if (groups.length === 0) {
    return <div className={styles.empty}>No upcoming events.</div>;
  }

  return (
    <div ref={parentRef} className={styles.agendaRoot}>
      {groups.map((group) => (
        <section key={group.key} className={styles.agendaGroup}>
          <div className={styles.agendaSticky}>
            <p className={styles.agendaDate}>{group.label}</p>
          </div>
          <div className={styles.agendaList}>
            {group.events.map((event) => (
              <AgendaEventRow
                key={event.id}
                event={event}
                color={colorFor(event.calendar_source_id)}
                timeZone={timeZone}
                onEventClick={onEventClick}
              />
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
