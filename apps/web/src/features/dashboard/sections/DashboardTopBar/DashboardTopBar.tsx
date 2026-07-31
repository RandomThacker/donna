"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

import { Icon } from "@/components/common";
import { useNotificationsCenter } from "@/features/notifications";

import { topBarStyles as styles } from "./DashboardTopBar.styles";
import type { DashboardTopBarProps } from "./DashboardTopBar.types";

/** Mobile-only app bar — the sidebar is hidden below md. */
export function DashboardTopBar({
  profileName,
  profileInitials,
  profileAvatarUrl,
}: DashboardTopBarProps) {
  const [avatarFailed, setAvatarFailed] = useState(false);
  const { badgeCount } = useNotificationsCenter();

  useEffect(() => {
    setAvatarFailed(false);
  }, [profileAvatarUrl]);

  const showAvatar = Boolean(profileAvatarUrl) && !avatarFailed;

  return (
    <header className={styles.root}>
      <Link href="/dashboard" className={styles.brand} aria-label="Donna home">
        <span className={styles.brandMark} aria-hidden>
          <span className={styles.brandCore} />
        </span>
        <span className={styles.brandWord}>Donna</span>
      </Link>

      <span className={styles.spacer} aria-hidden />

      <Link
        href="/dashboard/notifications"
        className={styles.action}
        aria-label={
          badgeCount > 0
            ? `Notifications, ${badgeCount} unread`
            : "Notifications"
        }
      >
        <Icon name="bell" className={styles.actionIcon} />
        {badgeCount > 0 ? (
          <span className={styles.badge} aria-hidden>
            {badgeCount > 99 ? "99+" : badgeCount}
          </span>
        ) : null}
      </Link>

      <Link
        href="/dashboard/settings"
        className={styles.avatar}
        aria-label={`Settings for ${profileName}`}
      >
        {showAvatar ? (
          <img
            src={profileAvatarUrl!}
            alt=""
            className={styles.avatarImage}
            referrerPolicy="no-referrer"
            onError={() => setAvatarFailed(true)}
          />
        ) : (
          profileInitials
        )}
      </Link>
    </header>
  );
}
