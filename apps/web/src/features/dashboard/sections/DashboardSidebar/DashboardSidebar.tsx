"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import { useNotificationsCenter } from "@/features/notifications";

import { sidebarStyles as styles } from "./DashboardSidebar.styles";
import type { DashboardSidebarProps } from "./DashboardSidebar.types";

export function DashboardSidebar({
  items,
  profileName,
  profileInitials,
  profileEmail,
  profileAvatarUrl,
  badgeByNavId,
}: DashboardSidebarProps) {
  const [avatarFailed, setAvatarFailed] = useState(false);
  const { badgeCount } = useNotificationsCenter();

  useEffect(() => {
    setAvatarFailed(false);
  }, [profileAvatarUrl]);

  const showAvatar = Boolean(profileAvatarUrl) && !avatarFailed;
  const email = profileEmail?.trim() || null;

  return (
    <aside className={styles.aside} aria-label="Primary">
      <Link href="/dashboard" className={styles.brand} aria-label="Donna home">
        <span className={styles.brandMark} aria-hidden>
          <span className={styles.brandCore} />
        </span>
        <span className={styles.brandWord}>Donna</span>
      </Link>
      <nav className={styles.nav}>
        {items.map((item) => {
          const badge =
            badgeByNavId?.[item.id] ??
            (item.id === "notifications" && badgeCount > 0
              ? badgeCount
              : undefined);
          return (
            <Link
              key={item.id}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={cn(styles.item, item.active && styles.itemActive)}
            >
              <Icon
                name={item.icon}
                className={cn("h-4 w-4", styles.itemIcon)}
              />
              <span className="min-w-0 flex-1 truncate">{item.label}</span>
              {badge != null && badge > 0 ? (
                <span className={styles.navBadge} aria-label={`${badge} unread`}>
                  {badge > 99 ? "99+" : badge}
                </span>
              ) : null}
            </Link>
          );
        })}
      </nav>
      <div className={styles.footer}>
        <div className={styles.profile}>
          <span className={styles.avatar}>
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
          </span>
          <div className={styles.profileMeta}>
            <p className={styles.profileName}>{profileName}</p>
            {email ? <p className={styles.profileHint}>{email}</p> : null}
          </div>
        </div>
      </div>
    </aside>
  );
}
