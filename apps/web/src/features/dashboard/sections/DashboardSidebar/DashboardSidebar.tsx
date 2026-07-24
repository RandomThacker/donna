import { Icon } from "@/components/common";
import { ThemeToggle } from "@/components/theme";
import { cn } from "@/lib/cn";

import { sidebarStyles as styles } from "./DashboardSidebar.styles";
import type { DashboardSidebarProps } from "./DashboardSidebar.types";

export function DashboardSidebar({
  items,
  profileName,
  profileInitials,
}: DashboardSidebarProps) {
  return (
    <aside className={styles.aside} aria-label="Primary">
      <a href="/dashboard" className={styles.brand} aria-label="Donna home">
        <span className={styles.brandMark} aria-hidden>
          <span className={styles.brandCore} />
        </span>
        <span className={styles.brandWord}>Donna</span>
      </a>
      <nav className={styles.nav}>
        {items.map((item) => (
          <a
            key={item.id}
            href={item.href}
            aria-current={item.active ? "page" : undefined}
            className={cn(styles.item, item.active && styles.itemActive)}
          >
            <Icon name={item.icon} className={cn("h-4 w-4", styles.itemIcon)} />
            {item.label}
          </a>
        ))}
      </nav>
      <div className={styles.footer}>
        <ThemeToggle className={styles.themeToggle} />
        <div className={styles.profile}>
          <span className={styles.avatar}>{profileInitials}</span>
          <div className={styles.profileMeta}>
            <p className={styles.profileName}>{profileName}</p>
            <p className={styles.profileHint}>Personal workspace</p>
          </div>
        </div>
      </div>
    </aside>
  );
}
