"use client";

import { usePathname } from "next/navigation";

import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { Integrations } from "./Integrations";
import { settingsStyles as styles } from "./Settings.styles";

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

export function IntegrationsPage() {
  const pathname = usePathname();
  const { user } = useAuth();
  const nav = navItemsForPath(pathname);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

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
          <div className={styles.workspaceInnerWide}>
            <Integrations />
          </div>
        </main>
      </div>
    </div>
  );
}
