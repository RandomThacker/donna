"use client";

import { cn } from "@/lib/cn";

import type {
  DonnaNotification,
  NotificationFilter,
  NotificationGroup,
} from "../../Notifications.types";
import { FILTER_OPTIONS } from "../../Notifications.utils";
import { NotificationCard } from "../NotificationCard/NotificationCard";
import { listStyles as styles } from "./NotificationList.styles";

type Props = {
  filter: NotificationFilter;
  onFilterChange: (filter: NotificationFilter) => void;
  search: string;
  onSearchChange: (value: string) => void;
  groups: NotificationGroup[];
  filteredCount: number;
  totalCount: number;
  hasMore: boolean;
  onLoadMore: () => void;
  isLoading: boolean;
  isError: boolean;
  saving?: boolean;
  onOpen: (id: string) => void;
  onRead: (id: string) => void;
  onDismiss: (id: string) => void;
};

export function NotificationList({
  filter,
  onFilterChange,
  search,
  onSearchChange,
  groups,
  filteredCount,
  totalCount,
  hasMore,
  onLoadMore,
  isLoading,
  isError,
  saving,
  onOpen,
  onRead,
  onDismiss,
}: Props) {
  const emptyUnread = filter === "unread" && filteredCount === 0 && totalCount > 0;
  const emptyAll = totalCount === 0 && !isLoading && !isError;

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className={styles.toolbar}>
        <input
          className={styles.search}
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search notifications…"
          aria-label="Search notifications"
        />
        <div className={styles.filters} role="tablist" aria-label="Filter">
          {FILTER_OPTIONS.map((option) => (
            <button
              key={option.id}
              type="button"
              role="tab"
              aria-selected={filter === option.id}
              className={cn(
                styles.filterChip,
                filter === option.id && styles.filterChipOn,
              )}
              onClick={() => onFilterChange(option.id)}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto scrollbar-hidden">
        {isLoading ? (
          <p className={styles.state}>Loading notifications…</p>
        ) : null}
        {isError ? (
          <p className={styles.state}>Couldn&apos;t load notifications.</p>
        ) : null}
        {emptyAll ? (
          <div className={styles.empty}>
            <p className={styles.emptyTitle}>No notifications yet.</p>
            <p className={styles.emptyBody}>
              When Donna queues reminders and event pings, they&apos;ll land
              here.
            </p>
          </div>
        ) : null}
        {emptyUnread ? (
          <div className={styles.empty}>
            <p className={styles.emptyTitle}>You&apos;re all caught up.</p>
            <p className={styles.emptyBody}>No unread notifications.</p>
          </div>
        ) : null}
        {!isLoading &&
        !isError &&
        !emptyAll &&
        !emptyUnread &&
        filteredCount === 0 ? (
          <div className={styles.empty}>
            <p className={styles.emptyTitle}>Nothing matches.</p>
            <p className={styles.emptyBody}>
              Try another filter or clear the search.
            </p>
          </div>
        ) : null}

        {groups.length > 0 ? (
          <div className={styles.groups}>
            {groups.map((group) => (
              <section key={group.key} aria-label={group.label}>
                <h3 className={styles.groupLabel}>{group.label}</h3>
                <ul className="space-y-0.5">
                  {group.items.map((item: DonnaNotification) => (
                    <li key={item.id}>
                      <NotificationCard
                        notification={item}
                        saving={saving}
                        onOpen={onOpen}
                        onRead={onRead}
                        onDismiss={onDismiss}
                      />
                    </li>
                  ))}
                </ul>
              </section>
            ))}
          </div>
        ) : null}

        {hasMore ? (
          <div className={styles.loadMoreWrap}>
            <button
              type="button"
              className={styles.loadMore}
              onClick={onLoadMore}
            >
              Load more
            </button>
          </div>
        ) : null}
      </div>
    </div>
  );
}
