"use client";

import { format, isToday, parseISO } from "date-fns";
import { useMemo } from "react";

import { Button, Icon } from "@/components/common";
import { calendarAgendaHref } from "@/features/calendar/Calendar.routes";
import { useCalendarDayEvents } from "@/features/calendar/useCalendarDayEvents";

import { BentoBox } from "../BentoBox";
import { focusStyles as styles } from "./DashboardFocus.styles";

const EMPTY_MESSAGES = [
  "Hooray - no meetings left today. Enjoy the space 🎉",
  "Your calendar is clear. A quiet stretch is a gift.",
  "Nothing on the books. Breathe, then choose what matters.",
  "No upcoming meetings. You've got room to think.",
];

function emptyMessageForToday(): string {
  const day = new Date().getDate();
  return EMPTY_MESSAGES[day % EMPTY_MESSAGES.length]!;
}

function formatMeetingWhen(startIso: string, allDay: boolean): string {
  const start = parseISO(startIso);
  if (allDay) {
    return isToday(start) ? "All day" : format(start, "EEE · All day");
  }
  const now = new Date();
  const mins = Math.round((start.getTime() - now.getTime()) / 60_000);
  if (mins <= 0) {
    return "Happening now";
  }
  if (mins < 60) {
    return `Starts in ${mins} min · ${format(start, "h:mm a")}`;
  }
  return format(start, "h:mm a");
}

export function DashboardFocus() {
  const { events, isLoading, isError } = useCalendarDayEvents();

  const nextMeeting = useMemo(() => {
    const now = Date.now();
    return events
      .filter((event) => {
        if (event.all_day) {
          return true;
        }
        return new Date(event.end_time).getTime() > now;
      })
      .sort((a, b) => {
        if (a.all_day && !b.all_day) {
          return 1;
        }
        if (!a.all_day && b.all_day) {
          return -1;
        }
        return (
          new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
        );
      })[0];
  }, [events]);

  return (
    <BentoBox className={styles.box} title="Upcoming meeting">
      {isLoading ? (
        <p className={styles.empty}>Checking your calendar…</p>
      ) : null}
      {isError ? (
        <p className={styles.empty}>Couldn&apos;t load upcoming meetings.</p>
      ) : null}
      {!isLoading && !isError && !nextMeeting ? (
        <div>
          <h3 className={styles.goal}>All clear</h3>
          <p className={styles.emptyDetail}>{emptyMessageForToday()}</p>
        </div>
      ) : null}
      {!isLoading && !isError && nextMeeting ? (
        <>
          <h3 className={styles.goal}>{nextMeeting.title}</h3>
          <p className={styles.when}>
            {formatMeetingWhen(nextMeeting.start_time, nextMeeting.all_day)}
          </p>
          {nextMeeting.location ? (
            <p className={styles.location}>{nextMeeting.location}</p>
          ) : null}
          <Button
            href={calendarAgendaHref(nextMeeting.id)}
            size="md"
            className={styles.cta}
            iconRight={<Icon name="arrow" className="h-4 w-4" />}
          >
            Join now
          </Button>
        </>
      ) : null}
    </BentoBox>
  );
}
