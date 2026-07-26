"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

import { Icon } from "@/components/common";
import { ThemeToggle } from "@/components/theme";
import { cn } from "@/lib/cn";

import { sidebarStyles as styles } from "./DashboardSidebar.styles";
import type { DashboardSidebarProps } from "./DashboardSidebar.types";

export function DashboardSidebar({
  items,
  profileName,
  profileInitials,
  profileAvatarUrl,
  onSignOut,
}: DashboardSidebarProps) {
  const [avatarFailed, setAvatarFailed] = useState(false);

  useEffect(() => {
    setAvatarFailed(false);
  }, [profileAvatarUrl]);

  const showAvatar = Boolean(profileAvatarUrl) && !avatarFailed;

  return (
    <aside className={styles.aside} aria-label="Primary">
      <Link href="/dashboard" className={styles.brand} aria-label="Donna home">
        <span className={styles.brandMark} aria-hidden>
          <span className={styles.brandCore} />
        </span>
        <span className={styles.brandWord}>Donna</span>
      </Link>
      <nav className={styles.nav}>
        {items.map((item) => (
          <Link
            key={item.id}
            href={item.href}
            aria-current={item.active ? "page" : undefined}
            className={cn(styles.item, item.active && styles.itemActive)}
          >
            <Icon name={item.icon} className={cn("h-4 w-4", styles.itemIcon)} />
            {item.label}
          </Link>
        ))}
      </nav>
      <div className={styles.footer}>
        <ThemeToggle className={styles.themeToggle} />
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
            <p className={styles.profileHint}>Personal workspace</p>
          </div>
        </div>
        {onSignOut ? (
          <button type="button" className={styles.signOut} onClick={onSignOut}>
            Sign out
          </button>
        ) : null}
      </div>
    </aside>
  );
}
