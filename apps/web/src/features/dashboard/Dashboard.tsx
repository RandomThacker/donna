"use client";

import { usePathname } from "next/navigation";

import { useAuth } from "@/features/auth";

import { getDashboardContent } from "./Dashboard.logic";
import { navItemsForPath } from "./dashboardNav";
import { dashboardStyles as styles } from "./Dashboard.styles";
import { DashboardFocus } from "./sections/DashboardFocus";
import { DashboardGreeting } from "./sections/DashboardGreeting";
import { DashboardInsights } from "./sections/DashboardInsights";
import { DashboardPhone } from "./sections/DashboardPhone";
import { DashboardQuickTasks } from "./sections/DashboardQuickTasks";
import { DashboardSidebar } from "./sections/DashboardSidebar";
import { DashboardTimeline } from "./sections/DashboardTimeline";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "D";
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase();
  }
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

export function Dashboard() {
  const pathname = usePathname();
  const { user } = useAuth();
  const { data } = getDashboardContent();
  const nav = navItemsForPath(pathname);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || data.profileName;
  const profileInitials = initialsFrom(profileName);
  const greetingName = profileName.split(/\s+/)[0] || data.greeting.name;

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={nav}
          profileName={profileName}
          profileInitials={profileInitials}
          profileEmail={user?.email}
          profileAvatarUrl={user?.avatar_url}
        />
        <main className={styles.workspace}>
          <div className={styles.workspaceInner}>
            <div className={styles.bento}>
              <DashboardGreeting
                greeting={{ ...data.greeting, name: greetingName }}
              />
              <DashboardFocus focus={data.focus} />
              <DashboardInsights insights={data.insights} />
              <DashboardTimeline items={data.timeline} />
              <DashboardQuickTasks tasks={data.tasks} />
            </div>
            <div className={styles.phoneMobile}>
              <DashboardPhone phone={data.phone} />
            </div>
          </div>
        </main>
        <aside className={styles.phoneColumn} aria-label="Donna phone">
          <DashboardPhone phone={data.phone} />
        </aside>
      </div>
    </div>
  );
}
