"use client";

import Link from "next/link";

import { calendarAgendaHref } from "@/features/calendar/Calendar.routes";
import { useCalendarDayEvents } from "@/features/calendar/useCalendarDayEvents";

import { BentoBox } from "../BentoBox";
import { calendarEventsToTimelineItems } from "./DashboardTimeline.logic";
import { timelineStyles as styles } from "./DashboardTimeline.styles";

export function DashboardTimeline() {
  const { events, timeZone, isLoading, isError } = useCalendarDayEvents();
  const items = calendarEventsToTimelineItems(events, timeZone);

  return (
    <BentoBox className={styles.box} title="Today's timeline">
      {isLoading ? (
        <p className={styles.state}>Loading today&apos;s schedule…</p>
      ) : null}
      {!isLoading && isError ? (
        <p className={styles.state}>Couldn&apos;t load today&apos;s events.</p>
      ) : null}
      {!isLoading && !isError && items.length === 0 ? (
        <div className={styles.state}>
          <p>Nothing on the calendar today.</p>
          <Link href={calendarAgendaHref()} className={styles.link}>
            Open agenda
          </Link>
        </div>
      ) : null}
      {!isLoading && !isError && items.length > 0 ? (
        <ol className={styles.list}>
          {items.map((item) => (
            <li key={item.id}>
              <Link
                href={calendarAgendaHref(item.id)}
                className={styles.item}
              >
                <time className={styles.time}>{item.time}</time>
                <div className={styles.itemBody}>
                  <p className={styles.title}>{item.title}</p>
                  {item.meta ? <p className={styles.meta}>{item.meta}</p> : null}
                </div>
              </Link>
            </li>
          ))}
        </ol>
      ) : null}
    </BentoBox>
  );
}
