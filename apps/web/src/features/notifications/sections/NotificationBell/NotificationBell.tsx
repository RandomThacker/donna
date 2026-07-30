"use client";

import { Icon } from "@/components/common";

import { bellStyles as styles } from "./NotificationBell.styles";

type Props = {
  unreadCount: number;
  onClick: () => void;
  buttonRef?: React.Ref<HTMLButtonElement>;
};

export function NotificationBell({ unreadCount, onClick, buttonRef }: Props) {
  const label =
    unreadCount > 0
      ? `Notifications, ${unreadCount} unread`
      : "Notifications";

  return (
    <button
      ref={buttonRef}
      type="button"
      className={styles.root}
      aria-label={label}
      onClick={onClick}
    >
      <Icon name="bell" className="h-4 w-4" />
      {unreadCount > 0 ? (
        <span className={styles.badge} aria-hidden>
          {unreadCount > 99 ? "99+" : unreadCount}
        </span>
      ) : null}
    </button>
  );
}
