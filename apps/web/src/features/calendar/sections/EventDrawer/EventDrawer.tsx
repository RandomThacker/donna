"use client";

import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";

import { Icon } from "@/components/common";

import type { CalendarEvent, CalendarSource } from "../../Calendar.types";
import { resolveCalendarTimeZone } from "../../Calendar.timezone";
import {
  formatEventDate,
  formatEventTime,
  parseAttendees,
  parseOrganizer,
} from "../../Calendar.utils";
import { isRecurring } from "../../Calendar.layout";
import { drawerStyles as styles } from "./EventDrawer.styles";
import { linkifyPlainText } from "./EventDrawer.linkify";

type EventDrawerProps = {
  event: CalendarEvent | null;
  source?: CalendarSource;
  calendarLabel?: string;
  color: string;
  onClose: () => void;
  timeZone?: string;
  onEdit?: (event: CalendarEvent) => void;
  onDelete?: (event: CalendarEvent) => void;
  deleting?: boolean;
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

function DescriptionField({ text }: { text: string }) {
  const trimmed = text.trim();
  if (!trimmed) {
    return null;
  }
  return <Field label="Description">{linkifyPlainText(trimmed)}</Field>;
}

export function EventDrawer({
  event,
  source,
  calendarLabel,
  color,
  onClose,
  timeZone: timeZoneProp,
  onEdit,
  onDelete,
  deleting,
}: EventDrawerProps) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const timeZone = resolveCalendarTimeZone(timeZoneProp);

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
  const editable = event.read_only === false && Boolean(event.mutation_id);

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
              {formatEventDate(event)} · {formatEventTime(event, timeZone)}
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
          <Field label="Timezone">{event.timezone || timeZone}</Field>
          <Field label="Calendar">
            {calendarLabel || source?.name || "Unknown calendar"}
          </Field>
          <Field label="Type">
            {event.timeline_type === "REMINDER" ? "Reminder" : "Event"}
          </Field>
          <Field label="Location">{event.location}</Field>
          {event.description ? (
            <DescriptionField text={event.description} />
          ) : null}
          <Field label="Organizer">
            {organizer ? organizer.displayName || organizer.email : null}
          </Field>
          <Field label="Attendees">
            {attendees.length > 0 ? (
              <ul className={styles.list}>
                {attendees.map((a, i) => (
                  <li key={`${a.email ?? a.displayName ?? i}`}>
                    {a.displayName || a.email || "Guest"}
                    {a.responseStatus ? (
                      <span className="text-donna-faint">
                        {" "}
                        · {a.responseStatus}
                      </span>
                    ) : null}
                  </li>
                ))}
              </ul>
            ) : null}
          </Field>
          <Field label="Status">
            <span className={styles.chip}>{event.status}</span>
            {isRecurring(event) || event.recurrence_rule ? (
              <span className={`${styles.chip} ml-2`}>Recurring</span>
            ) : null}
            {event.all_day ? (
              <span className={`${styles.chip} ml-2`}>All day</span>
            ) : null}
          </Field>
          {event.recurrence_rule ? (
            <Field label="Recurrence">{event.recurrence_rule}</Field>
          ) : null}
        </div>
        <footer className={styles.footer}>
          {editable ? (
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="inline-flex items-center gap-1.5 rounded-full border border-donna-border px-3 py-1.5 text-xs text-donna-text"
                onClick={() => onEdit?.(event)}
              >
                <Icon name="compose" className="h-3.5 w-3.5" />
                Edit
              </button>
              <button
                type="button"
                className="inline-flex items-center gap-1.5 rounded-full border border-rose-500/40 px-3 py-1.5 text-xs text-rose-300"
                disabled={deleting}
                onClick={() => {
                  if (
                    window.confirm(
                      `Delete “${event.title}”? This can’t be undone.`,
                    )
                  ) {
                    onDelete?.(event);
                  }
                }}
              >
                <Icon name="trash" className="h-3.5 w-3.5" />
                {deleting ? "Deleting…" : "Delete"}
              </button>
            </div>
          ) : (
            <span>Read-only · Synced from your calendar</span>
          )}
        </footer>
      </aside>
    </div>,
    document.body,
  );
}
