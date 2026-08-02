"use client";

import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import Link from "next/link";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import { useAuth } from "@/features/auth";
import { calendarAgendaHref } from "@/features/calendar/Calendar.routes";
import {
  endOfZonedDay,
  resolveCalendarTimeZone,
  startOfZonedDay,
} from "@/features/calendar/Calendar.timezone";
import { fetchTimeline } from "@/features/timeline/Timeline.api";
import { timelineQueryKeys } from "@/features/timeline/Timeline.logic";

import { BentoBox, bentoBoxStyles } from "../BentoBox";
import { timelineItemsToDashboardItems } from "./DashboardTimeline.logic";
import { timelineStyles as styles } from "./DashboardTimeline.styles";

export function DashboardTimeline() {
  const { user } = useAuth();
  const timeZone = resolveCalendarTimeZone(user?.timezone);
  const day = useMemo(() => new Date(), []);

  const range = useMemo(() => {
    const from = startOfZonedDay(day, timeZone);
    const to = endOfZonedDay(day, timeZone);
    return { from: from.toISOString(), to: to.toISOString() };
  }, [day, timeZone]);

  const timelineQuery = useQuery({
    queryKey: timelineQueryKeys.range(range.from, range.to),
    queryFn: ({ signal }) =>
      fetchTimeline({ from: range.from, to: range.to, signal }),
    staleTime: 30_000,
    refetchOnMount: "always",
  });

  const items = useMemo(
    () =>
      timelineItemsToDashboardItems(timelineQuery.data?.items ?? [], timeZone),
    [timelineQuery.data?.items, timeZone],
  );

  return (
    <BentoBox
      className={cn(styles.box, bentoBoxStyles.fixedPanel)}
      title="Today's timeline"
    >
      <div className={bentoBoxStyles.scrollBody}>
        {timelineQuery.isLoading ? (
          <p className={styles.state}>Loading today&apos;s schedule…</p>
        ) : null}
        {!timelineQuery.isLoading && timelineQuery.isError ? (
          <p className={styles.state}>Couldn&apos;t load today&apos;s schedule.</p>
        ) : null}
        {!timelineQuery.isLoading &&
        !timelineQuery.isError &&
        items.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyInner}>
              <span className={styles.emptyIcon} aria-hidden>
                <Icon name="calendar" className="h-5 w-5" />
              </span>
              <h3 className={styles.emptyTitle}>Clear day</h3>
              <p className={styles.emptyBody}>
                Nothing on the books. Enjoy the space — or add something when
                you&apos;re ready.
              </p>
              <Link href={calendarAgendaHref()} className={styles.emptyCta}>
                Open agenda
                <Icon name="arrow" className="h-3.5 w-3.5" />
              </Link>
            </div>
          </div>
        ) : null}
        {!timelineQuery.isLoading &&
        !timelineQuery.isError &&
        items.length > 0 ? (
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
                    {item.meta ? (
                      <p className={styles.meta}>{item.meta}</p>
                    ) : null}
                  </div>
                </Link>
              </li>
            ))}
          </ol>
        ) : null}
      </div>
    </BentoBox>
  );
}
