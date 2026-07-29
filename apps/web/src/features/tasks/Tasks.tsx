"use client";

import { useMemo, useState, type DragEvent } from "react";
import { usePathname } from "next/navigation";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import { useAuth } from "@/features/auth";
import { DateMark } from "@/features/calendar/sections/DateMark";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useFlipList } from "./Tasks.flip";
import { useTaskJournal } from "./Tasks.logic";
import { journalStyles as styles } from "./Tasks.styles";
import type { TaskOccurrence } from "./Tasks.types";

const DRAG_TYPE = "text/task-occurrence-id";

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

function JournalTaskRow({
  occurrence,
  onToggle,
  onDelete,
  deleting,
  draggingId,
  overId,
  onDragStartRow,
  onDragOverRow,
  onDropRow,
  onDragEndRow,
  setNodeRef,
}: {
  occurrence: TaskOccurrence;
  onToggle: (occurrence: TaskOccurrence) => void;
  onDelete: (occurrence: TaskOccurrence) => void;
  deleting: boolean;
  draggingId: string | null;
  overId: string | null;
  onDragStartRow: (id: string, event: DragEvent) => void;
  onDragOverRow: (id: string, event: DragEvent) => void;
  onDropRow: (id: string, event: DragEvent) => void;
  onDragEndRow: () => void;
  setNodeRef: (el: HTMLElement | null) => void;
}) {
  const isDragging = draggingId === occurrence.id;
  const isOver = overId === occurrence.id && draggingId !== occurrence.id;

  return (
    <li
      ref={setNodeRef}
      onDragOver={(event) => onDragOverRow(occurrence.id, event)}
      onDrop={(event) => onDropRow(occurrence.id, event)}
      className={cn(
        styles.item,
        isDragging && styles.itemDragging,
        isOver && styles.itemDropTarget,
      )}
    >
      <button
        type="button"
        className={cn(styles.checkbox, occurrence.completed && styles.checkboxOn)}
        aria-label={occurrence.completed ? "Mark incomplete" : "Mark complete"}
        onClick={() => onToggle(occurrence)}
      >
        {occurrence.completed ? (
          <Icon name="check" className="h-3 w-3" />
        ) : null}
      </button>
      <div className={styles.itemBody}>
        <p
          className={cn(
            styles.itemTitle,
            occurrence.completed && styles.itemTitleDone,
          )}
        >
          <span>{occurrence.title}</span>
          {occurrence.source === "carry_forward" ? (
            <span
              className={styles.carriedPill}
              title="Carried forward from a previous day"
            >
              Carried
            </span>
          ) : null}
        </p>
        {occurrence.project ? (
          <div className={styles.itemMeta}>
            <span>{occurrence.project}</span>
          </div>
        ) : null}
      </div>
      <button
        type="button"
        className={styles.deleteBtn}
        aria-label="Delete task"
        disabled={deleting}
        onClick={(event) => {
          event.stopPropagation();
          onDelete(occurrence);
        }}
      >
        <Icon name="trash" className="h-3.5 w-3.5" />
      </button>
      <span
        role="button"
        tabIndex={0}
        className={styles.dragHandle}
        aria-label="Drag to reorder"
        draggable
        onDragStart={(event) => onDragStartRow(occurrence.id, event)}
        onDragEnd={onDragEndRow}
      >
        ⋮⋮
      </span>
    </li>
  );
}

