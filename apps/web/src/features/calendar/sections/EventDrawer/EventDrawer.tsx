"use client";

import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";

import { Icon } from "@/components/common";

import type { CalendarEvent, CalendarSource } from "../../Calendar.types";
import {
  formatEventDate,
  formatEventTime,
  parseAttendees,
  parseOrganizer,
} from "../../Calendar.utils";
import { isRecurring } from "../../Calendar.layout";
import { drawerStyles as styles } from "./EventDrawer.styles";

type EventDrawerProps = {
  event: CalendarEvent | null;
  source?: CalendarSource;
  color: string;
  onClose: () => void;
};

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  if (!children) {
    return null;
  }
  return (
    <div className={styles.field}>
      <p className={styles.label}>{label}</p>
      <div className={styles.value}>{children}</div>
    </div>
  );
}

export function EventDrawer({ event, source, color, onClose }: EventDrawerProps) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!event) {
      return;
    }
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    };
    window.addEventListener("keydown", onKey);
    panelRef.current?.focus();
    return () => {
      document.body.style.overflow = previous;
      window.removeEventListener("keydown", onKey);
    };
  }, [event, onClose]);

  if (!event || typeof document === "undefined") {
    return null;
  }

  const organizer = parseOrganizer(event.organizer);
  const attendees = parseAttendees(event.attendees);

  return createPortal(
    <div className={styles.root} role="presentation">
      <button
        type="button"
        className={styles.backdrop}
        aria-label="Close event details"
        onClick={onClose}
      />
      <aside
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className={styles.panel}
      >
        <header className={styles.header}>
          <span
            className={styles.colorBar}
            style={{ backgroundColor: color }}
            aria-hidden
          />
          <div className={styles.headerText}>
            <h2 id={titleId} className={styles.title}>
              {event.title || "(No title)"}
            </h2>
            <p className={styles.subtitle}>
              {formatEventDate(event)} · {formatEventTime(event)}
            </p>
          </div>
          <button
            type="button"
            className={styles.close}
            aria-label="Close"
            onClick={onClose}
          >
            <Icon name="close" className="h-4 w-4" />
          </button>
        </header>

        <div className={styles.body}>
          <Field label="Timezone">
            {event.timezone || "UTC"}
          </Field>
          <Field label="Calendar">{source?.name ?? "Unknown calendar"}</Field>
          <Field label="Location">{event.location}</Field>
          <Field label="Description">{event.description}</Field>
          <Field label="Organizer">
            {organizer
              ? organizer.displayName || organizer.email
              : null}
          </Field>
          <Field label="Attendees">
            {attendees.length > 0 ? (
              <ul className={styles.list}>
                {attendees.map((a, i) => (
                  <li key={`${a.email ?? a.displayName ?? i}`}>
                    {a.displayName || a.email || "Guest"}
                    {a.responseStatus ? (
                      <span className="text-donna-faint"> · {a.responseStatus}</span>
                    ) : null}
                  </li>
                ))}
              </ul>
            ) : null}
          </Field>
          <Field label="Provider status">
            <span className={styles.chip}>{event.status}</span>
            {isRecurring(event) ? (
              <span className={`${styles.chip} ml-2`}>Recurring</span>
            ) : null}
            {event.all_day ? (
              <span className={`${styles.chip} ml-2`}>All day</span>
            ) : null}
          </Field>
        </div>
        <footer className={styles.footer}>Read-only · Synced from your calendar</footer>
      </aside>
    </div>,
    document.body,
  );
}
