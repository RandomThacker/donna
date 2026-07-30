"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { DonnaNotification } from "../../Notifications.types";
import {
  availableActions,
  cardIconName,
  failureReason,
  formatClock,
  formatDateTime,
  formatRelativeTime,
  notificationSource,
  statusColor,
  statusLabel,
} from "../../Notifications.utils";
import { cardStyles as styles } from "./NotificationCard.styles";

type Props = {
  notification: DonnaNotification;
  saving?: boolean;
  onOpen: (id: string) => void;
  onRead: (id: string) => void;
  onDismiss: (id: string) => void;
};

export function NotificationCard({
  notification,
  saving,
  onOpen,
  onRead,
  onDismiss,
}: Props) {
  const actions = availableActions(notification.status);
  const fail = failureReason(notification);
  const color = statusColor(notification.status);
  const icon = cardIconName(notification);
  const relative = formatRelativeTime(
    notification.scheduled_for || notification.created_at,
  );

  return (
    <div className={styles.root}>
      <button
        type="button"
        className="w-full text-left"
        onClick={() => onOpen(notification.id)}
      >
        <div className={styles.top}>
          <span className={styles.iconWrap} aria-hidden>
            <Icon name={icon} className="h-3.5 w-3.5" />
          </span>
          <div className={styles.main}>
            <div className={styles.titleRow}>
              <p className={styles.title}>{notification.title}</p>
              {relative ? <span className={styles.relative}>{relative}</span> : null}
            </div>
            {notification.body ? (
              <p className={styles.body}>{notification.body}</p>
            ) : null}
            <div className={styles.meta}>
              <span
                className={styles.chip}
                style={{
                  color,
                  backgroundColor: `${color}22`,
                }}
              >
                {statusLabel(notification.status)}
              </span>
              <span className={styles.mutedChip}>
                {notificationSource(notification)}
              </span>
              {notification.notification_type ? (
                <span className={styles.mutedChip}>
                  {notification.notification_type}
                </span>
              ) : null}
            </div>
            <div className={styles.times}>
              {notification.scheduled_for ? (
                <p>Scheduled {formatDateTime(notification.scheduled_for)}</p>
              ) : null}
              {notification.sent_at ? (
                <p>Sent {formatClock(notification.sent_at)}</p>
              ) : null}
            </div>
            {fail && notification.status.toUpperCase() === "FAILED" ? (
              <p className={styles.fail}>{fail}</p>
            ) : null}
          </div>
        </div>
      </button>
      {actions.length > 0 ? (
        <div className={styles.actions}>
          {actions.includes("read") ? (
            <button
              type="button"
              className={styles.actionBtn}
              disabled={saving}
              onClick={(e) => {
                e.stopPropagation();
                onRead(notification.id);
              }}
            >
              Mark Read
            </button>
          ) : null}
          {actions.includes("dismiss") ? (
            <button
              type="button"
              className={cn(styles.actionBtn)}
              disabled={saving}
              onClick={(e) => {
                e.stopPropagation();
                onDismiss(notification.id);
              }}
            >
              Dismiss
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