export function Tasks() {
  const pathname = usePathname();
  const { user } = useAuth();
  const journal = useTaskJournal();
  const nav = navItemsForPath(pathname);
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [overId, setOverId] = useState<string | null>(null);

  const occurrenceIds = useMemo(
    () => journal.occurrences.map((occurrence) => occurrence.id),
    [journal.occurrences],
  );
  const setFlipRef = useFlipList(occurrenceIds);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  const stats = journal.statistics;

  const onDragStartRow = (id: string, event: DragEvent) => {
    event.dataTransfer.setData(DRAG_TYPE, id);
    event.dataTransfer.setData("text/plain", id);
    event.dataTransfer.effectAllowed = "move";
    setDraggingId(id);
  };

  const onDragOverRow = (id: string, event: DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    if (overId !== id) {
      setOverId(id);
    }
  };

  const onDropRow = (id: string, event: DragEvent) => {
    event.preventDefault();
    const fromId =
      event.dataTransfer.getData(DRAG_TYPE) ||
      event.dataTransfer.getData("text/plain") ||
      draggingId;
    if (fromId) {
      journal.reorder(fromId, id);
    }
    setDraggingId(null);
    setOverId(null);
  };

  const onDragEndRow = () => {
    setDraggingId(null);
    setOverId(null);
  };

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
            <div className={styles.main}>
              <header className={styles.header}>
                <DateMark
                  date={journal.selectedDate}
                  onClick={journal.goToday}
                  className="min-w-0 flex-1"
                />
                <div className={styles.nav}>
                  <button
                    type="button"
                    className={styles.navBtn}
                    onClick={journal.goPrevDay}
                    aria-label="Previous day"
                  >
                    <Icon name="chevronLeft" className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    className={styles.todayBtn}
                    onClick={journal.goToday}
                  >
                    Today
                  </button>
                  <button
                    type="button"
                    className={styles.navBtn}
                    onClick={journal.goNextDay}
                    aria-label="Next day"
                  >
                    <Icon name="chevronRight" className="h-4 w-4" />
                  </button>
                </div>
              </header>

              <div className="flex min-w-0 flex-col gap-6">
                  <section className={styles.tasksCard} aria-label="Tasks">
                    <form
                      className={styles.addRow}
                      onSubmit={(event) => {
                        event.preventDefault();
                        journal.addTask();
                      }}
                    >
                      <input
                        className={styles.addInput}
                        placeholder="Add a task for this day…"
                        value={journal.draftTitle}
                        onChange={(event) =>
                          journal.setDraftTitle(event.target.value)
                        }
                      />
                      <button
                        type="submit"
                        className={styles.addBtn}
                        disabled={journal.isSaving}
                      >
                        Add
                      </button>
                    </form>

                    {journal.isLoading ? (
                      <p className={styles.state}>Loading notebook…</p>
                    ) : null}
                    {!journal.isLoading && journal.isError ? (
                      <p className={styles.state}>Couldn&apos;t load this day.</p>
                    ) : null}
                    {!journal.isLoading &&
                    !journal.isError &&
                    journal.occurrences.length === 0 ? (
                      <p className={styles.empty}>
                        Nothing here yet. Add a task — it stays on this day&apos;s
                        page.
                      </p>
                    ) : null}
                    {!journal.isLoading && journal.occurrences.length > 0 ? (
                      <ul className={styles.list}>
                        {journal.occurrences.map((occurrence) => (
                          <JournalTaskRow
                            key={occurrence.id}
                            occurrence={occurrence}
                            onToggle={journal.toggleComplete}
                            onDelete={journal.removeTask}
                            deleting={journal.isSaving}
                            draggingId={draggingId}
                            overId={overId}
                            onDragStartRow={onDragStartRow}
                            onDragOverRow={onDragOverRow}
                            onDropRow={onDropRow}
                            onDragEndRow={onDragEndRow}
                            setNodeRef={setFlipRef(occurrence.id)}
                          />
                        ))}
                      </ul>
                    ) : null}
                  </section>
              </div>
            </div>

            <aside className={styles.statsCard} aria-label="Statistics">
              <h2 className={styles.statsTitle}>Statistics</h2>
              {stats ? (
                <div className={styles.statsGrid}>
                  <div className={styles.statRow}>
                    <span className={styles.statLabel}>Completion</span>
                    <span className={styles.statValue}>
                      {Math.round(stats.completion_pct)}%
                    </span>
                  </div>
                  <div className={styles.statRow}>
                    <span className={styles.statLabel}>Completed</span>
                    <span className={styles.statValue}>{stats.completed}</span>
                  </div>
                  <div className={styles.statRow}>
                    <span className={styles.statLabel}>Pending</span>
                    <span className={styles.statValue}>{stats.pending}</span>
                  </div>
                  <div className={styles.statRow}>
                    <span className={styles.statLabel}>Carried forward</span>
                    <span className={styles.statValue}>
                      {stats.carried_forward}
                    </span>
                  </div>
                  <div className={styles.statRow}>
                    <span className={styles.statLabel}>Streak</span>
                    <span className={styles.statValue}>{stats.streak}</span>
                  </div>
                  {stats.average_completion_min != null ? (
                    <div className={styles.statRow}>
                      <span className={styles.statLabel}>Avg completion</span>
                      <span className={styles.statValue}>
                        {Math.round(stats.average_completion_min)}m
                      </span>
                    </div>
                  ) : null}
                </div>
              ) : (
                <p className={styles.state}>—</p>
              )}
            </aside>
          </div>
        </main>
      </div>
    </div>
  );
}
