"use client";

import { usePathname } from "next/navigation";

import { Icon } from "@/components/common";
import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useNotificationsCenter } from "./Notifications.logic";
import { notificationsPageStyles as styles } from "./Notifications.styles";
import { NotificationDetails } from "./sections/NotificationDetails/NotificationDetails";
import { NotificationList } from "./sections/NotificationList/NotificationList";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "D";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

export function Notifications() {
  const pathname = usePathname();
  const { user } = useAuth();
  const nav = navItemsForPath(pathname);
  const controller = useNotificationsCenter();

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
          <header className={styles.header}>
            <div className="min-w-0">
              <h1 className={styles.title}>Notifications</h1>
              <p className={styles.subtitle}>
                History, delivery status, and debugging for every ping Donna
                sends.
              </p>
            </div>
            {controller.badgeCount > 0 ? (
              <span className={styles.badge} aria-label="Unread count">
                {controller.badgeCount > 99 ? "99+" : controller.badgeCount}
              </span>
            ) : null}
          </header>

          <div className={styles.main}>
            <section className={styles.listPane} aria-label="Notification list">
              <NotificationList
                filter={controller.filter}
                onFilterChange={controller.setFilter}
                search={controller.search}
                onSearchChange={controller.setSearch}
                groups={controller.groups}
                filteredCount={controller.filtered.length}
                totalCount={controller.items.length}
                hasMore={controller.hasMore}
                onLoadMore={controller.loadMore}
                isLoading={controller.isLoading}
                isError={controller.isError}
                saving={controller.isSaving}
                onOpen={controller.openDetails}
                onRead={(id) => void controller.markRead(id)}
                onDismiss={(id) => void controller.dismiss(id)}
              />
            </section>

            <aside className={styles.detailsPane} aria-label="Notification details">
              {controller.selected ? (
                <NotificationDetails
                  notification={controller.selected}
                  saving={controller.isSaving}
                  onRead={controller.markRead}
                  onDismiss={controller.dismiss}
                  onCloseCenter={controller.backToList}
                />
              ) : (
                <div className={styles.detailsEmpty}>
                  Select a notification to inspect its timeline.
                </div>
              )}
            </aside>
          </div>
        </main>
      </div>

      {controller.selected ? (
        <div className={styles.mobileDetails} role="dialog" aria-modal="true">
          <header className={styles.mobileDetailsHeader}>
            <button
              type="button"
              className={styles.iconBtn}
              aria-label="Back to list"
              onClick={controller.backToList}
            >
              <Icon name="chevronLeft" className="h-4 w-4" />
            </button>
            <h2 className={styles.mobileDetailsTitle}>Details</h2>
          </header>
          <div className={styles.mobileDetailsBody}>
            <NotificationDetails
              notification={controller.selected}
              saving={controller.isSaving}
              onRead={controller.markRead}
              onDismiss={controller.dismiss}
              onCloseCenter={controller.backToList}
            />
          </div>
        </div>
      ) : null}
    </div>
  );
}
