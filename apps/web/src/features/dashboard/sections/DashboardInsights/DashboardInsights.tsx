import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { useAuth } from "@/features/auth";
import { useCalendarDayEvents } from "@/features/calendar/useCalendarDayEvents";
import { resolveCalendarTimeZone } from "@/features/calendar/Calendar.timezone";
import { fetchTaskDay } from "@/features/tasks/Tasks.api";
import { taskQueryKeys } from "@/features/tasks/Tasks.logic";

import { BentoBox } from "../BentoBox";
import {
  buildDashboardInsights,
  todayDateKey,
  tomorrowAnchor,
} from "./DashboardInsights.logic";
import { insightsStyles as styles } from "./DashboardInsights.styles";

export function DashboardInsights() {
  const { user } = useAuth();
  const timeZone = resolveCalendarTimeZone(user?.timezone);
  const today = useMemo(() => new Date(), []);
  const tomorrow = useMemo(
    () => tomorrowAnchor(today, timeZone),
    [today, timeZone],
  );
  const dateKey = todayDateKey(today, timeZone);

  const todayCal = useCalendarDayEvents(today);
  const tomorrowCal = useCalendarDayEvents(tomorrow);

  const tasksQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const insights = useMemo(
    () =>
      buildDashboardInsights({
        now: new Date(),
        timeZone,
        todayEvents: todayCal.events.map((event) => ({
          start_time: event.start_time,
          end_time: event.end_time,
          all_day: event.all_day,
          title: event.title,
        })),
        tomorrowEvents: tomorrowCal.events.map((event) => ({
          start_time: event.start_time,
          end_time: event.end_time,
          all_day: event.all_day,
          title: event.title,
        })),
        todayTasks: (tasksQuery.data?.occurrences ?? []).map((task) => ({
          completed: task.completed,
        })),
      }),
    [
      timeZone,
      todayCal.events,
      tomorrowCal.events,
      tasksQuery.data?.occurrences,
    ],
  );

  const isLoading =
    todayCal.isLoading || tomorrowCal.isLoading || tasksQuery.isLoading;
  const isError =
    todayCal.isError || tomorrowCal.isError || tasksQuery.isError;

  return (
    <BentoBox className={styles.box} title="Donna noticed">
      {isLoading ? (
        <p className={styles.state}>Looking over your day…</p>
      ) : null}
      {!isLoading && isError ? (
        <p className={styles.state}>Couldn&apos;t read your day just now.</p>
      ) : null}
      {!isLoading && !isError ? (
        <ul className={styles.list}>
          {insights.map((insight) => (
            <li key={insight.id} className={styles.item}>
              {insight.text}
            </li>
          ))}
        </ul>
      ) : null}
    </BentoBox>
  );
}
