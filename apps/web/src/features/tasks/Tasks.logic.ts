import {
  addDays,
  addMonths,
  endOfMonth,
  format,
  isToday,
  startOfMonth,
  subMonths,
} from "date-fns";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";

import {
  createTask,
  deleteTask,
  fetchTaskDay,
  fetchTaskHistory,
  reorderTaskOccurrences,
  updateTask,
  updateTaskOccurrence,
} from "./Tasks.api";
import {
  createTaskTag,
  deleteTaskTag,
  fetchTaskTags,
  updateTaskTagAssignments,
} from "./Tasks.tags.api";
import type { TaskDayResponse, TaskOccurrence } from "./Tasks.types";

export const taskQueryKeys = {
  all: ["tasks"] as const,
  day: (date: string) => ["tasks", "day", date] as const,
  history: (from: string, to: string) => ["tasks", "history", from, to] as const,
  tags: ["tasks", "tags"] as const,
};

export function formatJournalDate(date: Date): string {
  return format(date, "yyyy-MM-dd");
}

export function useTaskJournal() {
  const queryClient = useQueryClient();
  const [selectedDate, setSelectedDate] = useState(() => new Date());
  const [miniMonth, setMiniMonth] = useState(() => startOfMonth(new Date()));
  const [draftTitle, setDraftTitle] = useState("");
  const [filterTagIds, setFilterTagIds] = useState<string[]>([]);

  const dateKey = formatJournalDate(selectedDate);

  const dayQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const tagsQuery = useQuery({
    queryKey: taskQueryKeys.tags,
    queryFn: ({ signal }) => fetchTaskTags(signal),
  });

  const historyFrom = formatJournalDate(startOfMonth(miniMonth));
  const historyTo = formatJournalDate(endOfMonth(miniMonth));

  const historyQuery = useQuery({
    queryKey: taskQueryKeys.history(historyFrom, historyTo),
    queryFn: ({ signal }) => fetchTaskHistory(historyFrom, historyTo, signal),
  });

  useEffect(() => {
    setMiniMonth(startOfMonth(selectedDate));
  }, [selectedDate]);

  const historyByDate = useMemo(() => {
    const map = new Map<
      string,
      { total: number; completed: number; pending: number; carried: number }
    >();
    for (const day of historyQuery.data?.days ?? []) {
      map.set(day.date, day);
    }
    return map;
  }, [historyQuery.data?.days]);

  const occurrences = dayQuery.data?.occurrences ?? [];
  const statistics = dayQuery.data?.statistics;
  const tags = tagsQuery.data ?? [];

  const filteredOccurrences = useMemo(() => {
    const list =
      filterTagIds.length === 0
        ? occurrences
        : occurrences.filter((occurrence) => {
            const taskTagIds = (occurrence.tags ?? []).map((tag) => tag.id);
            return filterTagIds.some((id) => taskTagIds.includes(id));
          });
    // Incomplete first, completed at the bottom (matches API ORDER BY).
    return [...list].sort((a, b) => {
      if (a.completed !== b.completed) {
        return a.completed ? 1 : -1;
      }
      return a.sort_order - b.sort_order;
    });
  }, [filterTagIds, occurrences]);

  const invalidateDay = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: taskQueryKeys.day(dateKey) });
  }, [queryClient, dateKey]);

  const invalidateHistory = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: taskQueryKeys.history(historyFrom, historyTo),
    });
  }, [queryClient, historyFrom, historyTo]);

  const invalidateTags = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: taskQueryKeys.tags });
  }, [queryClient]);

  const createMutation = useMutation({
    mutationFn: (title: string) => createTask({ title, date: dateKey }),
    onSuccess: async (created) => {
      setDraftTitle("");
      const previous = queryClient.getQueryData<TaskDayResponse>(
        taskQueryKeys.day(dateKey),
      );
      const existing = previous?.occurrences ?? [];
      const others = existing.filter((o) => o.id !== created.id);
      const incomplete = others.filter((o) => !o.completed);
      const completed = others.filter((o) => o.completed);
      const nextIds = [
        created.id,
        ...incomplete.map((o) => o.id),
        ...completed.map((o) => o.id),
      ];

      queryClient.setQueryData<TaskDayResponse>(taskQueryKeys.day(dateKey), {
        date: dateKey,
        note: previous?.note ?? { content: "" },
        statistics: previous?.statistics ?? {
          total: nextIds.length,
          completed: completed.length,
          pending: incomplete.length + 1,
          carried: previous?.statistics?.carried ?? 0,
          completion_pct: previous?.statistics?.completion_pct ?? 0,
          completed_today: previous?.statistics?.completed_today ?? 0,
          carried_forward: previous?.statistics?.carried_forward ?? 0,
          longest_carried_streak:
            previous?.statistics?.longest_carried_streak ?? 0,
          streak: previous?.statistics?.streak ?? 0,
        },
        occurrences: [
          { ...created, sort_order: 0, completed: false },
          ...incomplete.map((o, i) => ({ ...o, sort_order: i + 1 })),
          ...completed.map((o, i) => ({
            ...o,
            sort_order: incomplete.length + 1 + i,
          })),
        ],
      });

      try {
        await reorderTaskOccurrences({
          date: dateKey,
          occurrence_ids: nextIds,
        });
      } catch {
        // Soft-fail: list already shows the new task on top locally.
      }
      await invalidateDay();
      await invalidateHistory();
    },
  });

  const completeMutation = useMutation({
    mutationFn: ({ id, completed }: { id: string; completed: boolean }) =>
      updateTaskOccurrence(id, { completed }),
    onSuccess: async () => {
      await invalidateDay();
      await invalidateHistory();
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({
      taskId,
      occurrenceId,
      title,
      date,
      currentDate,
    }: {
      taskId: string;
      occurrenceId: string;
      title: string;
      date: string;
      currentDate: string;
    }) => {
      await updateTask(taskId, { title });
      if (date !== currentDate) {
        await updateTaskOccurrence(occurrenceId, { date });
      }
    },
    onSuccess: async (_data, variables) => {
      await invalidateDay();
      await invalidateHistory();
      if (variables.date !== variables.currentDate) {
        await queryClient.invalidateQueries({
          queryKey: taskQueryKeys.day(variables.date),
        });
      }
    },
  });

  const reorderMutation = useMutation({
    mutationFn: (ids: string[]) =>
      reorderTaskOccurrences({ date: dateKey, occurrence_ids: ids }),
    onMutate: async (ids) => {
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.day(dateKey) });
      const previous = queryClient.getQueryData<TaskDayResponse>(
        taskQueryKeys.day(dateKey),
      );
      if (previous) {
        const byId = new Map(
          previous.occurrences.map((occurrence) => [occurrence.id, occurrence]),
        );
        const nextOccurrences = ids
          .map((id) => byId.get(id))
          .filter((occurrence): occurrence is TaskOccurrence => Boolean(occurrence))
          .map((occurrence, index) => ({ ...occurrence, sort_order: index }));
        queryClient.setQueryData<TaskDayResponse>(taskQueryKeys.day(dateKey), {
          ...previous,
          occurrences: nextOccurrences,
        });
      }
      return { previous };
    },
    onError: (_error, _ids, context) => {
      if (context?.previous) {
        queryClient.setQueryData(taskQueryKeys.day(dateKey), context.previous);
      }
    },
    onSettled: async () => {
      await invalidateDay();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (taskId: string) => deleteTask(taskId),
    onSuccess: async () => {
      await invalidateDay();
      await invalidateHistory();
    },
  });

  const createTagMutation = useMutation({
    mutationFn: createTaskTag,
    onSuccess: async () => {
      await invalidateTags();
    },
  });

  const deleteTagMutation = useMutation({
    mutationFn: deleteTaskTag,
    onSuccess: async (_data, tagId) => {
      await invalidateTags();
      await invalidateDay();
      setFilterTagIds((current) => current.filter((id) => id !== tagId));
    },
  });

  const assignTagsMutation = useMutation({
    mutationFn: ({
      taskId,
      tagIds,
    }: {
      taskId: string;
      tagIds: string[];
    }) => updateTaskTagAssignments(taskId, tagIds),
    onSuccess: async () => {
      await invalidateDay();
    },
  });

  const goToday = useCallback(() => {
    setSelectedDate(new Date());
  }, []);

  const goPrevDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, -1));
  }, []);

  const goNextDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, 1));
  }, []);

  const selectDay = useCallback((day: Date) => {
    setSelectedDate(day);
  }, []);

  const shiftMiniMonth = useCallback((direction: -1 | 1) => {
    setMiniMonth((month) =>
      direction === 1 ? addMonths(month, 1) : subMonths(month, 1),
    );
  }, []);

  const addTask = useCallback(
    (titleOverride?: string) => {
      const title = (titleOverride ?? draftTitle).trim();
      if (!title || createMutation.isPending) {
        return;
      }
      createMutation.mutate(title);
    },
    [createMutation, draftTitle],
  );

  const toggleComplete = useCallback(
    (occurrence: TaskOccurrence) => {
      completeMutation.mutate({
        id: occurrence.id,
        completed: !occurrence.completed,
      });
    },
    [completeMutation],
  );

  const editTask = useCallback(
    async (occurrence: TaskOccurrence, input: { title: string; date: string }) => {
      await updateMutation.mutateAsync({
        taskId: occurrence.task_id,
        occurrenceId: occurrence.id,
        title: input.title,
        date: input.date,
        currentDate: occurrence.date,
      });
      if (input.date !== occurrence.date) {
        setSelectedDate(new Date(`${input.date}T12:00:00`));
      }
    },
    [updateMutation],
  );

  const reorder = useCallback(
    (fromId: string, toId: string) => {
      if (fromId === toId) {
        return;
      }
      const ids = filteredOccurrences.map((o) => o.id);
      const fromIndex = ids.indexOf(fromId);
      const toIndex = ids.indexOf(toId);
      if (fromIndex < 0 || toIndex < 0) {
        return;
      }
      const next = [...ids];
      const [moved] = next.splice(fromIndex, 1);
      next.splice(toIndex, 0, moved!);
      reorderMutation.mutate(next);
    },
    [filteredOccurrences, reorderMutation],
  );

  const removeTask = useCallback(
    (occurrence: TaskOccurrence) => {
      deleteMutation.mutate(occurrence.task_id);
    },
    [deleteMutation],
  );

  const toggleFilterTag = useCallback((tagId: string) => {
    setFilterTagIds((current) =>
      current.includes(tagId)
        ? current.filter((id) => id !== tagId)
        : [...current, tagId],
    );
  }, []);

  const clearFilterTags = useCallback(() => {
    setFilterTagIds([]);
  }, []);

  const createTag = useCallback(
    (input: { name: string; color: string }) => {
      createTagMutation.mutate(input);
    },
    [createTagMutation],
  );

  const removeTag = useCallback(
    (tagId: string) => {
      deleteTagMutation.mutate(tagId);
    },
    [deleteTagMutation],
  );

  const setTaskTags = useCallback(
    (taskId: string, tagIds: string[]) => {
      assignTagsMutation.mutate({ taskId, tagIds });
    },
    [assignTagsMutation],
  );

  const titleLabel = isToday(selectedDate)
    ? `Today · ${format(selectedDate, "MMMM d, yyyy")}`
    : format(selectedDate, "EEEE, MMMM d, yyyy");

  return {
    selectedDate,
    dateKey,
    titleLabel,
    miniMonth,
    historyByDate,
    occurrences: filteredOccurrences,
    allOccurrences: occurrences,
    statistics,
    tags,
    filterTagIds,
    draftTitle,
    setDraftTitle,
    isLoading: dayQuery.isLoading,
    isError: dayQuery.isError,
    isSaving:
      createMutation.isPending ||
      completeMutation.isPending ||
      updateMutation.isPending ||
      reorderMutation.isPending ||
      deleteMutation.isPending ||
      createTagMutation.isPending ||
      deleteTagMutation.isPending ||
      assignTagsMutation.isPending,
    goToday,
    goPrevDay,
    goNextDay,
    selectDay,
    shiftMiniMonth,
    addTask,
    toggleComplete,
    editTask,
    reorder,
    removeTask,
    toggleFilterTag,
    clearFilterTags,
    createTag,
    removeTag,
    setTaskTags,
  };
}

export function sourceGlyph(source: TaskOccurrence["source"]): string | null {
  switch (source) {
    case "carry_forward":
      return "↻";
    case "recurring":
      return "🔁";
    case "ai":
      return "🤖";
    case "calendar":
      return "📅";
    default:
      return null;
  }
}
