"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { Icon } from "@/components/common";

import type { DonnaNotification } from "../../Notifications.types";
import {
  availableActions,
  buildStatusTimeline,
  calendarHrefForOccurrence,
  failureReason,
  formatClock,
  formatDateTime,
  isDevBuild,
  notificationSource,
  parsePayload,
  statusColor,
  statusLabel,
} from "../../Notifications.utils";
import { detailsStyles as styles } from "./NotificationDetails.styles";

type Props = {
  notification: DonnaNotification;
  saving?: boolean;
  onRead: (id: string) => Promise<void>;
  onDismiss: (id: string) => Promise<void>;
  onCloseCenter: () => void;
};

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  if (children == null || children === "") return null;
  return (
    <div className={styles.field}>
      <p className={styles.label}>{label}</p>
      <div className={styles.value}>{children}</div>
    </div>
  );
}

export function NotificationDetails({
  notification,
  saving,
  onRead,
  onDismiss,
  onCloseCenter,
}: Props) {
  const router = useRouter();
  const [devOpen, setDevOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const payload = useMemo(
    () => parsePayload(notification.payload),
    [notification.payload],
  );
  const timeline = useMemo(
    () => buildStatusTimeline(notification),
    [notification],
  );
  const actions = availableActions(notification.status);
  const fail = failureReason(notification);
  const color = statusColor(notification.status);
  const showDev = isDevBuild();

  async function openInTimeline() {
    const occurrenceId =
      notification.occurrence_id || payload.occurrenceId || null;
    if (!occurrenceId) return;
    if (notification.status.toUpperCase() === "SENT") {
      try {
        await onRead(notification.id);
      } catch {
        // Still navigate even if mark-read fails.
      }
    }
    onCloseCenter();
    router.push(calendarHrefForOccurrence(occurrenceId));
  }

  async function copyDev() {
    const blob = JSON.stringify(
      {
        id: notification.id,
        public_id: notification.public_id,
        occurrence_id: notification.occurrence_id,
        timeline_item_parent_id: notification.timeline_item_parent_id,
        payload: notification.payload,
        channel_delivery_status: notification.channel_delivery_status,
      },
      null,
      2,
    );
    try {
      await navigator.clipboard.writeText(blob);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1200);
    } catch {
      // ignore
    }
  }

  const occurrenceId =
    notification.occurrence_id || payload.occurrenceId || null;

  return (
    <div className={styles.root}>
      <div>
        <h3 className={styles.title}>{notification.title}</h3>
        {notification.body ? (
          <p className={styles.body}>{notification.body}</p>
        ) : null}
        <div className={styles.metaRow}>
          <span
            className={styles.statusChip}
            style={{ color, backgroundColor: `${color}22` }}
          >
            {statusLabel(notification.status)}
          </span>
          <span className={styles.chip}>{notificationSource(notification)}</span>
          {notification.notification_type ? (
            <span className={styles.chip}>{notification.notification_type}</span>
          ) : null}
        </div>
      </div>

      {fail && notification.status.toUpperCase() === "FAILED" ? (
        <p className={styles.fail}>{fail}</p>
      ) : null}

      <div className={styles.actions}>
        {occurrenceId ? (
          <button
            type="button"
            className={styles.primary}
            onClick={() => void openInTimeline()}
          >
            Open in Calendar
          </button>
        ) : null}
        {actions.includes("read") ? (
          <button
            type="button"
            className={styles.secondary}
            disabled={saving}
            onClick={() => void onRead(notification.id)}
          >
            Mark Read
          </button>
        ) : null}
        {actions.includes("dismiss") ? (
          <button
            type="button"
            className={styles.secondary}
            disabled={saving}
            onClick={() => void onDismiss(notification.id)}
          >
            Dismiss
          </button>
        ) : null}
      </div>

      <div className={styles.timeline}>
        <p className={styles.timelineTitle}>Status Timeline</p>
        {timeline.map((step) => (
          <div key={step.id} className={styles.step}>
            <span className={styles.stepTime}>
              {step.at && step.done ? formatClock(step.at) : "—"}
            </span>
            <span
              className={
                step.done ? styles.stepLabel : `${styles.stepLabel} ${styles.stepPending}`
              }
            >
              {step.label}
            </span>
          </div>
        ))}
      </div>

      <Field label="Scheduled">{formatDateTime(notification.scheduled_for)}</Field>
      <Field label="Sent">{formatDateTime(notification.sent_at)}</Field>
      <Field label="Created">{formatDateTime(notification.created_at)}</Field>
      <Field label="Channels">
        {notification.delivery_channels?.length
          ? notification.delivery_channels.join(", ")
          : "—"}
      </Field>
      <Field label="Channel delivery">
        {notification.channel_delivery_status
          ? JSON.stringify(notification.channel_delivery_status)
          : "—"}
      </Field>
      <Field label="Occurrence ID">{occurrenceId ?? "—"}</Field>
      <Field label="Parent ID">
        {notification.timeline_item_parent_id || payload.parentId || "—"}
      </Field>

      {showDev ? (
        <div className={styles.dev}>
          <button
            type="button"
            className={styles.devHead}
            onClick={() => setDevOpen((v) => !v)}
            aria-expanded={devOpen}
          >
            <span>Developer Info</span>
            <Icon
              name="chevronRight"
              className={`h-3.5 w-3.5 transition-transform ${devOpen ? "rotate-90" : ""}`}
            />
          </button>
          {devOpen ? (
            <div className={styles.devBody}>
              <div className="flex justify-end">
                <button
                  type="button"
                  className={styles.copyBtn}
                  onClick={() => void copyDev()}
                >
                  {copied ? "Copied" : "Copy"}
                </button>
              </div>
              <p className={styles.mono}>id: {notification.id}</p>
              <p className={styles.mono}>public_id: {notification.public_id}</p>
              <p className={styles.mono}>
                occurrence_id: {occurrenceId ?? "null"}
              </p>
              <p className={styles.mono}>
                parent_id:{" "}
                {notification.timeline_item_parent_id ||
                  payload.parentId ||
                  "null"}
              </p>
              <pre className={styles.pre}>
                {JSON.stringify(
                  {
                    payload: notification.payload,
                    channel_delivery_status:
                      notification.channel_delivery_status,
                  },
                  null,
                  2,
                )}
              </pre>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
