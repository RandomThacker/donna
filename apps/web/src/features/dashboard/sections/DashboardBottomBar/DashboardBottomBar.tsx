"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { bottomNavItemsForPath } from "../../dashboardNav";
import { bottomBarStyles as styles } from "./DashboardBottomBar.styles";

export function DashboardBottomBar() {
  const pathname = usePathname();
  const items = bottomNavItemsForPath(pathname);

  return (
    <div className={styles.root}>
      <nav className={styles.nav} aria-label="Primary">
        {items.map((item) => (
          <Link
            key={item.id}
            href={item.href}
            aria-current={item.active ? "page" : undefined}
            className={cn(styles.item, item.active && styles.itemActive)}
          >
            <span
              className={cn(
                styles.iconWrap,
                item.active && styles.iconWrapActive,
              )}
            >
              <Icon name={item.icon} className={styles.icon} />
            </span>
            <span className={styles.label}>{item.label}</span>
          </Link>
        ))}
      </nav>
    </div>
  );
}
