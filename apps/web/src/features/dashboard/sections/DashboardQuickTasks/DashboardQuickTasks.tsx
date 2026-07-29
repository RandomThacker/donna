"use client";

import { format } from "date-fns";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import {
  createTask,
  fetchTaskDay,
  updateTaskOccurrence,
} from "@/features/tasks/Tasks.api";
import { taskQueryKeys } from "@/features/tasks/Tasks.logic";

import { BentoBox, bentoBoxStyles } from "../BentoBox";
import { quickTasksStyles as styles } from "./DashboardQuickTasks.styles";

export function DashboardQuickTasks() {
  const queryClient = useQueryClient();
  const dateKey = format(new Date(), "yyyy-MM-dd");
  const [draftTitle, setDraftTitle] = useState("");

  const tasksQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: taskQueryKeys.day(dateKey) });
    await queryClient.invalidateQueries({ queryKey: ["home"] });
  };

  const createMutation = useMutation({
    mutationFn: (title: string) => createTask({ title, date: dateKey }),
    onSuccess: async () => {
      setDraftTitle("");
      await invalidate();
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, completed }: { id: string; completed: boolean }) =>
      updateTaskOccurrence(id, completed),
    onSuccess: async () => {
      await invalidate();
    },
  });

  const tasks = tasksQuery.data?.occurrences ?? [];
  const isSaving = createMutation.isPending;
  const canAdd = draftTitle.trim().length > 0 && !isSaving;

  const handleAdd = () => {
    const title = draftTitle.trim();
    if (!title || isSaving) {
      return;
    }
    createMutation.mutate(title);
  };

  return (
    <BentoBox
      className={cn(styles.box, bentoBoxStyles.fixedPanel)}
      title="Quick tasks"
    >
      <form
        className={styles.addRow}
        onSubmit={(event) => {
          event.preventDefault();
          handleAdd();
        }}
      >
        <input
          className={styles.addInput}
          placeholder="Add a task…"
          value={draftTitle}
          disabled={isSaving}
          onChange={(event) => setDraftTitle(event.target.value)}
          aria-label="New task"
        />
        <button type="submit" className={styles.addBtn} disabled={!canAdd}>
          Add
        </button>
      </form>

      <div className={bentoBoxStyles.scrollBody}>
        {tasksQuery.isLoading ? (
          <p className={styles.state}>Loading today&apos;s tasks…</p>
        ) : null}
        {tasksQuery.isError ? (
          <p className={styles.state}>Couldn&apos;t load tasks.</p>
        ) : null}
        {!tasksQuery.isLoading && !tasksQuery.isError && tasks.length === 0 ? (
          <p className={styles.empty}>Nothing on today&apos;s list yet.</p>
        ) : null}
        {!tasksQuery.isLoading && !tasksQuery.isError && tasks.length > 0 ? (
          <ul className={styles.list}>
            {tasks.map((task) => (
              <li key={task.id}>
                <button
                  type="button"
                  className={styles.item}
                  aria-pressed={task.completed}
                  disabled={toggleMutation.isPending}
                  onClick={() =>
                    toggleMutation.mutate({
                      id: task.id,
                      completed: !task.completed,
                    })
                  }
                >
                  <span
                    className={cn(styles.check, task.completed && styles.checkDone)}
                  >
                    <Icon name="check" className="h-2.5 w-2.5" />
                  </span>
                  <span
                    className={cn(
                      styles.labelText,
                      task.completed && styles.labelDone,
                    )}
                  >
                    {task.title}
                    {task.carried_forward || task.source === "carry_forward" ? (
                      <span className={styles.carried}>Carried</span>
                    ) : null}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </BentoBox>
  );
}
