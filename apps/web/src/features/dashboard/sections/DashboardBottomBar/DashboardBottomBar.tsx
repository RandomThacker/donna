"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { Icon } from "@/components/common";
import type { IconName } from "@/components/common";
import { cn } from "@/lib/cn";
import { useNotificationsCenter } from "@/features/notifications";

import { bottomNavItemsForPath } from "../../dashboardNav";
import { bottomBarStyles as styles } from "./DashboardBottomBar.styles";

export function DashboardBottomBar() {
  const pathname = usePathname();
  const router = useRouter();
  const items = bottomNavItemsForPath(pathname);
  const { badgeCount } = useNotificationsCenter();

  return (
    <div className={styles.root}>
      <nav className={styles.nav} aria-label="Primary">
        {items.map((item) => (
          <Link
            key={item.id}
            href={item.href}
            aria-current={item.active ? "page" : undefined}
            className={cn(styles.item, item.active && styles.itemActive)}
            onClick={(event) => {
              // Same-route taps (e.g. Home while already home) still need a
              // reliable response — scroll the workspace back to top.
              if (!item.active) return;
              event.preventDefault();
              if (item.id === "home") {
                router.push(item.href);
              }
              const scroller = document.querySelector(
                "main.overflow-y-auto, [data-dashboard-scroll]",
              );
              if (scroller instanceof HTMLElement) {
                scroller.scrollTo({ top: 0, behavior: "smooth" });
              } else {
                window.scrollTo({ top: 0, behavior: "smooth" });
              }
            }}
          >
            {item.active ? (
              <span className={styles.indicator} aria-hidden />
            ) : null}
            <span
              className={cn(
                styles.iconWrap,
                item.active && styles.iconWrapActive,
              )}
            >
              <Icon name={item.icon as IconName} className={styles.icon} />
              {item.id === "notifications" && badgeCount > 0 ? (
                <span className={styles.badge} aria-hidden>
                  {badgeCount > 99 ? "99+" : badgeCount}
                </span>
              ) : null}
            </span>
            <span
              className={cn(styles.label, item.active && styles.labelActive)}
            >
              {item.label}
            </span>
          </Link>
        ))}
      </nav>
    </div>
  );
}
