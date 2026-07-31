"use client";

import { format } from "date-fns";
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type DragEvent,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import { usePathname } from "next/navigation";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import { useAuth } from "@/features/auth";
import { DateMark } from "@/features/calendar/sections/DateMark";
import { MiniCalendar } from "@/features/calendar/sections/MiniCalendar";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useFlipList } from "./Tasks.flip";
import { useTaskJournal } from "./Tasks.logic";
import { journalStyles as styles } from "./Tasks.styles";
import type { TaskOccurrence } from "./Tasks.types";
import { TaskEditModal } from "./sections/TaskEditModal";
import { TaskTagPicker } from "./sections/TaskTagPicker";
import { TaskTagPill } from "./sections/TaskTagPill";
import { TaskTagsPanel } from "./sections/TaskTagsPanel";

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
  tags,
  onToggle,
  onDelete,
  onEdit,
  onTagsChange,
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
  tags: TaskOccurrence["tags"];
  onToggle: (occurrence: TaskOccurrence) => void;
  onDelete: (occurrence: TaskOccurrence) => void;
  onEdit: (occurrence: TaskOccurrence) => void;
  onTagsChange: (taskId: string, tagIds: string[]) => void;
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
  const selectedTagIds = (occurrence.tags ?? []).map((tag) => tag.id);
  const [menuOpen, setMenuOpen] = useState(false);
  const [tagOpen, setTagOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onPointerDown = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setMenuOpen(false);
        setTagOpen(false);
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, [menuOpen]);

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
        <div className={styles.itemRow}>
          <p
            className={cn(
              styles.itemTitle,
              occurrence.completed && styles.itemTitleDone,
            )}
          >
            {occurrence.title}
          </p>
          {occurrence.source === "carry_forward" ? (
            <span
              className={styles.carriedPill}
              title="Carried forward from a previous day"
            >
              Carried
            </span>
          ) : null}
          {(occurrence.tags ?? []).map((tag) => (
            <TaskTagPill key={tag.id} tag={tag} />
          ))}
        </div>
        {occurrence.project ? (
          <div className={styles.itemMeta}>
            <span>{occurrence.project}</span>
          </div>
        ) : null}
      </div>

      <div ref={menuRef} className={styles.menuRoot}>
        <button
          type="button"
          className={styles.menuTrigger}
          aria-label="Task actions"
          aria-expanded={menuOpen}
          disabled={deleting}
          onClick={() => {
            setMenuOpen((value) => !value);
            setTagOpen(false);
          }}
        >
          <Icon name="more" className="h-3.5 w-3.5" />
        </button>
        {menuOpen && !tagOpen ? (
          <div className={styles.menuPanel} role="menu">
            <button
              type="button"
              role="menuitem"
              className={styles.menuItem}
              onClick={() => {
                setMenuOpen(false);
                onEdit(occurrence);
              }}
            >
              <Icon name="compose" className="h-3.5 w-3.5" />
              Edit
            </button>
            <button
              type="button"
              role="menuitem"
              className={styles.menuItem}
              disabled={(tags ?? []).length === 0}
              onClick={() => setTagOpen(true)}
            >
              <Icon name="pin" className="h-3.5 w-3.5" />
              Tag
            </button>
            <button
              type="button"
              role="menuitem"
              className={cn(styles.menuItem, styles.menuItemDanger)}
              onClick={() => {
                setMenuOpen(false);
                onDelete(occurrence);
              }}
            >
              <Icon name="trash" className="h-3.5 w-3.5" />
              Delete
            </button>
          </div>
        ) : null}
        <TaskTagPicker
          tags={tags ?? []}
          selectedIds={selectedTagIds}
          disabled={deleting}
          hideTrigger
          open={tagOpen}
          onOpenChange={(next) => {
            setTagOpen(next);
            if (!next) setMenuOpen(false);
          }}
          onChange={(tagIds) => onTagsChange(occurrence.task_id, tagIds)}
        />
      </div>

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
  const [editing, setEditing] = useState<TaskOccurrence | null>(null);

  const occurrenceIds = useMemo(
    () => journal.occurrences.map((occurrence) => occurrence.id),
    [journal.occurrences],
  );
  const setFlipRef = useFlipList(occurrenceIds);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  const stats = journal.statistics;
  const isFiltered = journal.filterTagIds.length > 0;
  const showEmptyFiltered =
    !journal.isLoading &&
    !journal.isError &&
    journal.allOccurrences.length > 0 &&
    journal.occurrences.length === 0;

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

  const submitNewTask = (title: string) => {
    journal.addTask(title);
  };

  const onAddTaskSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const raw = new FormData(event.currentTarget).get("title");
    submitNewTask(typeof raw === "string" ? raw : "");
  };

  const onAddTaskKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== "Enter" || event.nativeEvent.isComposing) {
      return;
    }
    event.preventDefault();
    submitNewTask(event.currentTarget.value);
  };

  const tagsPanel = (
    <TaskTagsPanel
      tags={journal.tags}
      filterTagIds={journal.filterTagIds}
      isSaving={journal.isSaving}
      onToggleFilter={journal.toggleFilterTag}
      onClearFilters={journal.clearFilterTags}
      onCreateTag={journal.createTag}
      onDeleteTag={journal.removeTag}
    />
  );

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
            <header className={styles.header}>
              <DateMark
                date={journal.selectedDate}
                onClick={journal.goToday}
                className="min-w-0 flex-1 lg:hidden"
              />
              <p className="hidden min-w-0 flex-1 truncate text-sm font-medium text-donna-text lg:block">
                {journal.titleLabel}
              </p>
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

            <div className={styles.main}>
              <aside className={styles.sidebar}>
                <MiniCalendar
                  month={journal.miniMonth}
                  selected={journal.selectedDate}
                  onSelectDay={journal.selectDay}
                  onMonthShift={journal.shiftMiniMonth}
                  aria-label="Journal calendar"
                  dayExtra={(day) => {
                    const summary = journal.historyByDate.get(
                      format(day, "yyyy-MM-dd"),
                    );
                    if (!summary || summary.total <= 0) {
                      return null;
                    }
                    return `${summary.completed}/${summary.total}`;
                  }}
                />
                {tagsPanel}
              </aside>

              <div className={styles.tasksCol}>
                <section className={styles.tasksCard} aria-label="Tasks">
                  {isFiltered ? (
                    <div className={styles.filterRow}>
                      {journal.tags
                        .filter((tag) => journal.filterTagIds.includes(tag.id))
                        .map((tag) => (
                          <TaskTagPill
                            key={tag.id}
                            tag={tag}
                            selected
                            onClick={() => journal.toggleFilterTag(tag.id)}
                          />
                        ))}
                      <span className={styles.filterHint}>
                        Showing tasks with selected tags
                      </span>
                    </div>
                  ) : null}

                  <form className={styles.addRow} onSubmit={onAddTaskSubmit}>
                    <input
                      className={styles.addInput}
                      name="title"
                      placeholder="Add a task for this day…"
                      value={journal.draftTitle}
                      autoComplete="off"
                      enterKeyHint="done"
                      onKeyDown={onAddTaskKeyDown}
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

                  <div className={styles.tasksBody}>
                    {journal.isLoading ? (
                      <p className={styles.state}>Loading notebook…</p>
                    ) : null}
                    {!journal.isLoading && journal.isError ? (
                      <p className={styles.state}>
                        Couldn&apos;t load this day.
                      </p>
                    ) : null}
                    {!journal.isLoading &&
                    !journal.isError &&
                    journal.allOccurrences.length === 0 ? (
                      <p className={styles.empty}>
                        Nothing here yet. Add a task — it stays on this
                        day&apos;s page.
                      </p>
                    ) : null}
                    {showEmptyFiltered ? (
                      <p className={styles.empty}>
                        No tasks match the selected tags.
                      </p>
                    ) : null}
                    {!journal.isLoading && journal.occurrences.length > 0 ? (
                      <ul className={styles.list}>
                        {journal.occurrences.map((occurrence) => (
                          <JournalTaskRow
                            key={occurrence.id}
                            occurrence={occurrence}
                            tags={journal.tags}
                            onToggle={journal.toggleComplete}
                            onDelete={journal.removeTask}
                            onEdit={setEditing}
                            onTagsChange={journal.setTaskTags}
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
                  </div>
                </section>
              </div>

              <aside className={styles.statsCol}>
                <div className={styles.statsCard}>
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

                  {journal.tags.length > 0 ? (
                    <>
                      <h2 className={cn(styles.statsTitle, "mt-4")}>By Tag</h2>
                      <div className={styles.statsGrid}>
                        {journal.tags.map((tag) => {
                          const count = journal.allOccurrences.filter((o) =>
                            (o.tags ?? []).some((t) => t.id === tag.id),
                          ).length;
                          return (
                            <div key={tag.id} className={styles.statRow}>
                              <span
                                className={styles.statLabel}
                                style={{ color: tag.color }}
                              >
                                {tag.name}
                              </span>
                              <span className={styles.statValue}>{count}</span>
                            </div>
                          );
                        })}
                      </div>
                    </>
                  ) : null}
                </div>
              </aside>

              <div className={styles.mobileTags}>{tagsPanel}</div>
            </div>
          </div>
        </main>
      </div>

      <TaskEditModal
        open={Boolean(editing)}
        occurrence={editing}
        saving={journal.isSaving}
        onClose={() => setEditing(null)}
        onSave={async (input) => {
          if (!editing) return;
          await journal.editTask(editing, input);
        }}
      />
    </div>
  );
}
