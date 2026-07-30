"use client";

import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";

import { Icon } from "@/components/common";

import type { NotificationsCenterController } from "../../Notifications.logic";
import { NotificationDetails } from "../NotificationDetails/NotificationDetails";
import { NotificationList } from "../NotificationList/NotificationList";
import { drawerStyles as styles } from "./NotificationDrawer.styles";

type Props = {
  controller: NotificationsCenterController;
};

export function NotificationDrawer({ controller }: Props) {
  const titleId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const {
    open,
    closeDrawer,
    selected,
    backToList,
    filter,
    setFilter,
    search,
    setSearch,
    groups,
    filtered,
    items,
    hasMore,
    loadMore,
    isLoading,
    isError,
    isSaving,
    openDetails,
    markRead,
    dismiss,
  } = controller;

  useEffect(() => {
    if (!open) return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (selected) {
          backToList();
        } else {
          closeDrawer();
        }
      }
    };
    window.addEventListener("keydown", onKey);
    panelRef.current?.focus();
    return () => {
      document.body.style.overflow = previous;
      window.removeEventListener("keydown", onKey);
    };
  }, [open, selected, closeDrawer, backToList]);

  if (!open || typeof document === "undefined") {
    return null;
  }

  const showingDetails = Boolean(selected);

  return createPortal(
    <div className={styles.root} role="presentation">
      <button
        type="button"
        className={styles.backdrop}
        aria-label="Close notifications"
        onClick={closeDrawer}
      />
      <aside
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className={styles.panel}
      >
        <header className={styles.header}>
          {showingDetails ? (
            <button
              type="button"
              className={styles.iconBtn}
              aria-label="Back to notifications"
              onClick={backToList}
            >
              <Icon name="chevronLeft" className="h-4 w-4" />
            </button>
          ) : null}
          <h2 id={titleId} className={styles.headerTitle}>
            {showingDetails ? "Details" : "Notifications"}
          </h2>
          <button
            type="button"
            className={styles.iconBtn}
            aria-label="Close"
            onClick={closeDrawer}
          >
            <Icon name="close" className="h-4 w-4" />
          </button>
        </header>
        <div className={styles.body}>
          {showingDetails && selected ? (
            <NotificationDetails
              notification={selected}
              saving={isSaving}
              onRead={markRead}
              onDismiss={dismiss}
              onCloseCenter={closeDrawer}
            />
          ) : (
            <NotificationList
              filter={filter}
              onFilterChange={setFilter}
              search={search}
              onSearchChange={setSearch}
              groups={groups}
              filteredCount={filtered.length}
              totalCount={items.length}
              hasMore={hasMore}
              onLoadMore={loadMore}
              isLoading={isLoading}
              isError={isError}
              saving={isSaving}
              onOpen={openDetails}
              onRead={(id) => void markRead(id)}
              onDismiss={(id) => void dismiss(id)}
            />
          )}
        </div>
      </aside>
    </div>,
    document.body,
  );
}
